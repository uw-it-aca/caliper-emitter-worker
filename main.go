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
			Body:       "Error: " + err.Error(),
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
	requestJSON, err := parseBodyJSON(body)

	if err != nil {
		return err
	}

	secondsToRun := requestJSON["secondsToRun"].(float64)

	for i := 0; i < getNumThreads(); i++{
		go emit(secondsToRun)
	}

	emit(secondsToRun)

	return nil
}


func emit(secondsToRun float64){
	seconds := time.Duration(secondsToRun) * time.Second

	svc := getQueueConfig()
	payload := getPayload()
	msgInput := getMessageInput(payload)

	start := time.Now()
	for elapsed := time.Since(start); elapsed < seconds; elapsed = time.Since(start) {
		_, err := svc.SendMessageBatch(msgInput)
		check(err)
	}

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
	dat, err := ioutil.ReadFile("./public/generated.json")

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

// returns the URL to our queue
func getQueueURL()(string){
	return getEnvVar("SQS_URL")
}

/*
LambdaConfig is the type of the caliper-emitter-worker's configuration
 */
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

func getNumThreads()(int){
	num, err := strconv.Atoi(getEnvVar("NUM_THREADS"))

	if err != nil {
		panic(err)
	}

	return num
}

/*
Returns the value of the given environment variable key, and raises an error if it is not present
 */
func getEnvVar(key string)(string){
	url := os.Getenv(key)

	if url == "" {
		panic(errors.New("no value was found for the environement variable of : " + key))
	}

	return url
}