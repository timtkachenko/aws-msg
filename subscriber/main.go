package main

import (
	"aws-msg/chassis"
	"github.com/aws/aws-sdk-go/aws/session"
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
	// init subscriber
	subscriber := &MySubscriber{
		&mySubscriber{name: "myqueue"},
		chassis.NewSqsMediator(awsSession, "service-name", suffix, 5),
	}
	subscriber.Start()
}
