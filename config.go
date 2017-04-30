package main

import (
	"encoding/json"
	"io/ioutil"
	"strings"
)

type config struct {
	Server      string
	Port        uint16
	Nick        string
	User        string
	Channel     string
	Pass        string
	KarmaFile   string
	AirportFile string
	ApiKey      string
}

func (cfg *config) load(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	if !strings.HasPrefix(cfg.Channel, "#") {
		cfg.Channel = "#" + cfg.Channel
	}
	return nil
}
