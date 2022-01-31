package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/square/go-jose/v3/jwt"
	"go.opentelemetry.io/otel/attribute"
)

func extractBearerToken(ctx context.Context, header http.Header) (string, error) {
	_, span := tracer.Start(ctx, "OpaAuthorizerMiddleware")
	defer span.End()

	authHeader := header.Get("Authorization")
	components := strings.Fields(authHeader)
	if len(components) != 2 {
		err := errors.New("invalid authorization header")
		span.RecordError(err)
		return "", err
	}

	if components[0] != "Bearer" {
		err := fmt.Errorf("authorization header is not of 'Bearer' type. Found %s", components[0])
		span.RecordError(err)
		return "", err
	}

	_, err := jwt.ParseSigned(components[1])
	if err != nil {
		span.RecordError(err)
		return "", err
	}
	span.AddEvent("Found well-formed JWT bearer token in request Authorization header")
	// We will validate signature in OPA
	return components[1], nil
}

func OpaAuthorizerMiddleware(opaClient OpaClient) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Save ctx for resetting before going to next handler in chain
		savedCtx := c.Request.Context()
		defer func() {
			c.Request = c.Request.WithContext(savedCtx)
		}()

		ctx, span := tracer.Start(c.Request.Context(), "OpaAuthorizerMiddleware")
		defer span.End()
		// Set the context of the request to the current spans context, so that it is used throughout the logic of the middleware itself
		c.Request = c.Request.WithContext(ctx)

		path := c.Request.URL.Path
		jwToken, err := extractBearerToken(ctx, c.Request.Header)
		if err != nil {
			span.RecordError(err)
			// Continue -- If jwt is required, OPA will catch it
		}

		span.SetAttributes(attribute.String("Jwt", jwToken))

		allowed, err := opaClient.isAllowed(ctx, jwToken, path, c.Request.Method)
		if err != nil {
			span.RecordError(err)
			c.Status(http.StatusInternalServerError)
			return
		}

		if !allowed {
			span.AddEvent("Request is forbidden")
			c.AbortWithStatus(http.StatusForbidden)
			return
		} else {
			span.AddEvent("Request is allowed")
		}
		// We end the span before calling next because the middleware task is completed
		// span.End()
	}
}
