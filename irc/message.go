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

func newMessage(raw string) (*Message, error) {
	m := &Message{}

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

	return m, nil
}
