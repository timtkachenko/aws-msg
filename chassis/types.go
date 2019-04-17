package chassis

import (
	"github.com/aws/aws-sdk-go/service/sqs"
)

const EventType string = "eventType"

type (
	Queue struct {
		QueueUrl *string
		QueueArn *string
	}
	SubscriptionContext struct {
		Topic                  string
		Queue                  string
		DeadLetterQueue        string
		SubscriptionAttributes map[string]interface{}
	}
	Message struct {
		*sqs.Message
		Body   *MessageBody
		Remove func(*Message) error
	}
	MessageBody struct {
		Type              string
		MessageId         string
		TopicArn          string
		Message           string
		Timestamp         string
		SignatureVersion  string
		Signature         string
		SigningCertURL    string
		UnsubscribeURL    string
		MessageAttributes map[string]interface{}
	}
)

func (m *Message) Ack(err error) error {
	if err != nil {
		return err
	}
	return m.Remove(m)
}
