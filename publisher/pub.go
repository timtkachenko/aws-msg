package main

import (
	"aws-msg/chassis"
	"encoding/json"
	log "github.com/google/logger"
)

const EventType = "StatusChanged"

type (
	MyEventPublisher struct {
		*chassis.Exchange
	}

	MessageInput struct {
		Body interface{} `json:"body"`
	}
)

func (ep *MyEventPublisher) Publish(payload string) (err error) {
	msg, err := json.Marshal(&MessageInput{
		Body: payload,
	})
	if err != nil {
		log.Fatalln(err)
		return
	}
	log.Info("sent: ", string(msg))
	_, err = ep.Exchange.Publish(string(msg), EventType)
	if err != nil {
		log.Fatalln(err)
	}
	return
}
