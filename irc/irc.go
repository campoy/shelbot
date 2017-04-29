package irc

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/textproto"
	"strings"
	"sync"
	"time"
)

type Client struct {
	conn   io.ReadWriteCloser
	logger *log.Logger
	pause  time.Duration

	done chan struct{}
	once sync.Once

	messages        chan *Message
	privateMessages chan *PrivateMessage
}

func New(conn io.ReadWriteCloser, options ...Option) *Client {
	return &Client{
		conn:            conn,
		logger:          log.New(ioutil.Discard, "", 0),
		pause:           time.Second,
		done:            make(chan struct{}),
		messages:        make(chan *Message),
		privateMessages: make(chan *PrivateMessage),
	}
}

type Option func(*Client)

func WithLogger(logger *log.Logger) Option { return func(c *Client) { c.logger = logger } }
func WithPause(d time.Duration) Option     { return func(c *Client) { c.pause = d } }

func (c *Client) Messages() <-chan *Message               { return c.messages }
func (c *Client) PrivateMessages() <-chan *PrivateMessage { return c.privateMessages }

func (c *Client) Connect(nick, realName string) error {
	if err := c.send("USER %s 8 * :%s", nick, nick); err != nil {
		return err
	}
	return c.send("NICK %s", nick)
}

func (c *Client) Join(channel, key string) error          { return c.send("JOIN %s %s", channel, key) }
func (c *Client) JoinExclusive(channel, key string) error { return c.send("JOIN %s %s 0", channel, key) }

func (c *Client) Part(channel, msg string) error {
	if msg != "" {
		msg = ":" + msg
	}
	return c.send("PART %s %s", channel, msg)
}

func (c *Client) Send(target string, text string) error {
	prefix := fmt.Sprintf("PRIVMSG %s :", target)
	for {
		if len(prefix)+len(text) <= 400 {
			return c.send(prefix + text)
		}

		i := strings.LastIndex(text[:400-len(prefix)], " ")
		if err := c.send(prefix + text[:i]); err != nil {
			return err
		}
		text = text[i+1:]
	}
}

func (c *Client) Quit(msg string) error {
	if msg != "" {
		msg = ":" + msg
	}
	if err := c.send("QUIT %s", msg); err != nil {
		return err
	}
	c.once.Do(func() { close(c.done) })
	return nil
}

func (c *Client) send(format string, args ...interface{}) error {
	_, err := c.conn.Write([]byte(fmt.Sprintf(format+"\r\n", args...)))
	time.Sleep(c.pause)
	return err
}

func (c *Client) Listen() error {
	reader := textproto.NewReader(bufio.NewReader(c.conn))

	c.logger.Println("Ready to Listen")
	for {
		select {
		case <-c.done:
			close(c.privateMessages)
			close(c.messages)
			c.logger.Println("Listen exiting")
			return nil
		default:
			line, err := reader.ReadLine()
			if err != nil {
				return fmt.Errorf("could not read line: %v", err)
			}
			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, "PING") {
				c.send("PONG %s", strings.TrimPrefix(line, "PING"))
				continue
			}

			m, err := parseMessage(line)
			if err != nil {
				c.logger.Println("Error parsing raw message:", err)
			}
			switch m.Command {
			case "PRIVMSG":
				c.privateMessages <- privMsgFromMessage(m)
			default:
				select {
				case c.messages <- m:
				default:
				}
			}
		}
	}
}
