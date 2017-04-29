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

var (
	Debug = log.New(ioutil.Discard, "IRC: ", log.LstdFlags)
)

type Client struct {
	wg           sync.WaitGroup
	conn         io.ReadWriteCloser
	close        chan struct{}
	messages     chan *Message
	privMessages chan *PrivMsg
}

func New(conn io.ReadWriteCloser) *Client {
	return &Client{
		conn:         conn,
		close:        make(chan struct{}),
		messages:     make(chan *Message),
		privMessages: make(chan *PrivMsg),
	}
}

func (c *Client) Messages() <-chan *Message     { return c.messages }
func (c *Client) PrivMessages() <-chan *PrivMsg { return c.privMessages }

func (c *Client) send(line string) error {
	_, err := c.conn.Write([]byte(fmt.Sprintf("%s\r\n", line)))
	time.Sleep(1000 * time.Millisecond)
	return err
}

func (c *Client) Connect(nick, realName string) error {
	if err := c.send("USER " + nick + " 8 * :" + nick); err != nil {
		return fmt.Errorf("USER command: %v", err)
	}
	if err := c.send("NICK " + nick); err != nil {
		return fmt.Errorf("NICK command: %v", err)
	}
	return nil
}

func (c *Client) Join(channel string, key string) error {
	return c.send(fmt.Sprintf("JOIN %s %s", channel, key))
}

func (c *Client) JoinExclusive(channel string, key string) error {
	return c.send(fmt.Sprintf("JOIN %s %s 0", channel, key))
}

func (c *Client) Part(channel string, partMessage string) error {
	if partMessage != "" {
		partMessage = fmt.Sprintf(":%s", partMessage)
	}
	return c.send(fmt.Sprintf("PART %s %s", channel, partMessage))
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
	c.conn.Close()
	c.wg.Wait()
	return nil
}

func (c *Client) Listen() error {
	c.wg.Add(1)
	defer c.wg.Done()
	reader := bufio.NewReader(c.conn)
	response := textproto.NewReader(reader)
	Debug.Println("Ready to Listen")
	for {
		select {
		case <-c.close:
			Debug.Println("Listen exiting")
			return nil
		default:
			line, err := response.ReadLine()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					return nil
				}
				Debug.Println("Error calling ReadLine()")
				return fmt.Errorf("Error calling ReadLine(): %v", err)
			}
			Debug.Println(line)
			lineElements := strings.Fields(line)
			if lineElements[0] == "PING" {
				c.send("PONG " + lineElements[1])
				Debug.Println("PONG " + lineElements[1])
				continue
			}

			m, err := NewMessage(line)
			if err != nil {
				Debug.Println("Error parsing raw message:", err)
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
