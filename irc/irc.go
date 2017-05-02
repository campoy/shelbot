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
	conn         io.ReadWriter
	close        chan struct{}
	Messages     chan *Message
	PrivMessages chan *PrivMsg
}

func New(conn io.ReadWriter) *Client {
	return &Client{
		conn:         conn,
		close:        make(chan struct{}),
		Messages:     make(chan *Message),
		PrivMessages: make(chan *PrivMsg),
	}
}

func (c *Client) send(line string) {
	c.conn.Write([]byte(fmt.Sprintf("%s\r\n", line)))
	time.Sleep(1000 * time.Millisecond)
}

func (c *Client) Connect(nick, realName string) error {
	c.send("USER " + nick + " 8 * :" + realName)
	c.send("NICK " + nick)
	return nil
}

func (c *Client) Join(channel string, key string) {
	c.send(fmt.Sprintf("JOIN %s %s", channel, key))
}

func (c *Client) JoinExclusive(channel string, key string) {
	c.send(fmt.Sprintf("JOIN %s %s 0", channel, key))
}

func (c *Client) Part(channel string, partMessage string) {
	if partMessage != "" {
		partMessage = fmt.Sprintf(":%s", partMessage)
	}
	c.send(fmt.Sprintf("PART %s %s", channel, partMessage))
}

func (c *Client) PrivMsg(target string, text string) {
	response := fmt.Sprintf("PRIVMSG %s :%s", target, text)
	for len(text) > 0 {
		if len(response) > 400 {
			lastSpace := strings.LastIndex(response[:400], " ")
			text = response[lastSpace+1:]
			response = response[:lastSpace]
			c.send(response)
		} else {
			c.send(response)
			return
		}
		response = fmt.Sprintf("PRIVMSG %s :%s", target, text)
	}
}

func (c *Client) Quit(quitMessage string) {
	if quitMessage != "" {
		quitMessage = fmt.Sprintf(":%s", quitMessage)
	}
	c.send(fmt.Sprintf("QUIT %s", quitMessage))

	close(c.close)
	c.wg.Wait()
}

func (c *Client) Listen() {
	c.wg.Add(1)
	defer c.wg.Done()
	reader := bufio.NewReader(c.conn)
	response := textproto.NewReader(reader)
	Debug.Println("Ready to Listen")
	for {
		select {
		case <-c.close:
			Debug.Println("Listen exiting")
			return
		default:
			line, err := response.ReadLine()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					return
				}
				Debug.Println("Error calling ReadLine()")
				Debug.Fatalln(err)
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
				c.PrivMessages <- privMsgFromMessage(m)
			default:
				select {
				case c.Messages <- m:
				default:
				}
			}
		}
	}
}
