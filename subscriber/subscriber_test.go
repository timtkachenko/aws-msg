package main

import (
	"aws-msg/chassis"
	mocks "aws-msg/subscriber/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

var messageBodyStr = `{"hello":"world"}`

func buildMsg() *chassis.Message {
	msg := &chassis.Message{Body: &chassis.MessageBody{}}
	msg.Body.Message = messageBodyStr
	msg.Remove = func(message *chassis.Message) error {
		return nil
	}
	return msg
}

func TestHandleMsg(t *testing.T) {
	a := assert.New(t)

	msg := buildMsg()

	ctrl := gomock.NewController(t)
	mediator := mocks.NewMockMediator(ctrl)
	mediator.EXPECT().Receive(chassis.SubscriptionContext{}).Return([]*chassis.Message{msg}, nil)

	subs := mocks.NewMockMessageHandler(ctrl)
	subs.EXPECT().GetSubscriptionContext()
	subs.EXPECT().Handle(msg)

	am := &MySubscriber{subs, mediator}

	msgList := am.getMessages()
	for _, msg := range msgList {
		err := am.Handle(msg)
		a.Nil(err)
	}
}
