package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
)

type config struct {
	Server        string `json:"server"`
	Port          string `json:"port"`
	Nick          string `json:"nick"`
	User          string `json:"user"`
	Channel       string `json:"channel"`
	Pass          string `json:"pass"`
	pread, pwrite chan string
	conn          net.Conn
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

func (c *config) connect() (err error) {
	c.conn, err = net.Dial("tcp", c.Server+":"+c.Port)
	if err != nil {
		return fmt.Errorf("Failed to connect to IRC server: %s", err)
	}
	log.Println("Connected to IRC server", c.Server, c.conn.RemoteAddr())
	return nil
}

func (c *config) validate() error {
	if port, err := strconv.Atoi(c.Port); err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("The port option must be a number between 1 and 65535")
	}

	if !strings.HasPrefix(c.Channel, "#") {
		c.Channel = "#" + c.Channel
	}

	return nil
}
