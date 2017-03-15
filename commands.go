package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/davidjpeacock/conversions"
	"github.com/davidjpeacock/shelbot/irc"
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
}

func help(m *irc.PrivMsg) {
	var coms []string
	for com := range commands {
		coms = append(coms, fmt.Sprintf("\"%s\"", com))
	}
	conn.PrivMsg(bot.Channel, fmt.Sprintf("%s commands available: %s", bot.Nick, strings.Join(coms, ", ")))
	conn.PrivMsg(bot.Channel, "Karma can be adjusted thusly: \"foo++\" and \"bar--\"")
	log.Println("Shelbot help provided.")
}

func version(m *irc.PrivMsg) {
	conn.PrivMsg(bot.Channel, fmt.Sprintf("%s version %s.", bot.Nick, Version))
	log.Println("Shelbot version " + Version)
}

func convertmph(m *irc.PrivMsg) {
	lineElements := strings.Fields(m.Text)
	if len(lineElements) < 3 {
		response := fmt.Sprintf("Please provide a value.")
		conn.PrivMsg(bot.Channel, response)
		log.Println(response)
	} else {
		i, _ := strconv.Atoi(lineElements[2])
		mph := conversions.MPH(i)
		kmh := conversions.MPHToKMH(mph)

		response := fmt.Sprintf("%s is %s", mph, kmh)
		conn.PrivMsg(bot.Channel, response)
		log.Println(response)
	}
}

func convertkmh(m *irc.PrivMsg) {
	lineElements := strings.Fields(m.Text)
	if len(lineElements) < 3 {
		response := fmt.Sprintf("Please provide a value.")
		conn.PrivMsg(bot.Channel, response)
		log.Println(response)
	} else {
		i, _ := strconv.Atoi(lineElements[2])
		kmh := conversions.KMH(i)
		mph := conversions.KMHToMPH(kmh)

		response := fmt.Sprintf("%s is %s", kmh, mph)
		conn.PrivMsg(bot.Channel, response)
		log.Println(response)
	}
}

func convertc(m *irc.PrivMsg) {
	lineElements := strings.Fields(m.Text)
	if len(lineElements) < 3 {
		response := fmt.Sprintf("Please provide a value.")
		conn.PrivMsg(bot.Channel, response)
		log.Println(response)
	} else {
		i, _ := strconv.Atoi(lineElements[2])
		c := conversions.Celsius(i)
		f := conversions.CelsiusToFahrenheit(c)

		response := fmt.Sprintf("%s is %s", c, f)
		conn.PrivMsg(bot.Channel, response)
		log.Println(response)
	}
}

func convertf(m *irc.PrivMsg) {
	lineElements := strings.Fields(m.Text)
	if len(lineElements) < 3 {
		response := fmt.Sprintf("Please provide a value.")
		conn.PrivMsg(bot.Channel, response)
		log.Println(response)
	} else {
		i, _ := strconv.Atoi(lineElements[2])
		f := conversions.Fahrenheit(i)
		c := conversions.FahrenheitToCelsius(f)

		response := fmt.Sprintf("%s is %s", f, c)
		conn.PrivMsg(bot.Channel, response)
		log.Println(response)
	}
}

func query(m *irc.PrivMsg) {
	lineElements := strings.Fields(m.Text)
	if len(lineElements) > 2 {
		for _, q := range lineElements[2:] {
			karmaValue := k.query(q)
			response := fmt.Sprintf("Karma for %s is %d.", q, karmaValue)
			conn.PrivMsg(bot.Channel, response)
			log.Println(response)
		}
	}
}

func ten(m *irc.PrivMsg) {
	lineElements := strings.Fields(m.Text)
	var p []Pair
	for k, v := range k.db {
		p = append(p, Pair{k, v})
	}

	switch lineElements[1] {
	case "topten":
		sort.Slice(p, func(i, j int) bool { return p[i].Value > p[j].Value })
	case "bottomten":
		sort.Slice(p, func(i, j int) bool { return p[i].Value < p[j].Value })
	}

	for i := 0; i < 10 && i < len(p); i++ {
		response := fmt.Sprintf("Karma for %s is %d.", p[i].Key, p[i].Value)
		conn.PrivMsg(bot.Channel, response)
		log.Println(response)
	}
}
