package main

import (
	"bufio"
	"log"
	"net"
	"net/textproto"
	"os"
	"os/user"
	"regexp"
	"strconv"
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

	rKarmaIncrement := regexp.MustCompile(`[A-z]\+\+$`) // matches string++ at EOL
	rKarmaDecrement := regexp.MustCompile(`[A-z]\-\-$`) // matches string-- at EOL

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
		} else if rKarmaIncrement.MatchString(lineElements[len(lineElements)-1]) {
			var handle = lineElements[len(lineElements)-1]
			handle = strings.TrimPrefix(handle, ":")
			if err != nil {
				log.Println("Failed to trim prefix", err)
			}
			handle = strings.TrimSuffix(handle, "++")
			if err != nil {
				log.Println("Failed to trim suffix", err)
			}
			karmaTotal := 0
			karmaTotal = karmaIncrement(handle)
			if err != nil {
				log.Println("Error: ", err)
			}
			karmaString := strconv.Itoa(karmaTotal)
			conn.Write([]byte("PRIVMSG " + ircbot.channel + " :Karma for " + handle + " now " + karmaString + "\r\n"))
			log.Println("Karma for " + handle + " now " + karmaString)

			writeKarmaFileJSON()
		} else if rKarmaDecrement.MatchString(lineElements[len(lineElements)-1]) {
			var handle = lineElements[len(lineElements)-1]
			handle = strings.TrimPrefix(handle, ":")
			if err != nil {
				log.Println("Failed to trim prefix", err)
			}
			handle = strings.TrimSuffix(handle, "--")
			if err != nil {
				log.Println("Failed to trim suffix", err)
			}
			karmaTotal := 0
			karmaTotal = karmaDecrement(handle)
			if err != nil {
				log.Println("Error: ", err)
			}
			karmaString := strconv.Itoa(karmaTotal)
			conn.Write([]byte("PRIVMSG " + ircbot.channel + " :Karma for " + handle + " now " + karmaString + "\r\n"))
			log.Println("Karma for " + handle + " now " + karmaString)

			writeKarmaFileJSON()
		}
	}
	logFile.Close()
}
