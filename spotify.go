package main

import (
	"golang.org/x/oauth2"
	"github.com/zmb3/spotify"
	"os"
	"io/ioutil"
	"fmt"
	"github.com/ahl5esoft/golang-underscore"
	"net/http"
	"net/url"
)

type Config struct {
	redirectUri string
	clientID string
	clientSecret string
	scopes []string
	state string
}

var (
	config Config
	auth spotify.Authenticator
	tokens []oauth2.Token
	Clients []spotify.Client
)

func buildRedirectUri() string {
	uri := url.URL{
		Scheme: "http",
		Host: os.Getenv("HTTP_HOST") + ":" + os.Getenv("HTTP_PORT"),
		Path: os.Getenv("SPOTIFY_REDIRECT_URI"),
	}

	return uri.String()
}

func buildSpotifyConfig() Config {
	return Config {
		clientID: os.Getenv("SPOTIFY_CLIENT_ID"),
		clientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		redirectUri: buildRedirectUri(),
		scopes: []string{"user-modify-playback-state", "user-read-currently-playing", "user-read-recently-played"},
		state: "nostate",
	}
}


func loadStoredTokens() {
	filePath := os.Getenv("TOKEN_FILE")

	content, err := ioutil.ReadFile(filePath)

	if err != nil {
	fmt.Println(err)
	}

	underscore.ParseJson(string(content), &tokens)
}

func StoreTokens() {
	json, err := underscore.ToJson(tokens)

	if err != nil {
		fmt.Println(err)
		return
	}

	ioutil.WriteFile(os.Getenv("TOKEN_FILE"), []byte(json), os.FileMode(0777))
}


func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.Token(config.state, r)
	w.Write([]byte("<script>window.close();</script>"))

	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusNotFound)
		return
	}

	AuthoriseClient(token)
	tokens = append(tokens, *token)
}

func GetAuthoriseUrl() string {
	return auth.AuthURL(config.state)
}

func authoriseStoredTokens() {
	for _, token := range tokens {
		AuthoriseClient(&token)
	}
}

func AuthoriseClient(token *oauth2.Token) {
	client := auth.NewClient(token)

	user, err := client.CurrentUser()

	if err == nil {
		fmt.Println(err)
	}

	var name string

	if user.DisplayName == "" {
		name = user.ID
	} else {
		name = user.DisplayName
	}

	println(name + " is now listening")

	Clients = append(Clients, client)
}

func init() {
	config = buildSpotifyConfig()

	auth = spotify.NewAuthenticator(config.redirectUri, config.scopes...)
	auth.SetAuthInfo(config.clientID, config.clientSecret)

	http.HandleFunc(os.Getenv("SPOTIFY_REDIRECT_URI"), RedirectHandler)

	loadStoredTokens()
	authoriseStoredTokens()
}