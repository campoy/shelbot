package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/davidjpeacock/shelbot/irc"
)

const Version = "2.5.3"

var (
	homeDir   string
	bot       *config
	client    *irc.Client
	karma     karmaMap
	karmaFile string
	apiKey    string
	limits    = make(map[string]time.Time)
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

	confFile := flag.String("config", filepath.Join(homeDir, ".shelbot.conf"), "config file to be used with shelbot")
	flag.StringVar(&karmaFile, "karmaFile", filepath.Join(homeDir, ".shelbot.json"), "karma db file")
	debug := flag.Bool("debug", false, "Enable debug (print log to screen)")
	v := flag.Bool("v", false, "Prints Shelbot version")
	airportFile := flag.String("airportFile", filepath.Join(homeDir, "airports.csv"), "airport data csv file")
	flag.StringVar(&apiKey, "forecastioKey", "", "Forcast.io API key")
	flag.Parse()

	if !*debug {
		f, err := os.OpenFile(filepath.Join(homeDir, ".shelbot.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(f)
		irc.Debug.SetOutput(f)
		defer f.Close()
	} else {
		irc.Debug.SetOutput(os.Stdout)
	}

	if err = LoadAirports(*airportFile); err != nil {
		log.Fatalln("Error loading airports file:", err)
	}

	if *v {
		fmt.Println("Shelbot version " + Version)
		return
	}

	if bot, err = loadConfig(*confFile); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	if karma, err = readKarma(karmaFile); err != nil {
		log.Fatalf("Error loading karma DB: %s", err)
	}
	defer func() {
		if err := writeKarma(karmaFile, karma); err != nil {
			log.Printf("could not save karma: %v", err)
		}
	}()

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", bot.Server, bot.Port))
	if err != nil {
		log.Fatalf("could not connect to %s:%d: %v", bot.Server, bot.Port, err)
	}
	defer conn.Close()

	client = irc.New(conn)
	if err = client.Connect(bot.Nick, bot.User); err != nil {
		log.Fatal(err)
	}

	go handleSigterm()

	client.Join(bot.Channel, "")
	client.PrivMsg(bot.Channel, fmt.Sprintf("%s version %s reporting for duty", bot.Nick, Version))

	go handleMessages(client.PrivMessages())

	if err := client.Listen(); err != nil {
		log.Fatal(err)
	}
}

func handleSigterm() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Received SIGTERM, exiting")
	if err := client.Quit("Bazinga!"); err != nil {
		log.Fatalf("could not quit gracefully: %v", err)
	}
}

func handleMessages(msgs <-chan *irc.PrivMsg) {
	for msg := range client.PrivMessages() {
		lineElements := strings.Fields(msg.Text)

		if lineElements[0] == bot.Nick {
			if commandFunc, ok := commands[lineElements[1]]; ok {
				msg.Text = strings.Join(lineElements[1:], " ")
				commandFunc(msg)
			}
			continue
		}

		if commandFunc, ok := commands[lineElements[0]]; ok && !strings.HasPrefix(msg.Channel, "#") {
			commandFunc(msg)
			continue
		}

		var handle string
		var karmaFunc func(string) int
		switch {
		case strings.HasSuffix(msg.Text, "++"):
			handle = strings.TrimSuffix(lineElements[len(lineElements)-1], "++")
			karmaFunc = karma.increment
		case strings.HasSuffix(msg.Text, "--"):
			handle = strings.TrimSuffix(lineElements[len(lineElements)-1], "--")
			karmaFunc = karma.decrement
		default:
			continue
		}
		if lastK, ok := limits[msg.User]; (ok && lastK.Add(60*time.Second).Before(time.Now())) || !ok {
			karmaTotal := karmaFunc(handle)
			response := fmt.Sprintf("Karma for %s now %d", handle, karmaTotal)
			client.PrivMsg(msg.ReplyChannel, response)
			log.Println(response)

			if err := writeKarma(karmaFile, karma); err != nil {
				log.Fatalf("Error saving karma db: %s", err)
			}
			limits[msg.User] = time.Now()
		} else if !lastK.Add(60 * time.Second).Before(time.Now()) {
			log.Println(msg.Nick, "has already sent a karma message in the last 60 seconds")
		}
	}
}
