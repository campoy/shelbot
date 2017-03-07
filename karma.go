package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/user"
)

var karmaDB = make(map[string]int)

func karmaIncrement(item string) int {
	karmaDB[item]++
	return karmaDB[item]
}

func karmaDecrement(item string) int {
	karmaDB[item]--
	return karmaDB[item]
}

func readKarmaFileJSON() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(usr.HomeDir + "/.shelbot.json"); err != nil {
		if os.IsNotExist(err) {
			log.Println("No karma JSON found.")
		}
		return
	}
	log.Println("Loading karma JSON from disk and populating karmaDB map.")
	karmaFileJSON, err := ioutil.ReadFile(usr.HomeDir + "/.shelbot.json")
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(karmaFileJSON, &karmaDB)
	if err != nil {
		log.Fatal(err)
	}
}

func writeKarmaFileJSON() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	marshaledKarmaData, err := json.MarshalIndent(karmaDB, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Writing karma JSON to file.")
	ioutil.WriteFile(usr.HomeDir+"/.shelbot.json", marshaledKarmaData, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
