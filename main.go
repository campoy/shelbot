package main

import (
	"bufio"
	"log"
	"net"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"
)

const VERSION = "1.0.0"

type Config struct {
	server        string
	port          string
	nick          string
	user          string
	channel       string
	pass          string
	pread, pwrite chan string
	conn          net.Conn
}

func LaunchBot() *Config {
	return &Config{server: "irc.freenode.org",
		port:    "6667",
		nick:    "shelbot",
		channel: "#shelly",
		pass:    "",
		conn:    nil,
		user:    "Sheldon Cooper"}
}

func (bot *Config) Connect() (conn net.Conn, err error) {
	conn, err = net.Dial("tcp", bot.server+":"+bot.port)
	if err != nil {
		log.Fatal("Failed to connect to IRC server", err)
	}
	bot.conn = conn
	log.Println("Connected to IRC server", bot.server, bot.conn.RemoteAddr())
	return bot.conn, nil
}

func main() {
	ReadKarmaFileJSON()

	ircbot := LaunchBot()
	conn, _ := ircbot.Connect()
	conn.Write([]byte("USER " + ircbot.nick + " 8 * :" + ircbot.nick + "\r\n"))
	conn.Write([]byte("NICK " + ircbot.nick + "\r\n"))
	conn.Write([]byte("JOIN " + ircbot.channel + "\r\n"))
	conn.Write([]byte("PRIVMSG " + ircbot.channel + " :Sheldon bot version " + VERSION + " reporting for duty.\r\n"))
	defer conn.Close()

	ri := regexp.MustCompile(`[A-z]\+\+$`) // Karma increment
	rd := regexp.MustCompile(`[A-z]\-\-$`) // Karma decrement

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
		} else if ri.MatchString(lineElements[len(lineElements)-1]) {
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
			karmaTotal = KarmaIncrement(handle)
			if err != nil {
				log.Println("Error: ", err)
			}
			karmaString := strconv.Itoa(karmaTotal)
			conn.Write([]byte("PRIVMSG " + ircbot.channel + " :Karma for " + handle + " now " + karmaString + "\r\n"))
			log.Println("Karma for " + handle + " now " + karmaString)

			WriteKarmaFileJSON()
		} else if rd.MatchString(lineElements[len(lineElements)-1]) {
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
			karmaTotal = KarmaDecrement(handle)
			if err != nil {
				log.Println("Error: ", err)
			}
			karmaString := strconv.Itoa(karmaTotal)
			conn.Write([]byte("PRIVMSG " + ircbot.channel + " :Karma for " + handle + " now " + karmaString + "\r\n"))
			log.Println("Karma for " + handle + " now " + karmaString)

			WriteKarmaFileJSON()
		}
	}
}
