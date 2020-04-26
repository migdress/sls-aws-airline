package internal

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type Enqueuer struct {
	client *sqs.SQS
}

func (e *Enqueuer) SendMsg(msg interface{}, queue string) error {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	queueURL, err := e.client.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(queue),
	})
	if err != nil {
		return err
	}

	_, err = e.client.SendMessage(&sqs.SendMessageInput{
		DelaySeconds: aws.Int64(10),
		MessageBody:  aws.String(string(msgBytes)),
		QueueUrl:     queueURL.QueueUrl,
	})
	if err != nil {
		return err
	}

	return nil
}

func NewEnqueuer(client *sqs.SQS) *Enqueuer {
	return &Enqueuer{
		client: client,
	}
}
