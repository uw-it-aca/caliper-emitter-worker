package main

import (
	"io/ioutil"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"encoding/json"
	"errors"
	"time"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"strconv"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Handler is executed by AWS Lambda in the main function. Once the request
// is processed, it returns an Amazon API Gateway response object to AWS Lambda
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	body := request.Body

	err := emitEvents(body)

	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Malformed request",
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Caliper Emitter Worker successfully triggered",
		Headers: map[string]string{
			"Content-Type": "text/html",
		},
	}, nil

}

func main() {
	// Respond to the API Gateway Request
	lambda.Start(Handler)
}

func emitEvents(body string)(error){
	json, err := parseBodyJSON(body)

	if err != nil {
		return err
	}

	secondsToRun := json["secondsToRun"].(float64)
	seconds := time.Duration(secondsToRun) * time.Second

	start := time.Now()

	svc := getQueueConfig()
	payload := getPayload()
	msg_input := getMessageInput(payload)

	for elapsed := time.Since(start); elapsed < seconds; elapsed = time.Since(start) {
		_, err := svc.SendMessageBatch(msg_input)
		check(err)
	}

	return nil
}

func getMessageInput(payload string) *sqs.SendMessageBatchInput{
	qURL := getQueueURL()

	entries := make([]*sqs.SendMessageBatchRequestEntry, 0)

	for i := 0; i < 10; i++ {
		id := strconv.FormatInt(int64(i), 10)
		entry := &sqs.SendMessageBatchRequestEntry{
			MessageBody: aws.String(payload),
			Id:          &id,
		}
		entries = append(entries, entry)
	}

	return &sqs.SendMessageBatchInput{
		Entries:  entries,
		QueueUrl: &qURL,
	}
}

func getPayload()(string){
	dat, err := ioutil.ReadFile("./generated.json")

	check(err)

	return string(dat)
}

func getQueueConfig() *sqs.SQS{
	config := aws.Config{
		Region: aws.String("us-west-2"),
	}
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            config,
	}))

	svc := sqs.New(sess)
	return svc

}

func getQueueURL()(string){
	// URL to our queue
	return os.Getenv("SQS_URL")
}

type LambdaConfig struct {
	secondsToRun int
}
/*
The body of the request must follow the following format or it will be rejected
{
    "secondsToRun": {int}
}
 */
func parseBodyJSON(body string)(map[string]interface{}, error){
	var dat map[string]interface{}

	err := json.Unmarshal([]byte(body), &dat)

	if _, ok := dat["secondsToRun"]; !ok{
		return dat,errors.New("required parameter \"secondsToRun\" not found")
	}

	return dat, err
}