package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	oauthString = "random-string"
	googleAuth  *oauth2.Config
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	googleAuth = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/callback",
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
}

func main() {
	http.HandleFunc("/", handleMain)
	http.HandleFunc("/login", handleGoogleLogin)
	http.HandleFunc("/callback", handleGoogleCallback)
	http.ListenAndServe(":8080", nil)
}

func handleMain(w http.ResponseWriter, r *http.Request) {
	htmlIndex := `<html>
    <body>
    <a href ="/login">Google Log IN</a>
    </body>
    </html>`
	fmt.Fprintf(w, htmlIndex)
}

func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleAuth.AuthCodeURL(oauthString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	content, err := getUserInfo(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	fmt.Fprintf(w, "Contet: %s\n", content)
}

func getUserInfo(state string, code string) ([]byte, error) {
	if state != oauthString {
		return nil, fmt.Errorf("invalid oauth state")
	}

	token, err := googleAuth.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	response, err := http.Get(
		"https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken,
	)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}

	defer response.Body.Close()
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}
	return contents, nil
}
