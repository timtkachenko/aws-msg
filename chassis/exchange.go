package chassis

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"time"
)

type Exchange struct {
	client   *sns.SNS
	suffix   string
	settings SubscriptionContext
	TopicArn *string
}

func NewExchange(session *session.Session, suffix string, settings SubscriptionContext) (*Exchange, error) {
	exch := &Exchange{
		client:   sns.New(session),
		suffix:   suffix,
		settings: settings,
	}
	_, err := exch.getTopicArn(settings.Topic)
	if err != nil {
		return nil, err
	}
	return exch, nil
}

func (ex *Exchange) buildTopicName(topic string) string {
	return topic + "-" + ex.suffix
}

func (ex *Exchange) getTopicArn(source string) (arn *string, err error) {
	if ex.TopicArn == nil {
		name := ex.buildTopicName(source)
		ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
		defer cancel()
		result, err := ex.client.CreateTopicWithContext(ctx, &sns.CreateTopicInput{Name: aws.String(name)})
		if err != nil {
			return arn, err
		}
		ex.TopicArn = result.TopicArn
	}
	return ex.TopicArn, nil
}

func (ex *Exchange) Subscribe(qArn *string) (*sns.SubscribeOutput, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	result, err := ex.client.SubscribeWithContext(ctx, &sns.SubscribeInput{
		TopicArn: ex.TopicArn,
		Endpoint: qArn,
		Protocol: aws.String("sqs"),
	})
	if err != nil {
		return nil, err
	}
	if err := ex.SetSubscriptionAttributes(result.SubscriptionArn, "FilterPolicy"); err != nil {
		return nil, err
	}
	return result, nil
}

func (ex *Exchange) SetSubscriptionAttributes(SubscriptionArn *string, attributeName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	attributeValue, ok := ex.settings.SubscriptionAttributes[attributeName]
	if !ok {
		//empty value
		attributeValue = struct{}{}
	}

	rawValue, err := json.Marshal(attributeValue)
	if err != nil {
		return err
	}
	strValue := string(rawValue)
	_, err = ex.client.SetSubscriptionAttributesWithContext(ctx, &sns.SetSubscriptionAttributesInput{
		SubscriptionArn: SubscriptionArn,
		AttributeName:   &attributeName,
		AttributeValue:  &strValue,
	})
	if err != nil {
		return err
	}
	return nil
}

func (ex *Exchange) Publish(payload string, eventType string) (*sns.PublishOutput, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	resp, err := ex.client.PublishWithContext(ctx, &sns.PublishInput{
		Message: aws.String(payload),
		MessageAttributes: map[string]*sns.MessageAttributeValue{
			EventType: {
				DataType:    aws.String("String"),
				StringValue: aws.String(eventType),
			},
		},
		TopicArn: ex.TopicArn,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
