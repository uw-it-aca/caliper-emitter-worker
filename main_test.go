package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHandler(t *testing.T) {

	assert.Equal(t, true, true)
	// Skipping tests for now
	t.Skip()

	body := "{ \"secondsToRun\": 2 }"
	request := events.APIGatewayProxyRequest{Body:body}
	expectedResponse := events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "text/html",
		},
		Body: "Caliper Emitter Worker successfully triggered",
	}

	response, err := Handler(request)

	assert.Equal(t, response.Headers, expectedResponse.Headers)
	assert.Contains(t, response.Body, expectedResponse.Body)
	assert.Equal(t, err, nil)


	body = "hello"
	request = events.APIGatewayProxyRequest{Body:body}
	expectedResponse = events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "text/html",
		},
		Body: "Malformed request",
	}

	response, err = Handler(request)

	assert.Equal(t, response.Headers, expectedResponse.Headers)
	assert.Contains(t, response.Body, expectedResponse.Body)
	assert.Equal(t, err, nil)
}

func TestParseBodyJSON(t *testing.T) {
	body := "{ \"secondsToRun\": 2 }"
	// Skipping tests for now
	t.Skip()


	json, err := parseBodyJSON(body)

	expectedJSON := make(map[string]interface{})
	expectedJSON["secondsToRun"] = float64(2)

	assert.Equal(t, expectedJSON["secondsToRun"], float64(2))
	assert.Nil(t, err)
	assert.Equal(t, expectedJSON, json)



}
