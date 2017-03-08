package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"os"
	"os/user"
	"strings"
)

const version = "1.1.0"

type config struct {
	server        string
	port          string
	nick          string
	user          string
	channel       string
	pass          string
	pread, pwrite chan string
	conn          net.Conn
}

func launchBot() *config {
	return &config{server: "irc.freenode.org",
		port:    "6667",
		nick:    "shelbot",
		channel: "#shelly",
		pass:    "",
		conn:    nil,
		user:    "Sheldon Cooper"}
}

func (bot *config) connect() (conn net.Conn, err error) {
	conn, err = net.Dial("tcp", bot.server+":"+bot.port)
	if err != nil {
		log.Fatal("Failed to connect to IRC server", err)
	}
	bot.conn = conn
	log.Println("Connected to IRC server", bot.server, bot.conn.RemoteAddr())
	return bot.conn, nil
}

func main() {
	var karmaFunc func(string) int
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	logFile, err := os.OpenFile(usr.HomeDir+"/.shelbot.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)

	readKarmaFileJSON()

	ircbot := launchBot()
	conn, _ := ircbot.connect()
	conn.Write([]byte("USER " + ircbot.nick + " 8 * :" + ircbot.user + "\r\n"))
	conn.Write([]byte("NICK " + ircbot.nick + "\r\n"))
	conn.Write([]byte("JOIN " + ircbot.channel + "\r\n"))
	conn.Write([]byte("PRIVMSG " + ircbot.channel + " :Sheldon bot version " + version + " reporting for duty.\r\n"))
	defer conn.Close()

	reader := bufio.NewReader(conn)
	response := textproto.NewReader(reader)
	for {
		line, err := response.ReadLine()
		if err != nil {
			break
		}
		log.Println(line)

		lineElements := strings.Fields(line)
		if lineElements[0] == "PING" {
			conn.Write([]byte("PONG " + lineElements[1] + "\r\n"))
			log.Println("PONG " + lineElements[1])
			continue
		}

		if !strings.HasSuffix(line, "++") && !strings.HasSuffix(line, "--") || len(lineElements) < 2 {
			continue
		}

		var handle = strings.Trim(lineElements[len(lineElements)-1], ":+-")

		if strings.HasSuffix(line, "++") {
			karmaFunc = karmaIncrement
		} else if strings.HasSuffix(line, "--") {
			karmaFunc = karmaDecrement
		}

		karmaTotal := karmaFunc(handle)
		response := fmt.Sprintf("Karma for %s now %d", handle, karmaTotal)
		conn.Write([]byte(fmt.Sprintf("PRIVMSG %s :%s\r\n", ircbot.channel, response)))
		log.Println(response)

		writeKarmaFileJSON()
	}
	logFile.Close()
}
