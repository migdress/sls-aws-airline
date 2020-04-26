package internal

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/xeipuuv/gojsonschema"
)

func Respond(statusCode int, body string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: body,
	}
}

func Error(statusCode int, err error) events.APIGatewayProxyResponse {
	responseBytes, _ := json.Marshal(map[string]interface{}{
		"errors": []string{err.Error()},
	})

	return Respond(statusCode, string(responseBytes))
}

func SchemaErrors(statusCode int, schemaErrors []gojsonschema.ResultError) events.APIGatewayProxyResponse {
	errors := []string{}

	for _, error := range schemaErrors {
		errString := fmt.Sprintf("%v", error)
		errors = append(errors, errString)
	}

	body, _ := json.Marshal(map[string]interface{}{
		"errors": errors,
	})

	return Respond(statusCode, string(body))
}
