package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	cv "github.com/nirasan/go-oauth-pkce-code-verifier"
	"golang.org/x/oauth2"
)

type OauthTokenHandler struct {
	ctx          context.Context
	config       *oauth2.Config
	tokFile      string             // File path to cache the token
	t            chan *oauth2.Token // to feed the token from the redirect uri
	codeVerifier *cv.CodeVerifier
}

type IdToken struct {
	IdToken string    `json:"id_token"`
	Expiry  time.Time `json:"expiry,omitempty"`
}

//func NewOauthTokenHandler(authUrl, tokenUrl, clientID, clientSecret string) *OauthTokenHandler {
func NewOauthTokenHandler(authUrl, tokenUrl, clientID, tokFile string) *OauthTokenHandler {
	oa := &OauthTokenHandler{
		ctx: context.Background(),
		config: &oauth2.Config{
			ClientID: clientID,
			//ClientSecret: clientSecret,
			Scopes:      []string{"openid", "email"},
			RedirectURL: "http://localhost:7773/oauth2/callback",
			Endpoint: oauth2.Endpoint{
				AuthURL:  authUrl,
				TokenURL: tokenUrl,
			},
		},
		t:       make(chan *oauth2.Token),
		tokFile: tokFile,
	}

	return oa
}

// prepareTokenWebChalenge: returns a url to follow oauth
func (oa *OauthTokenHandler) prepareTokenWebChalenge() (string, error) {
	codeVerifier, err := cv.CreateCodeVerifier()
	if err != nil {
		return "", fmt.Errorf("Cannot create a code verifier: %v", err)
	}
	oa.codeVerifier = codeVerifier
	codeChallenge := oa.codeVerifier.CodeChallengeS256()
	codeChallengeOpt := oauth2.SetAuthURLParam("code_challenge", codeChallenge)
	codeChallengeMethodOpt := oauth2.SetAuthURLParam("code_challenge_method", "S256")

	url := oa.config.AuthCodeURL(
		"state-token",
		oauth2.AccessTypeOnline,
		codeChallengeOpt,
		codeChallengeMethodOpt,
	)
	return url, nil
}

func (oa *OauthTokenHandler) getTokenFromFile() (*IdToken, error) {
	f, err := os.Open(oa.tokFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &IdToken{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func (oa *OauthTokenHandler) saveToken(token *IdToken) {
	fmt.Printf("Saving credential file to: %s\n", oa.tokFile)
	f, err := os.OpenFile(oa.tokFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func (oa *OauthTokenHandler) ExchangeToken(code string) (*IdToken, error) {
	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token.
	codeVerifierOpt := oauth2.SetAuthURLParam("code_verifier", oa.codeVerifier.String())
	tok, err := oa.config.Exchange(oa.ctx, code, codeVerifierOpt)
	if err != nil {
		log.Fatal(err)
	}

	rawIDToken, ok := tok.Extra("id_token").(string)
	if !ok {
		return &IdToken{}, fmt.Errorf("Cannot get id_token data from token")
	}

	id_token := &IdToken{
		IdToken: rawIDToken,
		Expiry:  tok.Expiry,
	}

	oa.saveToken(id_token)

	return id_token, nil

}
