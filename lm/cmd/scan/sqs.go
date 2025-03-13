package scan

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"

	sqsh "github.com/organicveggie/livemusic/lm/aws/sqs"
)

type QueueOut struct {
	queueName string
	queueURL  string

	session *session.Session
	sqsSvc  *sqs.SQS
}

func newQueueOut(profile, queueName string) (*QueueOut, error) {
	q := &QueueOut{
		queueName: queueName,
	}

	profileName := profile
	if profile == "" {
		profileName = "default"
	}

	// Create the AWS session
	var err error
	q.session, err = session.NewSessionWithOptions(session.Options{
		// Specify profile to load for the session's config
		Profile: profileName,

		// Provide SDK Config options, such as Region.
		Config: aws.Config{
			Region: aws.String("us-west-2"),
		},

		// Force enable Shared Config support
		SharedConfigState: session.SharedConfigEnable,
	})

	if err != nil {
		return nil, fmt.Errorf("error creating AWS session: %v", err)
	}
	if _, err = q.session.Config.Credentials.Get(); err != nil {
		return nil, fmt.Errorf("error loading AWS credentials: %v", err)
	}

	// Create the SQS client
	q.sqsSvc = sqs.New(q.session)

	// Retrieve the SQS queue URL
	if q.queueURL, err = sqsh.GetQueueURL(q.sqsSvc, q.queueName); err != nil {
		return nil, err
	}

	return q, nil
}

func (q *QueueOut) Close() error {
	return nil
}

func (q *QueueOut) AddFile(filename string) error {
	msg := sqs.SendMessageInput{
		MessageBody: aws.String(filename),
		QueueUrl:    &q.queueURL,
	}
	sendOutput, err := q.sqsSvc.SendMessage(&msg)
	if err != nil {
		return fmt.Errorf("error sending queue message for %q to %s: %v", filename, q.queueName, err)
	}

	fmt.Printf("[%s]: %q\n", *sendOutput.MessageId, filename)
	return nil
}
