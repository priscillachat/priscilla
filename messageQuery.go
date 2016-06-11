package main

import (
	"strings"
)

type messageBlock struct {
	Message       string    `json:"message,omitempty"`
	From          string    `json:"from,omitempty"`
	Room          string    `json:"room,omitempty"`
	Mentioned     bool      `json:"mentioned,omitempty"`
	Stripped      string    `json:"stripped,omitempty"`
	MentionNotify []string  `json:"mentionnotify,omitempty"`
	User          *UserInfo `json:"user,omitempty"`
}

type UserInfo struct {
	Id      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Mention string `string:"mention,omitempty"`
	Email   string `string:"email,omitempty"`
}

func (m *messageBlock) handleMessage(source string,
	dispatch chan<- *dispatcherRequest) {

	logger.Debug.Println("Message: ", m.Message)
	logger.Debug.Println("From: ", m.From)
	logger.Debug.Println("Room: ", m.Room)

	prefixMatch := false

	if len(m.Message) > conf.prefixLen &&
		m.Message[0:conf.prefixLen] == conf.Prefix {

		prefixMatch = true
	}

	if prefixMatch {
		logger.Debug.Println("Prefix matched!")
		trimmed := strings.TrimLeft(m.Message[conf.prefixLen:], " ")

		if !checkHelp(trimmed, source, m.Room, dispatch) {
			if !triggerActiveResponders(prefixAResponders, trimmed, source, m,
				false, dispatch) {

				triggerPassiveResponders(prefixPResponders, trimmed, source,
					m.Room, m.From, false, dispatch)
			}
		}
	} else {
		logger.Debug.Println("Non-prefix match triggered!")

		if !triggerActiveResponders(noPrefixAResponders, m.Message, source, m,
			false, dispatch) {

			matched := triggerPassiveResponders(noPrefixPResponders, m.Message,
				source, m.Room, m.From, false, dispatch)
			if !matched && m.Mentioned {
				trimmed := strings.TrimLeft(m.Stripped, " ")

				if !checkHelp(trimmed, source, m.Room, dispatch) {
					logger.Debug.Println("Mention match triggered!")

					if !triggerActiveResponders(mentionAResponders, m.Message,
						source, m, true, dispatch) {

						triggerPassiveResponders(mentionPResponders, trimmed,
							source, m.Room, m.From, true, dispatch)
					}
				}
			}
		}
	}
}
