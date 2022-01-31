package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
)

type UserHandler struct {
	OpaClient OpaClient
}

func getUsername(ctx context.Context, idString string) (string, error) {
	_, span := tracer.Start(ctx, "getUsername()")
	defer span.End()

	var user string
	switch idString {
	case "1":
		user = "alice"
	case "2":
		user = "mallory"
	default:
		return "", errors.New("user not found")
	}
	span.AddEvent(fmt.Sprintf("Found user: %v", user))
	return user, nil
}

// GetUser
func (app *UserHandler) GetUser(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "getUser()")
	defer span.End()

	id := c.Param("id")
	span.SetAttributes(attribute.String("userId", id))
	username, err := getUsername(ctx, id)
	if err != nil {
		span.RecordError(err)
		c.AbortWithError(404, err)
		return
	}
	span.SetAttributes(attribute.String("userId", id))

	c.JSON(200, gin.H{
		"username": username,
	})
}
