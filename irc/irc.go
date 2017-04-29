package irc

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/textproto"
	"strings"
	"sync"
	"time"
)

var (
	Debug = log.New(ioutil.Discard, "IRC: ", log.LstdFlags)
)

type Conn struct {
	wg           sync.WaitGroup
	conn         net.Conn
	server       string
	close        chan struct{}
	Messages     chan *Message
	PrivMessages chan *PrivMsg
}

func New(server string, port uint16) *Conn {
	return &Conn{
		server:       fmt.Sprintf("%s:%d", server, port),
		close:        make(chan struct{}),
		Messages:     make(chan *Message),
		PrivMessages: make(chan *PrivMsg),
	}
}

func (c *Conn) send(line string) error {
	_, err := c.conn.Write([]byte(fmt.Sprintf("%s\r\n", line)))
	time.Sleep(1000 * time.Millisecond)
	return err
}

func (c *Conn) Connect(nick, realName string) error {
	var err error

	c.conn, err = net.Dial("tcp", c.server)
	if err != nil {
		return fmt.Errorf("Failed to connect to IRC server: %s", err)
	}
	Debug.Println("Connected to IRC server", c.server, c.conn.RemoteAddr())

	c.send("USER " + nick + " 8 * :" + nick)
	c.send("NICK " + nick)
	return nil
}

func (c *Conn) Join(channel string, key string) error {
	return c.send(fmt.Sprintf("JOIN %s %s", channel, key))
}

func (c *Conn) JoinExclusive(channel string, key string) error {
	return c.send(fmt.Sprintf("JOIN %s %s 0", channel, key))
}

func (c *Conn) Part(channel string, partMessage string) error {
	if partMessage != "" {
		partMessage = fmt.Sprintf(":%s", partMessage)
	}
	return c.send(fmt.Sprintf("PART %s %s", channel, partMessage))
}

func (c *Conn) PrivMsg(target string, text string) error {
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

func (c *Conn) Quit(quitMessage string) error {
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

func (c *Conn) Listen() error {
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
