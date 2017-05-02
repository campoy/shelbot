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
	homeDir string
	bot     *config
	client  *irc.Client
	k       *karma
	apiKey  string
	limits  = make(map[string]time.Time)
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

	confFile := flag.String("config", filepath.Join(homeDir, ".shelbot.conf"), "config file to be used with shelbot")
	karmaFile := flag.String("karmaFile", filepath.Join(homeDir, ".shelbot.json"), "karma db file")
	debug := flag.Bool("debug", false, "Enable debug (print log to screen)")
	v := flag.Bool("v", false, "Prints Shelbot version")
	airportFile := flag.String("airportFile", filepath.Join(homeDir, "airports.csv"), "airport data csv file")
	flag.StringVar(&apiKey, "forecastioKey", "", "Forcast.io API key")
	flag.Parse()

	logger := log.New(os.Stdout, "IRC: ", log.LstdFlags)
	if !*debug {
		logFile, err = os.OpenFile(filepath.Join(homeDir, ".shelbot.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(logFile)
		logger.SetOutput(logFile)
		defer logFile.Close()
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

	if k, err = readKarmaFileJSON(*karmaFile); err != nil {
		log.Fatalf("Error loading karma DB: %s", err)
	}

	netConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", bot.Server, bot.Port))
	if err != nil {
		log.Fatalf("Failed to connect to IRC server: %s", err)
	}
	defer netConn.Close()

	log.Println("Connected to IRC server", fmt.Sprintf("%s:%d", bot.Server, bot.Port), netConn.RemoteAddr())

	client = irc.New(netConn,
		irc.WithPause(500*time.Millisecond),
		irc.WithLogger(logger))

	if err = client.Connect(bot.Nick, bot.User); err != nil {
		log.Fatal(err)
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("Received SIGTERM, exiting")
		if err := client.Quit("Bazinga!"); err != nil {
			log.Printf("Could not exit gracefully: %v", err)
		}
	}()

	if err := client.Join(bot.Channel, ""); err != nil {
		log.Fatalf("could not join channel: %v", err)
	}
	err = client.Send(bot.Channel, fmt.Sprintf("%s version %s reporting for duty", bot.Nick, Version))
	if err != nil {
		log.Fatalf("Could not send hello: %v", err)
	}

	go handleMessages(client.PrivateMessages())

	listenErr := client.Listen()
	if err = k.save(); err == nil {
		if f, ok := k.dbFile.(*os.File); ok {
			f.Close()
		}
	}

	if listenErr != nil {
		log.Fatal(err)
	}
}

func handleMessages(msgs <-chan *irc.PrivateMessage) {
	for msg := range msgs {
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
			karmaFunc = k.increment
		case strings.HasSuffix(msg.Text, "--"):
			handle = strings.TrimSuffix(lineElements[len(lineElements)-1], "--")
			karmaFunc = k.decrement
		default:
			continue
		}
		if lastK, ok := limits[msg.User]; (ok && lastK.Add(60*time.Second).Before(time.Now())) || !ok {
			karmaTotal := karmaFunc(handle)
			response := fmt.Sprintf("Karma for %s now %d", handle, karmaTotal)
			if err := client.Send(msg.ReplyChannel, response); err != nil {
				log.Printf("Could not send message: %v", err)
				continue
			}
			log.Println(response)

			if err := k.save(); err != nil {
				log.Fatalf("Error saving karma db: %s", err)
			}
			limits[msg.User] = time.Now()
		} else if !lastK.Add(60 * time.Second).Before(time.Now()) {
			log.Println(msg.Nick, "has already sent a karma message in the last 60 seconds")
		}
	}
}
