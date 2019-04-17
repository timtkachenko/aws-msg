package main

import (
	"aws-msg/chassis"
	log "github.com/google/logger"
	"runtime"
	"time"
)

//go:generate mockgen -destination=./mocks/subscriber.go -source=./subscriber.go -package=mocks

type (
	Mediator interface {
		Receive(settings chassis.SubscriptionContext) ([]*chassis.Message, error)
		GetWaitTimeSeconds() int64
	}
	MessageHandler interface {
		Handle(message *chassis.Message) error
		GetSubscriptionContext() chassis.SubscriptionContext
	}
	MySubscriber struct {
		MessageHandler
		Mediator
	}
)

func (ms *MySubscriber) Start() {
	for {
		ms.run()
		time.Sleep(time.Second * time.Duration(ms.GetWaitTimeSeconds()))
	}
}

func (ms *MySubscriber) run() {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
		}
	}()
	msgList := ms.getMessages()
	if msgList == nil {
		return
	}
	log.Info("received count: ", len(msgList))
	for _, msg := range msgList {
		log.Info("received:", msg.Body.Message)
		if err := ms.Handle(msg); err != nil {
			log.Errorf("unacknowledged: %v", err)
		}
	}
}

func (ms *MySubscriber) getMessages() []*chassis.Message {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			log.Errorf("recovered: %v", r)
		}
	}()

	msg, err := ms.Receive(ms.GetSubscriptionContext())
	if err != nil {
		panic(err)
	}
	return msg
}
