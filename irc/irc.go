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

	wg    sync.WaitGroup
	close chan struct{}

	messages     chan *Message
	privMessages chan *PrivMsg
}

func New(conn io.ReadWriteCloser, options ...Option) *Client {
	return &Client{
		conn:         conn,
		logger:       log.New(ioutil.Discard, "", 0),
		close:        make(chan struct{}),
		messages:     make(chan *Message),
		privMessages: make(chan *PrivMsg),
	}
}

type Option func(*Client)

func WithLogger(logger *log.Logger) Option {
	return func(c *Client) { c.logger = logger }
}

func (c *Client) Messages() <-chan *Message     { return c.messages }
func (c *Client) PrivMessages() <-chan *PrivMsg { return c.privMessages }

func (c *Client) send(format string, args ...interface{}) error {
	_, err := c.conn.Write([]byte(fmt.Sprintf(format+"\r\n", args...)))
	time.Sleep(1000 * time.Millisecond)
	return err
}

func (c *Client) Connect(nick, realName string) error {
	if err := c.send("USER %s 8 * :%s", nick, nick); err != nil {
		return err
	}
	return c.send("NICK %s", nick)
}

func (c *Client) Join(channel string, key string) error {
	return c.send("JOIN %s %s", channel, key)
}

func (c *Client) JoinExclusive(channel string, key string) error {
	return c.send("JOIN %s %s 0", channel, key)
}

func (c *Client) Part(channel string, partMessage string) error {
	if partMessage != "" {
		partMessage = fmt.Sprintf(":%s", partMessage)
	}
	return c.send("PART %s %s", channel, partMessage)
}

func (c *Client) PrivMsg(target string, text string) error {
	response := fmt.Sprintf("PRIVMSG %s :%s", target, text)
	for len(text) > 0 {
		if len(response) > 400 {
			lastSpace := strings.LastIndex(response[:400], " ")
			text = response[lastSpace+1:]
			response = response[:lastSpace]
			if err := c.send(response); err != nil {
				return err
			}
		} else {
			return c.send(response)
		}
		response = fmt.Sprintf("PRIVMSG %s :%s", target, text)
	}
	return nil
}

func (c *Client) Quit(quitMessage string) error {
	if quitMessage != "" {
		quitMessage = fmt.Sprintf(":%s", quitMessage)
	}
	if err := c.send(fmt.Sprintf("QUIT %s", quitMessage)); err != nil {
		return err
	}

	close(c.close)
	c.wg.Wait()
	return nil
}

func (c *Client) Listen() error {
	c.wg.Add(1)
	defer c.wg.Done()
	reader := bufio.NewReader(c.conn)
	response := textproto.NewReader(reader)
	c.logger.Println("Ready to Listen")
	for {
		select {
		case <-c.close:
			close(c.privMessages)
			close(c.messages)
			c.logger.Println("Listen exiting")
			return nil
		default:
			line, err := response.ReadLine()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					return nil
				}
				c.logger.Println("Error calling ReadLine()")
				return fmt.Errorf("Error calling ReadLine(): %v", err)
			}
			c.logger.Println(line)
			lineElements := strings.Fields(line)
			if lineElements[0] == "PING" {
				c.send("PONG " + lineElements[1])
				c.logger.Println("PONG " + lineElements[1])
				continue
			}

			m, err := NewMessage(line)
			if err != nil {
				c.logger.Println("Error parsing raw message:", err)
			}
			switch m.Command {
			case "PRIVMSG":
				c.privMessages <- privMsgFromMessage(m)
			default:
				select {
				case c.messages <- m:
				default:
				}
			}
		}
	}
}
