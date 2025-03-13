package analyze

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"

	sqsh "github.com/organicveggie/livemusic/lm/aws/sqs"
)

type SQSSource struct {
	ch              AnalyzeChan
	maxMessages     int
	queueName       string
	waitTimeoutSecs int

	client   *sqs.SQS
	queueURL string
	session  *session.Session
}

func newSQSSource(ch AnalyzeChan, profile, queueName string) (*SQSSource, error) {
	q := &SQSSource{
		ch:              ch,
		maxMessages:     10,
		queueName:       queueName,
		waitTimeoutSecs: 15,
	}

	// Setup AWS session options
	options := session.Options{
		// Provide SDK Config options, such as Region.
		Config: aws.Config{
			Region: aws.String("us-west-2"),
		},
		// Force enable Shared Config support
		SharedConfigState: session.SharedConfigEnable,
	}
	if profile != "" {
		// Specify profile to load for the session's config
		options.Profile = profile
	}

	// Create the AWS session
	var err error
	q.session, err = session.NewSessionWithOptions(options)
	if err != nil {
		return nil, fmt.Errorf("error creating AWS session: %v", err)
	}
	if _, err = q.session.Config.Credentials.Get(); err != nil {
		return nil, fmt.Errorf("error loading AWS credentials: %v", err)
	}

	// Create the SQS client
	q.client = sqs.New(q.session)

	// Retrieve the SQS queue URL
	if q.queueURL, err = sqsh.GetQueueURL(q.client, q.queueName); err != nil {
		return nil, err
	}

	return q, nil
}

func (sq *SQSSource) Close() error {
	return nil
}

func (sq *SQSSource) Analyze() error {
	hasMoreMessages := true
	for hasMoreMessages {
		msgIn := sqs.ReceiveMessageInput{
			MaxNumberOfMessages: aws.Int64(int64(sq.maxMessages)),
			MessageAttributeNames: []*string{
				aws.String(sqs.QueueAttributeNameAll),
			},
			QueueUrl:        &sq.queueURL,
			WaitTimeSeconds: aws.Int64(int64(sq.waitTimeoutSecs)),
		}

		recvMsg, err := sq.client.ReceiveMessage(&msgIn)
		if err != nil {
			return fmt.Errorf("error reading SQS messages from %s: %v", sq.queueURL, err)
		}

		for _, msg := range recvMsg.Messages {
			fmt.Printf("Received [%s] %q\n", *msg.MessageId, *msg.Body)
			sq.ch <- *msg.Body

			if _, err := sq.client.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      &sq.queueURL,
				ReceiptHandle: msg.ReceiptHandle,
			}); err != nil {
				return fmt.Errorf("error deleting [%s] from %s: %v", *msg.ReceiptHandle, sq.queueURL, err)
			}
		}

		hasMoreMessages = len(recvMsg.Messages) > 0
	}

	return nil
}
