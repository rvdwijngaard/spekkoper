package middleware

import (
	"context"
	"fmt"
	"sync"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"github.com/authorizerdev/authorizer-go"
)

var secrets struct {
	AuthorizerClientID    string
	AuthorizerURL         string
	AuthorizerRedirectURL string
}

// AuthHandler can be named whatever you prefer (but must be exported).
//
//encore:authhandler
func AuthHandler(ctx context.Context, token string) (auth.UID, error) {
	var authorizerClient *authorizer.AuthorizerClient
	init := sync.Once{}
	init.Do(func() {
		defaultHeaders := map[string]string{}
		var err error
		authorizerClient, err = authorizer.NewAuthorizerClient(secrets.AuthorizerClientID, secrets.AuthorizerURL, secrets.AuthorizerRedirectURL, defaultHeaders)
		if err != nil {
			panic(err)
		}
	})
	// Validate the token and look up the user id and user data,
	res, err := authorizerClient.ValidateJWTToken(&authorizer.ValidateJWTTokenInput{Token: token, TokenType: authorizer.TokenTypeIDToken})
	if err != nil {
		rlog.Error("could not validate jwt token", "err", err)
		return "", &errs.Error{Code: errs.Unauthenticated, Message: "could not validate the token"}
	}
	if !res.IsValid {
		return "", &errs.Error{Code: errs.Unauthenticated, Message: "token not valid"}
	}

	profile, err := authorizerClient.GetProfile(map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	})
	if err != nil {
		return "", &errs.Error{Code: errs.Internal, Message: "could not get user profile"}
	}
	return auth.UID(profile.ID), nil
}
