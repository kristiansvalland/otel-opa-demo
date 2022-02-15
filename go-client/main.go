package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	SERVER_SCHEME = "http"
	SERVER_HOST   = "go-server:8080"
	SERVICE_NAME  = "go-client"
)

var tracer trace.Tracer

type User struct {
	Username string   `json:"username"`
	UserID   int      `json:"userid"`
	Roles    []string `json:"roles"`
}

var users map[int]User = map[int]User{
	1: {Username: "alice", Roles: []string{"admin"}, UserID: 1},
	2: {Username: "bob", Roles: []string{"user"}, UserID: 2},
	3: {Username: "mallory", Roles: []string{"user"}, UserID: 3},
	4: {Username: "unknown", Roles: []string{"user"}, UserID: 4},
}

var c int = 0

func getRandomUser() *User {
	id := c%5 + 1
	c++
	// id := rand.Intn(4) + 1
	if user, ok := users[id]; ok {
		return &user
	}
	return nil
}

func getUrl(user *User) string {
	var id = 1
	if user != nil {
		id = user.UserID
	}
	return fmt.Sprintf("http://go-server:8080/users/%v", id)
}

// createRequest sets up a request that can be used with http.Client.Do.
// It is not attached to a context, so it needs to used with http.Request.WithContext(ctx)
func (client *Client) createRequest(ctx context.Context) (*http.Request, error) {
	ctx, span := tracer.Start(ctx, "getRequest()")
	defer span.End()

	user := getRandomUser()
	request, err := http.NewRequest("GET", getUrl(user), nil)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	if user != nil {
		jwt, err := client.jwt.getToken(ctx, *user)
		if err != nil {
			log.Fatalf("Error creating token with user=%v: %v", user, err)
		}
		log.Print("JWT: ", jwt)
		request.Header.Set("Authorization", "Bearer "+jwt)
	}
	return request, nil

}

// doRequest makes a random request to the backend
func (client *Client) doRequest(ctx context.Context) {
	spanctx, span := tracer.Start(ctx, "getUser")
	defer span.End()
	request, err := client.createRequest(spanctx)
	if err != nil {
		span.RecordError(err)
		return
	}
	request = request.WithContext(spanctx)

	hc := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	resp, err := hc.Do(request)
	if err != nil {
		span.RecordError(err)
		return
	}

	attrs := semconv.HTTPAttributesFromHTTPStatusCode(resp.StatusCode)
	spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCode(resp.StatusCode)
	span.SetAttributes(attrs...)
	span.SetStatus(spanStatus, spanMessage)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
	}

	log.Print("Response:", resp, "body:", string(body))
}

type Client struct {
	jwt        *JwtService
	httpClient *http.Client
}

func main() {
	ctx := context.Background()
	shutdown := initTracer(ctx)
	defer shutdown()
	tracer = otel.Tracer("go-client")
	log.Printf("Finished setting up tracer")
	client := Client{
		jwt:        NewJwtService(),
		httpClient: otelhttp.DefaultClient,
	}

	c := 5

	for i := 0; i < c; i++ {
		log.Printf("Running %v of %v requests", i+1, c)
		client.doRequest(ctx)
		time.Sleep(5 * time.Second)
	}
}
