package main

import (
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

var (
	endpoint = "https://message-queue.api.cloud.yandex.net"
	region   = "ru-central1"
	qName    = "snapshot-tasks"
)

func createQueue() (string, string) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Endpoint: &endpoint,
			Region:   &region,
		},
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := sqs.New(sess)

	res, _ := svc.CreateQueue(&sqs.CreateQueueInput{
		QueueName: &qName,
	})

	all := "All"

	attrs, _ := svc.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		AttributeNames: []*string{&all},
		QueueUrl:       res.QueueUrl,
	})
	queueArn := attrs.Attributes["QueueArn"]
	return *res.QueueUrl, *queueArn
}

func set(name, value, data string) string {
	newRecord := name + "=" + value + "\n"
	r := regexp.MustCompile(name + "=.*\n")
	if r.MatchString(data) {
		data = r.ReplaceAllString(data, newRecord)
	} else {
		data = data + "\n" + newRecord
	}
	return data
}

func main() {
	data, err := ioutil.ReadFile(".env")
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}
	dataStr := string(data)
	queueUrl, queueArn := createQueue()
	dataStr = set("QUEUE_URL", queueUrl, dataStr)
	dataStr = set("QUEUE_ARN", queueArn, dataStr)

	_ = ioutil.WriteFile(".env", []byte(dataStr), 0777)

}
