package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/xeipuuv/gojsonschema"
)

func EmptyResponse(statusCode int) events.APIGatewayProxyResponse {
	r := events.APIGatewayProxyResponse{
		StatusCode:      statusCode,
		IsBase64Encoded: false,
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
	}
	return r
}

func JSONStruct(statusCode int, structure interface{}) events.APIGatewayProxyResponse {
	data, err := json.Marshal(structure)
	if err != nil {
		return JSONError(http.StatusInternalServerError, err)
	}
	return JSONResponse(statusCode, data)
}

func JSONResponse(statusCode int, body []byte) events.APIGatewayProxyResponse {
	buf := bytes.Buffer{}
	json.HTMLEscape(&buf, body)
	r := events.APIGatewayProxyResponse{
		StatusCode:      statusCode,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
	return r
}

func JSONError(statusCode int, err error) events.APIGatewayProxyResponse {
	body, _ := json.Marshal(map[string]interface{}{
		"error": err.Error(),
	})
	return JSONResponse(statusCode, body)
}

func JSONSchemaError(statusCode int, schemaErrors []gojsonschema.ResultError) events.APIGatewayProxyResponse {
	errors := []string{}
	for _, error := range schemaErrors {
		errString := fmt.Sprintf("%v", error)
		errors = append(errors, errString)
	}
	body, _ := json.Marshal(map[string]interface{}{
		"errors": errors,
	})
	return JSONResponse(statusCode, body)
}
