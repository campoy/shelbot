package main

import (
	"encoding/json"
	"io/ioutil"
	"strings"
)

type config struct {
	Server        string `json:"server"`
	Port          uint16 `json:"port"`
	Nick          string `json:"nick"`
	User          string `json:"user"`
	Channel       string `json:"channel"`
	Pass          string `json:"pass"`
	pread, pwrite chan string
}

func loadConfig(confFile string) (*config, error) {
	data, err := ioutil.ReadFile(confFile)
	if err != nil {
		return nil, err
	}

	var c config
	if err = json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	if err = c.validate(); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *config) validate() error {
	if !strings.HasPrefix(c.Channel, "#") {
		c.Channel = "#" + c.Channel
	}

	return nil
}
