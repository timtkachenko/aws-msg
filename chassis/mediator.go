package chassis

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"sync"
	"time"
)

const (
	DelaySeconds           string = "0"
	MessageRetentionPeriod string = "259200" // 3 days
	VisibilityTimeout      int64  = 5
	WaitTimeSeconds        int64  = 5
	MaxNumberOfMessages    int64  = 10
	MaxReceiveCount        int64  = 5
)

type (
	SqsMediator struct {
		sync.RWMutex
		session         *session.Session
		client          *sqs.SQS
		exchange        *Exchange
		waitTimeSeconds int64
		prefix          string
		suffix          string
		TopicArn        *string
		SubscriptionArn *string
		queueMap        map[string]*Queue
	}
)

func NewSqsMediator(session *session.Session, prefix string, suffix string, waitTimeSeconds int64) *SqsMediator {
	return &SqsMediator{
		session:         session,
		client:          sqs.New(session),
		waitTimeSeconds: waitTimeSeconds,
		prefix:          prefix,
		suffix:          suffix,
		queueMap:        map[string]*Queue{},
	}
}

func (sm *SqsMediator) GetWaitTimeSeconds() int64 {
	if sm.waitTimeSeconds > 0 {
		return sm.waitTimeSeconds
	}
	return WaitTimeSeconds
}

func (sm *SqsMediator) makeQueueName(name string) string {
	return sm.prefix + "-" + name + "-" + sm.suffix
}
func (sm *SqsMediator) getQueueUrl(qName string) (url *string, err error) {
	name := sm.makeQueueName(qName)
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()
	if q, ok := sm.queueMap[name]; !ok || q == nil {
		urlOutput, err := sm.client.GetQueueUrlWithContext(ctx, &sqs.GetQueueUrlInput{QueueName: aws.String(name)})

		if awsError, ok := err.(awserr.Error); ok && awsError.Code() == sqs.ErrCodeQueueDoesNotExist {
			createQueueOutput, err := sm.client.CreateQueueWithContext(ctx, &sqs.CreateQueueInput{
				QueueName: aws.String(name),
				Attributes: map[string]*string{
					"DelaySeconds":           aws.String(DelaySeconds),
					"MessageRetentionPeriod": aws.String(MessageRetentionPeriod),
				},
			})
			if err != nil {
				return url, err
			}
			sm.Lock()
			defer sm.Unlock()
			sm.queueMap[name] = &Queue{QueueUrl: createQueueOutput.QueueUrl}
			return sm.queueMap[name].QueueUrl, nil
		} else if err != nil {
			return url, err
		}
		sm.Lock()
		defer sm.Unlock()
		sm.queueMap[name] = &Queue{QueueUrl: urlOutput.QueueUrl}
	}
	return sm.queueMap[name].QueueUrl, nil
}
func (sm *SqsMediator) getQueueArn(qName string) (arn *string, err error) {
	name := sm.makeQueueName(qName)
	var result *sqs.GetQueueAttributesOutput
	var url *string
	if url, err = sm.getQueueUrl(qName); err == nil {
		ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
		defer cancel()
		result, err = sm.client.GetQueueAttributesWithContext(ctx, &sqs.GetQueueAttributesInput{
			QueueUrl:       url,
			AttributeNames: []*string{aws.String("QueueArn")},
		})
	}
	if err != nil {
		return arn, err
	}
	sm.Lock()
	defer sm.Unlock()
	sm.queueMap[name].QueueArn = result.Attributes["QueueArn"]
	return sm.queueMap[name].QueueArn, nil
}

func (sm *SqsMediator) setQueueAttr(settings SubscriptionContext) (err error) {
	var (
		queueArn           *string
		queueUrl           *string
		topicArn           *string
		deadLetterQueueArn *string
	)
	queueArn, err = sm.getQueueArn(settings.Queue)
	if err != nil {
		return
	}
	queueUrl, err = sm.getQueueUrl(settings.Queue)
	if err != nil {
		return
	}
	topicArn, err = sm.exchange.getTopicArn(settings.Topic)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	policy := &map[string]interface{}{
		"Version": "2012-10-17",
		"Id":      *queueArn + "/SQSDefaultPolicy",
		"Statement": []map[string]interface{}{
			{"Sid": "Sid" + time.Now().Format("2006"),
				"Effect": "Allow",
				"Principal": map[string]string{
					"AWS": "*",
				},
				"Action":   "SQS:SendMessage",
				"Resource": queueArn,
				"Condition": map[string]interface{}{
					"ArnEquals": map[string]string{
						"aws:TopicArn": *topicArn,
					},
				},
			},
		},
	}
	policyByte, err := json.Marshal(policy)
	if err != nil {
		return err
	}
	attributes := map[string]*string{
		"Policy": aws.String(string(policyByte)),
	}
	queueAttributesInput := &sqs.SetQueueAttributesInput{
		Attributes: attributes,
		QueueUrl:   queueUrl,
	}
	if settings.DeadLetterQueue != "" {
		deadLetterQueueArn, err = sm.getQueueArn(settings.DeadLetterQueue)
		if err != nil {
			return err
		}
		redrivePolicy := &map[string]interface{}{
			"deadLetterTargetArn": deadLetterQueueArn,
			"maxReceiveCount":     MaxReceiveCount,
		}
		redrivePolicyByte, err := json.Marshal(redrivePolicy)
		if err != nil {
			return err
		}
		queueAttributesInput.Attributes["RedrivePolicy"] = aws.String(string(redrivePolicyByte))
	}

	_, err = sm.client.SetQueueAttributesWithContext(ctx, queueAttributesInput)
	if err != nil {
		return err
	}
	return
}

func (sm *SqsMediator) buildMessage(msg *sqs.Message, queueName string) (*Message, error) {
	var body string
	if body = *msg.Body; body == "" {
		return nil, nil
	}
	messageBody := &MessageBody{}
	message := &Message{msg, messageBody, nil}
	message.Remove = func(message *Message) error {
		return sm.delete(message, queueName)
	}
	if err := json.Unmarshal([]byte(body), messageBody); err != nil {
		return nil, err
	}

	return message, nil
}

func (sm *SqsMediator) delete(message *Message, queue string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	qUrl, err := sm.getQueueUrl(queue)
	if err != nil {
		return err
	}
	_, err = sm.client.DeleteMessageWithContext(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      qUrl,
		ReceiptHandle: message.ReceiptHandle,
	})
	if err != nil {
		return err
	}
	return nil
}

func (sm *SqsMediator) Subscribe(settings SubscriptionContext) error {
	qArn, err := sm.getQueueArn(settings.Queue)
	if err != nil {
		return err
	}
	sm.exchange, err = NewExchange(sm.session, sm.suffix, settings)
	if err != nil {
		return err
	}
	result, err := sm.exchange.Subscribe(qArn)
	if err != nil {
		return err
	}
	sm.Lock()
	sm.SubscriptionArn = result.SubscriptionArn
	sm.Unlock()
	return sm.setQueueAttr(settings)
}

func (sm *SqsMediator) Receive(subscriptionContext SubscriptionContext) ([]*Message, error) {
	if sm.SubscriptionArn == nil {
		if err := sm.Subscribe(subscriptionContext); err != nil {
			panic(err)
		}
	}
	queueUrl, err := sm.getQueueUrl(subscriptionContext.Queue)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()
	receiveMessageOutput, err := sm.client.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
		AttributeNames:        []*string{aws.String("SentTimestamp")},
		MaxNumberOfMessages:   aws.Int64(MaxNumberOfMessages),
		MessageAttributeNames: []*string{aws.String("All")},
		QueueUrl:              queueUrl,
		VisibilityTimeout:     aws.Int64(VisibilityTimeout),
		WaitTimeSeconds:       aws.Int64(sm.GetWaitTimeSeconds()),
	})

	if len(receiveMessageOutput.Messages) == 0 {
		return nil, nil
	}

	messages := make([]*Message, 0, len(receiveMessageOutput.Messages))
	for _, message := range receiveMessageOutput.Messages {
		enhancedMessage, err := sm.buildMessage(message, subscriptionContext.Queue)
		if err != nil {
			return nil, err
		}
		messages = append(messages, enhancedMessage)
	}

	return messages, nil
}
