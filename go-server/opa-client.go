package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const OPA_URL = "http://opa:8181/v1/data/main"

type OpaClient struct {
	Client *http.Client
}

type OpaInput struct {
	Jwt    string `json:"jwt"`
	Path   string `json:"path"`
	Method string `json:"method"`
}
type OpaRequest struct {
	Input OpaInput `json:"input"`
}

type OpaResponse struct {
	Result OpaResult `json:"result"`
}
type OpaResult struct {
	Allow  bool     `json:"allow"`
	Reason []string `json:"reason"`
}

func (app *OpaClient) requestOpa(ctx context.Context, payload interface{}) (*OpaResponse, error) {
	// Reuse the span
	span := trace.SpanFromContext(ctx)

	log.Print("Payload to opa:", payload)
	bytePayload, _ := json.Marshal(payload)
	request, err := http.NewRequestWithContext(ctx, "POST", OPA_URL, bytes.NewBuffer(bytePayload))
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	span.AddEvent("Querying OPA for decision")
	response, err := app.Client.Do(request)
	span.AddEvent("Request to OPA finished")
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer response.Body.Close()

	// Read the response and unmarshal it into result
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	var result OpaResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &result, nil
}

func (app *OpaClient) isAllowed(ctx context.Context, jwt, path, method string) (bool, error) {
	ctx, span := tracer.Start(ctx, "isAllowed")
	defer span.End()
	payload := OpaRequest{
		Input: OpaInput{
			Jwt:    jwt,
			Path:   path,
			Method: method,
		},
	}
	result, err := app.requestOpa(ctx, payload)
	if err != nil {
		span.RecordError(err)
		return false, err
	}
	if !result.Result.Allow {
		span.SetAttributes(
			attribute.StringSlice("OPA.policy.reason", result.Result.Reason),
			attribute.Bool("OPA.policy.allow", result.Result.Allow),
		)
		span.AddEvent(fmt.Sprintf("Allow is false with reasons %v", result.Result.Reason))
	}
	return result.Result.Allow, nil
}
