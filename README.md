# Shelbot - An IRC karma bot.

Shelbot is a simple IRC karma bot written in Golang.

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

## Karma usage

Shelbot's lexer is currently very simple and limited.  Increasing and decreasing karma is done idiomatically.

`string++`

`string--`

For data persistence, Shelbot stores karma as a JSON in `~/.shelbot.json`

## License and Copyright

Copyright (c) 2017 David J Peacock - david.j.peacock@gmail.com

Please see LICENSE file for details.
