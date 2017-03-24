# Shelbot - An IRC karma bot.

Shelbot is a simple IRC karma bot written in Golang.

[![Go Report Card](https://goreportcard.com/badge/github.com/davidjpeacock/shelbot)](https://goreportcard.com/report/github.com/davidjpeacock/shelbot)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/davidjpeacock/shelbot/master/LICENSE)

## Configuration

Create a JSON configuration file with IRC details. By default Shelbot will look for this file in `~/.shelbot.conf`, this can be changed with the command line option `-config <file>` Example:

```
{
	"server":  "irc.freenode.org",
	"port":    6667,
	"nick":    "shelbot",
	"channel": "#shelly",
	"pass":    "",
	"user":    "Sheldon Cooper"
}
```

## Command line flags

Several options are available through commandline flags. One example is data persistence; Shelbot stores karma as a JSON in the default location`~/.shelbot.json`, this can be configured with the command line option `-karmaFile <file>`

For a complete list of commandline flags, see `shelbot -h`.

## Usage with systemd

Running shelbot via systemd is a fantastic way to daemonize shelbot.  Using the provided service file, shelbot will start on bootup, and restart in the event of a crash.

1. Build shelbot for your platform of choice, copy the binary over to your system
2. Tailor the shelbot.service file provided under the systemd directory to suit your USER and GROUP
3. `cp shelbot.service /etc/systemd/system`
4. `sudo systemctl daemon-reload`
5. `systemctl start|stop|status shelbot`

Shelbot logs to `~/.shelbot.log`

## Shelbot usage

In channel or via private message, you can invoke shelbot commands as follows:

Channel method: `shelbot help`
Private message method: `help`

Karma can be increased or decreased via `foo++` and `bar--` respectively.

## Extra configuration

Certain commands require extra configuration.  These are listed as follows:

* `geoip` requires one of the databases from Maxmind.  The free to use [GeoLite2 database](http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.mmdb.gz) is a good choice.

## License and Copyright

Copyright (c) 2017 David J Peacock - david.j.peacock@gmail.com

Please see LICENSE file for details.
