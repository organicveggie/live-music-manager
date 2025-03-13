package sqs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/sqs"
)

// GetQueueURL retrieves the SQS Queue URL for a given queue name.
func GetQueueURL(sqsSvc *sqs.SQS, queueName string) (string, error) {
	url, err := sqsSvc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})
	if err != nil {
		return "", fmt.Errorf("error retrieving queue URL for queue %q: %v", queueName, err)
	}
	return *url.QueueUrl, nil
}
