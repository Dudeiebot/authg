package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Dudeiebot/dlog"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleAuth  *oauth2.Config
	oauthString = "random-string"
	logger      = dlog.NewLog(dlog.LevelTrace)
)

type Content struct {
	Id             string `json:"id"`
	Email          string `json:"email"`
	Verified_email bool   `json:"verified_email"`
}

func init() {
	// Initialize logger

	err := godotenv.Load()
	if err != nil {
		logger.Log(context.Background(), dlog.LevelFatal, "Error loading .env file")
		os.Exit(1)
	}

	googleAuth = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/callback",
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	logger.Info("OAuth configuration initialized")
}

func main() {
	http.HandleFunc("/", handleMain)
	http.HandleFunc("/login", handleGoogleLogin)
	http.HandleFunc("/callback", handleGoogleCallback)

	logger.Info("Starting server on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Log(context.Background(), dlog.LevelFatal, "", "Error", err)
		os.Exit(1)
	}
}

func handleMain(w http.ResponseWriter, r *http.Request) {
	logger.Info("Handling main request", "path", r.URL.Path)
	htmlIndex := `<html>
    <body>
    <a href="/login">Google Log In</a>
    </body>
    </html>`
	fmt.Fprintf(w, htmlIndex)
}

func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	logger.Info("Handling Google login request")
	url := googleAuth.AuthCodeURL(oauthString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	logger.Info("Handling Google callback")
	content, err := getUserInfo(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		logger.Error("Error getting user info", "Error", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	var c Content
	if err := json.Unmarshal(content, &c); err != nil {
		logger.Error("Error unmarshalling json", "Error", err)
	}
	logger.Info("User info retrieved successfully")

	var str strings.Builder
	str.WriteString("id: ")
	str.WriteString(c.Id)
	str.WriteByte(' ')
	str.WriteString("email: ")
	str.WriteString(c.Email)
	fmt.Fprintf(w, str.String())
}

func getUserInfo(state string, code string) ([]byte, error) {
	if state != oauthString {
		logger.Warn("Invalid OAuth state received")
		return nil, fmt.Errorf("invalid oauth state")
	}

	token, err := googleAuth.Exchange(context.Background(), code)
	if err != nil {
		logger.Error("Code exchange failed", "error", err)
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	logger.Info("Access token obtained", "token_type", token.TokenType)

	response, err := http.Get(
		"https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken,
	)
	if err != nil {
		logger.Error("Failed getting user info", "error", err)
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Error("Failed reading response body", "error", err)
		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}

	logger.Info("User info retrieved", "content_length", len(contents))
	return contents, nil
}
