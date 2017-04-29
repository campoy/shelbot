package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/davidjpeacock/conversions"
	"github.com/davidjpeacock/shelbot/irc"

	"github.com/alsm/forecastio"
	geoip2 "github.com/oschwald/geoip2-golang"
)

var commands = make(map[string]func(*irc.PrivMsg))

func init() {
	commands["help"] = help
	commands["version"] = version
	commands["convertmph"] = convertmph
	commands["convertkmh"] = convertkmh
	commands["convertc"] = convertc
	commands["convertf"] = convertf
	commands["query"] = query
	commands["topten"] = ten
	commands["bottomten"] = ten
	commands["geoip"] = geoip
	commands["wiki"] = wiki
	commands["weather"] = weather
}

func help(m *irc.PrivMsg) {
	var coms []string
	for com := range commands {
		coms = append(coms, fmt.Sprintf("\"%s\"", com))
	}
	client.PrivMsg(m.ReplyChannel, fmt.Sprintf("%s commands available: %s", bot.Nick, strings.Join(coms, ", ")))
	client.PrivMsg(m.ReplyChannel, "Karma can be adjusted thusly: \"foo++\" and \"bar--\"")
	log.Println("Shelbot help provided.")
}

func version(m *irc.PrivMsg) {
	client.PrivMsg(m.ReplyChannel, fmt.Sprintf("%s version %s.", bot.Nick, Version))
	log.Println("Shelbot version " + Version)
}

func geoip(m *irc.PrivMsg) {
	db, err := geoip2.Open(filepath.Join(homeDir, "GeoLite2-City.mmdb"))
	if err != nil {
		log.Fatal(err)
	}
	lineElements := strings.Fields(m.Text)
	if len(lineElements) < 2 {
		response := fmt.Sprintf("Please provide a value.")
		client.PrivMsg(m.ReplyChannel, response)
		log.Println(response)
	} else {
		ip := net.ParseIP(lineElements[1])
		if ip == nil {
			if ips, err := net.LookupIP(lineElements[1]); err != nil || len(ips) == 0 {
				client.PrivMsg(m.ReplyChannel, fmt.Sprintf("I'm sorry %s, %s doesn't seem to be a valid ip address or host", m.Nick, lineElements[1]))
				return
			} else {
				ip = ips[0]
				client.PrivMsg(m.ReplyChannel, fmt.Sprintf("Resolved %s to %s", lineElements[1], ip))
			}
		}
		record, err := db.City(ip)
		if err != nil {
			log.Fatal(err)
		}
		if record == nil {
			client.PrivMsg(m.ReplyChannel, fmt.Sprintf("I'm sorry %s, I couldn't find any information for %s", m.Nick, lineElements[1]))
			return
		}
		if cityName, ok := record.City.Names["en"]; ok {
			response := fmt.Sprintf("English city name: %v", cityName)
			client.PrivMsg(m.ReplyChannel, response)
			log.Println(response)
		}
		if record.Subdivisions != nil {
			if subdivName, ok := record.Subdivisions[0].Names["en"]; ok {
				response := fmt.Sprintf("English subdivision name: %v", subdivName)
				client.PrivMsg(m.ReplyChannel, response)
				log.Println(response)
			}
		}
		if cName, ok := record.Country.Names["en"]; ok {
			response := fmt.Sprintf("English country name: %v", cName)
			client.PrivMsg(m.ReplyChannel, response)
			log.Println(response)
		}
		if cityName, ok := record.City.Names["ja"]; ok {
			response := fmt.Sprintf("Japanese city name: %v", cityName)
			client.PrivMsg(m.ReplyChannel, response)
			log.Println(response)
		}
		if record.Subdivisions != nil {
			if subdivName, ok := record.Subdivisions[0].Names["ja"]; ok {
				response := fmt.Sprintf("Japanese subdivision name: %v", subdivName)
				client.PrivMsg(m.ReplyChannel, response)
				log.Println(response)
			}
		}
		if cName, ok := record.Country.Names["ja"]; ok {
			response := fmt.Sprintf("Japanese country name: %v", cName)
			client.PrivMsg(m.ReplyChannel, response)
			log.Println(response)
		}
		response := fmt.Sprintf("ISO country code: %v", record.Country.IsoCode)
		client.PrivMsg(m.ReplyChannel, response)
		log.Println(response)
		response = fmt.Sprintf("Time zone: %v", record.Location.TimeZone)
		client.PrivMsg(m.ReplyChannel, response)
		log.Println(response)
		response = fmt.Sprintf("Coordinates: %v, %v", record.Location.Latitude, record.Location.Longitude)
		client.PrivMsg(m.ReplyChannel, response)
		log.Println(response)
		response = fmt.Sprintf("Google Maps: https://www.google.com/maps/@%v,%v,15z", record.Location.Latitude, record.Location.Longitude)
		client.PrivMsg(m.ReplyChannel, response)
		log.Println(response)
	}
}

func convertmph(m *irc.PrivMsg) {
	lineElements := strings.Fields(m.Text)
	if len(lineElements) < 2 {
		response := fmt.Sprintf("Please provide a value.")
		client.PrivMsg(m.ReplyChannel, response)
		log.Println(response)
	} else {
		i, _ := strconv.Atoi(lineElements[1])
		mph := conversions.MPH(i)
		kmh := conversions.MPHToKMH(mph)

		response := fmt.Sprintf("%s is %s", mph, kmh)
		client.PrivMsg(m.ReplyChannel, response)
		log.Println(response)
	}
}

func convertkmh(m *irc.PrivMsg) {
	lineElements := strings.Fields(m.Text)
	if len(lineElements) < 2 {
		response := fmt.Sprintf("Please provide a value.")
		client.PrivMsg(m.ReplyChannel, response)
		log.Println(response)
	} else {
		i, _ := strconv.Atoi(lineElements[1])
		kmh := conversions.KMH(i)
		mph := conversions.KMHToMPH(kmh)

		response := fmt.Sprintf("%s is %s", kmh, mph)
		client.PrivMsg(m.ReplyChannel, response)
		log.Println(response)
	}
}

func convertc(m *irc.PrivMsg) {
	lineElements := strings.Fields(m.Text)
	if len(lineElements) < 2 {
		response := fmt.Sprintf("Please provide a value.")
		client.PrivMsg(m.ReplyChannel, response)
		log.Println(response)
	} else {
		i, _ := strconv.Atoi(lineElements[1])
		c := conversions.Celsius(i)
		f := conversions.CelsiusToFahrenheit(c)

		response := fmt.Sprintf("%s is %s", c, f)
		client.PrivMsg(m.ReplyChannel, response)
		log.Println(response)
	}
}

func convertf(m *irc.PrivMsg) {
	lineElements := strings.Fields(m.Text)
	if len(lineElements) < 2 {
		response := fmt.Sprintf("Please provide a value.")
		client.PrivMsg(m.ReplyChannel, response)
		log.Println(response)
	} else {
		i, _ := strconv.Atoi(lineElements[1])
		f := conversions.Fahrenheit(i)
		c := conversions.FahrenheitToCelsius(f)

		response := fmt.Sprintf("%s is %s", f, c)
		client.PrivMsg(m.ReplyChannel, response)
		log.Println(response)
	}
}

func query(m *irc.PrivMsg) {
	lineElements := strings.Fields(m.Text)
	if len(lineElements) > 1 {
		for _, q := range lineElements[1:] {
			karmaValue := karma.query(q)
			response := fmt.Sprintf("Karma for %s is %d.", q, karmaValue)
			client.PrivMsg(m.ReplyChannel, response)
			log.Println(response)
		}
	}
}

func ten(m *irc.PrivMsg) {
	lineElements := strings.Fields(m.Text)
	var p []Pair
	for k, v := range karma.db {
		p = append(p, Pair{k, v})
	}

	switch lineElements[0] {
	case "topten":
		sort.Slice(p, func(i, j int) bool { return p[i].Value > p[j].Value })
	case "bottomten":
		sort.Slice(p, func(i, j int) bool { return p[i].Value < p[j].Value })
	}

	for i := 0; i < 10 && i < len(p); i++ {
		response := fmt.Sprintf("Karma for %s is %d.", p[i].Key, p[i].Value)
		client.PrivMsg(m.ReplyChannel, response)
		log.Println(response)
	}
}

func wiki(m *irc.PrivMsg) {
	var wikiLookup struct {
		Batchcomplete string `json:"batchcomplete"`
		Query         struct {
			Pages map[string]struct {
				Pageid               int       `json:"pageid"`
				Ns                   int       `json:"ns"`
				Title                string    `json:"title"`
				Extract              string    `json:"extract"`
				Contentmodel         string    `json:"contentmodel"`
				Pagelanguage         string    `json:"pagelanguage"`
				Pagelanguagehtmlcode string    `json:"pagelanguagehtmlcode"`
				Pagelanguagedir      string    `json:"pagelanguagedir"`
				Touched              time.Time `json:"touched"`
				Lastrevid            int       `json:"lastrevid"`
				Length               int       `json:"length"`
				Fullurl              string    `json:"fullurl"`
				Editurl              string    `json:"editurl"`
				Canonicalurl         string    `json:"canonicalurl"`
			} `json:"pages"`
		} `json:"query"`
	}

	lineElements := strings.Fields(m.Text)

	resp, err := http.Get("https://en.wikipedia.org/w/api.php?format=json&action=query&prop=extracts|info&redirects&exintro=&inprop=url&explaintext=&titles=" + html.EscapeString(strings.Join(lineElements[1:], "%20")))
	if err != nil || resp.StatusCode != 200 {
		client.PrivMsg(m.ReplyChannel, fmt.Sprintf("Sorry %s, there was an error looking up a wiki article on %s", m.Nick, strings.Join(lineElements[1:], " ")))
		return
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&wikiLookup); err != nil {
		client.PrivMsg(m.ReplyChannel, fmt.Sprintf("Sorry %s, there was an error looking up a wiki article on %s", m.Nick, strings.Join(lineElements[1:], " ")))
		return
	}
	for _, entry := range wikiLookup.Query.Pages {
		client.PrivMsg(m.ReplyChannel, strings.Split(entry.Extract, "\n")[0])
		client.PrivMsg(m.ReplyChannel, entry.Fullurl)
		log.Println("Wikipedia extract provided:", entry.Fullurl)
	}
}

func weather(m *irc.PrivMsg) {
	lineElements := strings.Fields(m.Text)
	if apiKey == "" || len(lineElements) < 2 {
		// need an airport to search for
		// Do not add key; it goes in main.go with other flag defaults
		// No api key, no weather - forecastio.io for key
		return
	}

	a := LookupAirport(lineElements[1])
	if a == nil {
		client.PrivMsg(m.ReplyChannel, fmt.Sprintf("Sorry %s, I couldn't find an airport with that code", m.Nick))
		return
	}

	c := forecastio.NewConnection(apiKey)
	c.SetUnits("si")
	f, err := c.Forecast(a.Latitude, a.Longitude, nil, false)
	if err != nil || f == nil {
		client.PrivMsg(m.ReplyChannel, fmt.Sprintf("Sorry %s, there was an error looking up the weather for %s", m.Nick, a.Name))
		log.Println(err)
		return
	}

	response := fmt.Sprintf("The weather at %s is %s and %.1fC", a.Name, f.Currently.Summary, f.Currently.Temperature)
	client.PrivMsg(m.ReplyChannel, response)
	log.Println(response)
}
