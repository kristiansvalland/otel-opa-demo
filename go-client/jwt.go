package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"log"

	"github.com/square/go-jose/v3"
	"github.com/square/go-jose/v3/jwt"
)

type JwtService struct {
	signer *jose.Signer
}

func NewJwtService() *JwtService {
	signer, err := getSigner()
	if err != nil {
		log.Fatal("error creating signer", err)
	}
	return &JwtService{
		signer: signer,
	}
}

func getSigner() (*jose.Signer, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	sig := jose.SigningKey{
		Algorithm: jose.RS256,
		Key:       key,
	}
	opts := &jose.SignerOptions{}
	signer, err := jose.NewSigner(sig, opts)
	if err != nil {
		return nil, err
	}
	return &signer, nil
}
func (j *JwtService) getToken(ctx context.Context, user User) (string, error) {
	_, span := tracer.Start(ctx, "getToken()")
	defer span.End()

	builder := jwt.Signed(*j.signer)
	jwt, err := builder.Claims(user).CompactSerialize()
	return jwt, err
}
