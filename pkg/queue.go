package misaki

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type SQSQueue struct {
	client *sqs.SQS
	queue_url string
}

type SQSMessage struct {
	Content string
	receipt string
	q *SQSQueue
}

func NewSQSQueue(queue_name, region, key_id, access_key string) (*SQSQueue, error) {

	aws_config := aws.NewConfig()
	if region != "" {
		aws_config.Region = aws.String(region)
	}
	if key_id != "" || access_key != "" {
		creds := credentials.NewStaticCredentials(key_id, access_key, "")
		aws_config.Credentials = creds
	}

	session, err := session.NewSession(aws_config)
	if err != nil {
		return nil, err
	}

	client := sqs.New(session)

	params := sqs.GetQueueUrlInput{
		QueueName: aws.String(queue_name),
	}
	resp, err := client.GetQueueUrl(&params)

	if err != nil {
		return nil, err
	}

	q := SQSQueue {
		client: client,
		queue_url: *resp.QueueUrl,
	}

	return &q, nil
}

func (q *SQSQueue) WaitMessage() (*SQSMessage, error) {

	params := sqs.ReceiveMessageInput {
		QueueUrl: aws.String(q.queue_url),
		MaxNumberOfMessages: aws.Int64(1),
		WaitTimeSeconds: aws.Int64(20),
	}

	for {
		resp, err := q.client.ReceiveMessage(&params)
		if err != nil {
			return nil, err
		}
		if len(resp.Messages) == 0 {
			continue
		}
		resp_msg := resp.Messages[0]
		body := resp_msg.Body
		if body == nil || resp_msg.ReceiptHandle == nil {
			return nil, fmt.Errorf("empty SQS message")
		}
		m := SQSMessage{
			Content: *body,
			receipt: *resp_msg.ReceiptHandle,
			q: q,
		}
		return &m, nil
	}
}

func get_label() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic("rand.Read failed")
	}
	return hex.EncodeToString(bytes)
}

func (q *SQSQueue) PostMessage(content string) error {
	params := sqs.SendMessageInput {
		MessageBody: aws.String(content),
		QueueUrl: aws.String(q.queue_url),
		MessageGroupId: aws.String(get_label()),
		MessageDeduplicationId: aws.String(get_label()),
	}
	_, err := q.client.SendMessage(&params)
	return err
}

func (m *SQSMessage) Delete() error {
	params := sqs.DeleteMessageInput {
		QueueUrl: aws.String(m.q.queue_url),
		ReceiptHandle: &m.receipt,
	}
	_, err := m.q.client.DeleteMessage(&params)
	return err
}
