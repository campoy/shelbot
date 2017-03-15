package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/davidjpeacock/shelbot/irc"
)

const Version = "2.1.1"

var (
	homeDir string
	bot     *config
	conn    *irc.Conn
	k       *karma
)

func init() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	homeDir = usr.HomeDir
}

type Pair struct {
	Key   string
	Value int
}

func main() {
	var err error
	var logFile *os.File
	var karmaFunc func(string) int

	confFile := flag.String("config", filepath.Join(homeDir, ".shelbot.conf"), "config file to be used with shelbot")
	karmaFile := flag.String("karmaFile", filepath.Join(homeDir, ".shelbot.json"), "karma db file")
	debug := flag.Bool("debug", false, "Enable debug (print log to screen)")
	v := flag.Bool("v", false, "Prints Shelbot version")
	flag.Parse()

	if !*debug {
		logFile, err = os.OpenFile(filepath.Join(homeDir, ".shelbot.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(logFile)
		irc.Debug.SetOutput(logFile)
	} else {
		irc.Debug.SetOutput(os.Stdout)
	}

	if *v {
		fmt.Println("Shelbot version " + Version)
		return
	}

	if bot, err = loadConfig(*confFile); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	if k, err = readKarmaFileJSON(*karmaFile); err != nil {
		log.Fatalf("Error loading karma DB: %s", err)
	}

	conn = irc.New(bot.Server, bot.Port, bot.Nick, bot.User)

	if err = conn.Connect(); err != nil {
		log.Fatal(err)
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("Received SIGTERM, exiting")
		conn.Quit("Bazinga!")
		if err = k.save(); err == nil {
			if f, ok := k.dbFile.(*os.File); ok {
				f.Close()
			}
		}
		if !*debug {
			logFile.Close()
		}
		os.Exit(0)
	}()

	conn.Join(bot.Channel, "")
	conn.PrivMsg(bot.Channel, fmt.Sprintf("%s version %s reporting for duty", bot.Nick, Version))

	go conn.Listen()

	for msg := range conn.PrivMessages {
		lineElements := strings.Fields(msg.Text)

		if lineElements[0] == bot.Nick {

			if commandFunc, ok := commands[lineElements[1]]; ok {
				commandFunc(msg)
			}

			continue
		}

		var handle string
		switch {
		case strings.HasSuffix(msg.Text, "++"):
			handle = strings.TrimSuffix(lineElements[len(lineElements)-1], "++")
			karmaFunc = k.increment
		case strings.HasSuffix(msg.Text, "--"):
			handle = strings.TrimSuffix(lineElements[len(lineElements)-1], "++")
			karmaFunc = k.decrement
		default:
			continue
		}

		karmaTotal := karmaFunc(handle)
		response := fmt.Sprintf("Karma for %s now %d", handle, karmaTotal)
		conn.PrivMsg(bot.Channel, response)
		log.Println(response)

		if err = k.save(); err != nil {
			log.Fatalf("Error saving karma db: %s", err)
		}
	}
}
