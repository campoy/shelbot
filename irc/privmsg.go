package irc

import (
	"strings"
)

type PrivateMessage struct {
	User         string
	Nick         string
	Channel      string
	Text         string
	ReplyChannel string
}

func privMsgFromMessage(m *Message) (p *PrivateMessage) {
	p = &PrivateMessage{}
	if sourceParts := strings.SplitN(m.Origin, "!", 2); len(sourceParts) == 2 {
		p.Nick = sourceParts[0]
		p.User = sourceParts[1]
	} else {
		p.Nick = m.Origin
	}
	channelAndText := strings.SplitN(m.Parameters, ":", 2)
	p.Channel = strings.TrimSpace(channelAndText[0])
	p.Text = channelAndText[1]
	if !strings.HasPrefix(p.Channel, "#") {
		p.ReplyChannel = p.Nick
	} else {
		p.ReplyChannel = p.Channel
	}

	return
}
