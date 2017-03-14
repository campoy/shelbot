package irc

import (
	"strings"
)

type PrivMsg struct {
	User    string
	Nick    string
	Channel string
	Text    string
}

func privMsgFromMessage(m *Message) (p *PrivMsg) {
	p = &PrivMsg{}
	if sourceParts := strings.SplitN(m.Origin, "!", 2); len(sourceParts) == 2 {
		p.Nick = sourceParts[0]
		p.User = sourceParts[1]
	} else {
		p.Nick = m.Origin
	}
	channelAndText := strings.SplitN(m.Parameters, ":", 2)
	p.Channel = strings.TrimSpace(channelAndText[0])
	p.Text = channelAndText[1]

	return
}
