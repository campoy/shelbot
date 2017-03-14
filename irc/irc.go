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
)

var (
	Debug = log.New(ioutil.Discard, "IRC: ", log.LstdFlags)
)

type Conn struct {
	wg           sync.WaitGroup
	conn         net.Conn
	server       string
	nick         string
	realName     string
	close        chan struct{}
	Messages     chan *Message
	PrivMessages chan *PrivMsg
}

func New(server string, port uint16, nick string, realName string) *Conn {
	c := &Conn{
		server:       fmt.Sprintf("%s:%d", server, port),
		nick:         nick,
		realName:     realName,
		close:        make(chan struct{}),
		Messages:     make(chan *Message),
		PrivMessages: make(chan *PrivMsg),
	}

	return c
}

func (c *Conn) send(line string) {
	c.conn.Write([]byte(fmt.Sprintf("%s\r\n", line)))
}

func (c *Conn) Connect() error {
	var err error

	c.conn, err = net.Dial("tcp", c.server)
	if err != nil {
		return fmt.Errorf("Failed to connect to IRC server: %s", err)
	}
	Debug.Println("Connected to IRC server", c.server, c.conn.RemoteAddr())

	c.send("USER " + c.nick + " 8 * :" + c.nick)
	c.send("NICK " + c.nick)
	return nil
}

func (c *Conn) Join(channel string, key string) {
	c.send(fmt.Sprintf("JOIN %s %s", channel, key))
}

func (c *Conn) JoinExclusive(channel string, key string) {
	c.send(fmt.Sprintf("JOIN %s %s 0", channel, key))
}

func (c *Conn) Part(channel string, partMessage string) {
	if partMessage != "" {
		partMessage = fmt.Sprintf(":%s", partMessage)
	}
	c.send(fmt.Sprintf("PART %s %s", channel, partMessage))
}

func (c *Conn) PrivMsg(target string, text string) {
	c.send(fmt.Sprintf("PRIVMSG %s :%s", target, text))
}

func (c *Conn) Quit(quitMessage string) {
	if quitMessage != "" {
		quitMessage = fmt.Sprintf(":%s", quitMessage)
	}
	c.send(fmt.Sprintf("QUIT :%s", quitMessage))

	close(c.close)
	c.conn.Close()
	c.wg.Wait()
}

func (c *Conn) Listen() {
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
