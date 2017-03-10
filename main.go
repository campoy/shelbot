package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/textproto"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
)

const version = "1.2.1"

var homeDir string

func init() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	homeDir = usr.HomeDir
}

func main() {
	var bot *config
	var k *karma
	var err error
	var karmaFunc func(string) int

	logFile, err := os.OpenFile(filepath.Join(homeDir, ".shelbot.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)

	confFile := flag.String("config", filepath.Join(homeDir, ".shelbot.conf"), "config file to be used with shelbot")
	karmaFile := flag.String("karmaFile", filepath.Join(homeDir, ".shelbot.json"), "karma db file")
	v := flag.Bool("v", false, "Prints Shelbot version")
	flag.Parse()

	if *v {
		fmt.Println("Sheldon bot version " + version)
		return
	}

	if bot, err = loadConfig(*confFile); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	if k, err = readKarmaFileJSON(*karmaFile); err != nil {
		log.Fatalf("Error loading karma DB: %s", err)
	}

	if err = bot.connect(); err != nil {
		log.Fatal(err)
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("Received SIGTERM, exiting")
		bot.conn.Write([]byte("QUIT :Bazinga!\r\n"))
		bot.conn.Close()
		if err = k.save(); err == nil {
			if f, ok := k.dbFile.(*os.File); ok {
				f.Close()
			}
		}
		os.Exit(0)
	}()

	bot.conn.Write([]byte("USER " + bot.Nick + " 8 * :" + bot.User + "\r\n"))
	bot.conn.Write([]byte("NICK " + bot.Nick + "\r\n"))
	bot.conn.Write([]byte("JOIN " + bot.Channel + "\r\n"))
	bot.conn.Write([]byte("PRIVMSG " + bot.Channel + " :Shelbot version " + version + " reporting for duty.\r\n"))

	reader := bufio.NewReader(bot.conn)
	response := textproto.NewReader(reader)
	for {
		line, err := response.ReadLine()
		if err != nil {
			break
		}
		log.Println(line)

		lineElements := strings.Fields(line)
		if lineElements[0] == "PING" {
			bot.conn.Write([]byte("PONG " + lineElements[1] + "\r\n"))
			log.Println("PONG " + lineElements[1])
			continue
		}

		if lineElements[1] == "PRIVMSG" && lineElements[3] == ":shelbot" {
			if lineElements[4] == "rank" {
				bot.conn.Write([]byte("PRIVMSG " + bot.Channel + " :Rank: \r\n"))
				log.Println("Rank: ")
				// TODO: ranking algorithm and display of top ten to channel
			}

			if lineElements[4] == "version" {
				bot.conn.Write([]byte("PRIVMSG " + bot.Channel + " :Shelbot version " + version + ".\r\n"))
				log.Println("Shelbot version " + version)
			}
			continue
		}

		if !strings.HasSuffix(line, "++") && !strings.HasSuffix(line, "--") || len(lineElements) < 2 {
			continue
		}

		var handle = strings.Trim(lineElements[len(lineElements)-1], ":+-")

		if strings.HasSuffix(line, "++") {
			karmaFunc = k.increment
		} else if strings.HasSuffix(line, "--") {
			karmaFunc = k.decrement
		}

		karmaTotal := karmaFunc(handle)
		response := fmt.Sprintf("Karma for %s now %d", handle, karmaTotal)
		bot.conn.Write([]byte(fmt.Sprintf("PRIVMSG %s :%s\r\n", bot.Channel, response)))
		log.Println(response)

		if err = k.save(); err != nil {
			log.Fatalf("Error saving karma db: %s", err)
		}
	}
	logFile.Close()
}
