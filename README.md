# Shelbot - An IRC karma bot.

Shelbot is a simple IRC karma bot written in Golang.

[![Build Status](https://travis-ci.org/davidjpeacock/shelbot.svg?branch=master)](https://travis-ci.org/davidjpeacock/shelbot)
[![Go Report Card](https://goreportcard.com/badge/github.com/davidjpeacock/shelbot)](https://goreportcard.com/report/github.com/davidjpeacock/shelbot)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/davidjpeacock/shelbot/master/LICENSE)

## Configuration

Alter the configuration struct in main.go to suit.  Example:

```
func LaunchBot() *Config {
	return &Config{server: "irc.freenode.org",
		port:    "6667",
		nick:    "shelbot",
		channel: "#shelly",
		pass:    "",
		conn:    nil,
		user:    "Sheldon Cooper"}
}
```

## Usage with systemd

1. Build shelbot for your platform of choice, copy the binary over to your system
2. Tailor the shelbot.service file provided under the systemd directory to suit your USER and GROUP
3. `cp shelbot.service /etc/systemd/system`
4. `sudo systemctl daemon-reload`

## Karma usage

Shelbot's lexer is currently very simple and limited.  Increasing and decreasing karma is done idiomatically.

`string++`

`string--`

For data persistence, Shelbot stores karma as a JSON in `~/.shelbot.json`

## License and Copyright

Copyright (c) 2017 David J Peacock - david.j.peacock@gmail.com

Please see LICENSE file for details.
