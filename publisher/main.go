package main

import (
	"aws-msg/chassis"
	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/google/logger"
)

const ExchangeName = "topic"

var (
	suffix     = "stage" // development, test, production
	awsSession *session.Session
)

func newAwsSession() *session.Session {
	return session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
}

func main() {
	awsSession = newAwsSession()

	var err error
	myExchange, err := chassis.NewExchange(awsSession, suffix, chassis.SubscriptionContext{Topic: ExchangeName})
	if err != nil {
		log.Fatalln(err)
	}
	myPublisher := &MyEventPublisher{myExchange}

	if err := myPublisher.Publish("hello"); err != nil {
		log.Fatalln(err)
	}
}
