package irc

import (
	"fmt"
	"strconv"
	"strings"
)

type Message struct {
	Origin     string
	Command    string
	ReplyCode  int
	Parameters string
}

func parseMessage(raw string) (*Message, error) {
	var m Message

	parts := strings.Fields(raw)
	if len(parts) < 2 {
		return nil, fmt.Errorf("Received message was too short")
	}

	if parts[0][0] == ':' {
		//first element starts with a : so is a source
		m.Origin = strings.TrimPrefix(parts[0], ":")
		parts = parts[1:]
	}

	if n, err := strconv.Atoi(parts[0]); err == nil {
		m.ReplyCode = n
	} else {
		m.Command = parts[0]
	}

	m.Parameters = strings.Join(parts[1:], " ")
	return &m, nil
}

type PrivateMessage struct {
	User         string
	Nick         string
	Channel      string
	Text         string
	ReplyChannel string
}

func privMsgFromMessage(m *Message) *PrivateMessage {
	p := &PrivateMessage{
		Nick: m.Origin,
	}
	if sourceParts := strings.SplitN(m.Origin, "!", 2); len(sourceParts) == 2 {
		p.Nick, p.User = sourceParts[0], sourceParts[1]
	}

	parts := strings.SplitN(m.Parameters, ":", 2)
	p.Channel, p.Text = strings.TrimSpace(parts[0]), parts[1]
	p.ReplyChannel = p.Channel
	if !strings.HasPrefix(p.ReplyChannel, "#") {
		p.ReplyChannel = p.Nick
	}

	return p
}
