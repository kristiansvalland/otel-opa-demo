package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

const (
	SERVICE_NAME = "go-server"
)

type App struct {
	UserHandler UserHandler
}

func main() {
	log.Printf("Starting app...")
	ctx := context.Background()
	shutdown := initTracer(ctx)
	defer shutdown()
	log.Printf("Got tracer")
	client := otelhttp.DefaultClient
	opaClient := OpaClient{Client: client}
	userHandler := UserHandler{OpaClient: opaClient}
	app := App{UserHandler: userHandler}

	tracer = otel.Tracer("go-server")

	r := gin.New()
	r.Use(otelgin.Middleware("my-server"))
	r.Use(OpaAuthorizerMiddleware(opaClient))
	r.GET("/users/:id", app.UserHandler.GetUser)
	log.Printf("Starting to listen")
	_ = r.Run(":8080")
}
