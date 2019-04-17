package main

import (
	"aws-msg/chassis"
	log "github.com/google/logger"
)

const EventType = "StatusChanged"

type mySubscriber struct {
	name string
}

func (s mySubscriber) GetSubscriptionContext() chassis.SubscriptionContext {
	return chassis.SubscriptionContext{
		Topic:           ExchangeName,
		Queue:           s.name,
		DeadLetterQueue: "DLQ",
		// comment out FilterPolicy to receive all events
		SubscriptionAttributes: map[string]interface{}{
			"FilterPolicy": map[string][]string{
				chassis.EventType: {EventType},
			},
		},
	}
}
func (s mySubscriber) Handle(msg *chassis.Message) error {
	log.Infof("handling: %v", msg.Body.Message)

	return msg.Ack(nil)
}
