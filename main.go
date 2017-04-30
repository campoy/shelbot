package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/davidjpeacock/shelbot/irc"
)

const Version = "2.5.3"

type bot struct {
	*irc.Client

	config   config
	karma    karma
	airports airports
}

func newBot(conf string) (*bot, error) {
	var b bot
	if err := b.config.load(conf); err != nil {
		return nil, fmt.Errorf("could not read config file: %s", err)
	}
	if err := b.karma.load(b.config.KarmaFile); err != nil {
		return nil, fmt.Errorf("could not read karma from %s: %v", b.config.KarmaFile, err)
	}
	if err := b.airports.Load(b.config.AirportFile); err != nil {
		log.Fatalf("could not load airports: %v", err)
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", b.config.Server, b.config.Port))
	if err != nil {
		return nil, fmt.Errorf("could not connect to %s:%d: %v", b.config.Server, b.config.Port, err)
	}
	b.Client = irc.New(conn)
	if err = b.Connect(b.config.Nick, b.config.User); err != nil {
		return nil, fmt.Errorf("could not connect: %v", err)
	}

	return &b, nil
}

func main() {
	var (
		confFile = flag.String("config", ".shelconfig.conf", "config file to be used with shelbot")
		version  = flag.Bool("v", false, "Prints Shelbot version")
	)
	flag.Parse()

	if *version {
		fmt.Println("Shelbot version " + Version)
		return
	}

	b, err := newBot(*confFile)
	if err != nil {
		log.Fatal(err)
	}

	go b.handleSigterm()
	go b.handleMessages()

	b.Join(b.config.Channel, "")
	b.Send(b.config.Channel, fmt.Sprintf("%s version %s reporting for duty", b.config.Nick, Version))

	listenErr := b.Listen()

	if err := b.karma.write(b.config.KarmaFile); err != nil {
		log.Printf("could not save karma: %v", err)
	}

	if listenErr != nil {
		log.Fatal(err)
	}
}

func (b *bot) handleSigterm() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Received SIGTERM, exiting")
	if err := b.Quit("Bazinga!"); err != nil {
		log.Fatalf("could not quit gracefully: %v", err)
	}
}

func (b *bot) handleMessages() {
	limits := make(map[string]time.Time)
	for msg := range b.PrivateMessages() {
		lineElements := strings.Fields(msg.Text)

		if lineElements[0] == b.config.Nick {
			if f, ok := commands[lineElements[1]]; ok {
				msg.Text = strings.Join(lineElements[1:], " ")
				f(b, msg)
			}
			continue
		}

		if f, ok := commands[lineElements[0]]; ok && !strings.HasPrefix(msg.Channel, "#") {
			f(b, msg)
			continue
		}

		var handle string
		var karmaFunc func(string) int
		switch {
		case strings.HasSuffix(msg.Text, "++"):
			handle = strings.TrimSuffix(lineElements[len(lineElements)-1], "++")
			karmaFunc = b.karma.increment
		case strings.HasSuffix(msg.Text, "--"):
			handle = strings.TrimSuffix(lineElements[len(lineElements)-1], "--")
			karmaFunc = b.karma.decrement
		default:
			continue
		}
		if lastK, ok := limits[msg.User]; (ok && lastK.Add(60*time.Second).Before(time.Now())) || !ok {
			karmaTotal := karmaFunc(handle)
			response := fmt.Sprintf("Karma for %s now %d", handle, karmaTotal)
			b.Send(msg.ReplyChannel, response)
			log.Println(response)
			limits[msg.User] = time.Now()
		} else if !lastK.Add(60 * time.Second).Before(time.Now()) {
			log.Println(msg.Nick, "has already sent a karma message in the last 60 seconds")
		}
	}
}
