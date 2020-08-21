package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/garyburd/go-oauth/oauth"
	"net/http"
	"net/url"
	"os"
)

type Account struct {
	ID              string `json:"id_str"`
	ScreenName      string `json:"screen_name"`
	ProfileImageURL string `json:"profile_image_url"`
	Email           string `json:"email"`
	Lang            string `json:"lang"`
}

var (
	CONSUMER_KEY    = os.Getenv("CONSUMER_KEY")
	CONSUMER_SECRET = os.Getenv("CONSUMER_SECRET")
	test            = "http://127.0.0.1:8080/callback"
)

func main() {

	http.HandleFunc("/auth", AuthTwitter)
	http.HandleFunc("/callback", Callback)
	http.HandleFunc("/redirect", Redirect)
	http.ListenAndServe(":8080", nil)

}

func Redirect(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "redirect success")
}

func connectAPI() *anaconda.TwitterApi {
	anaconda.SetConsumerKey(CONSUMER_KEY)
	anaconda.SetConsumerSecret(CONSUMER_SECRET)

	return anaconda.NewTwitterApi("", "")
}

func AuthTwitter(w http.ResponseWriter, req *http.Request) {
	api := connectAPI()

	uri, _, error := api.AuthorizationURL(test)
	if error != nil {
		fmt.Println(error)
	}

	http.Redirect(w, req, uri, http.StatusFound)
}

func Callback(w http.ResponseWriter, req *http.Request) {

	token := req.URL.Query().Get("oauth_token")
	secret := req.URL.Query().Get("oauth_verifier")
	api := connectAPI()

	cred, _, error := api.GetCredentials(&oauth.Credentials{
		Token: token,
	}, secret)
	if error != nil {
		fmt.Println(error)
	}

	userApi := anaconda.NewTwitterApi(cred.Token, cred.Secret)
	account, err := GetAccount(userApi.Credentials, CONSUMER_KEY, CONSUMER_SECRET)
	if err != nil {
		fmt.Println(err)
		return
	}
	email := account.Email

	v := url.Values{}
	user, _ := userApi.GetUsersShow(account.ScreenName, v)

	fmt.Println(user)
	fmt.Println(email)

	http.Redirect(w, req, "http://127.0.0.1:8080/redirect", http.StatusFound)
}

func GetConnect(tempCredKey string, tokenCredKey string) *oauth.Client {
	return &oauth.Client{
		TemporaryCredentialRequestURI: "https://api.twitter.com/oauth/request_token",
		ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authorize",
		TokenRequestURI:               "https://api.twitter.com/oauth/access_token",
		Credentials: oauth.Credentials{
			Token:  tempCredKey,
			Secret: tokenCredKey,
		},
	}
}

func GetAccount(at *oauth.Credentials, consumer_key string, consumer_secret string) (*Account, error) {
	oc := GetConnect(consumer_key, consumer_secret)

	v := url.Values{}
	v.Set("include_email", "true")
	v.Encode()

	resp, err := oc.Get(nil, at, "https://api.twitter.com/1.1/account/verify_credentials.json", v)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return nil, errors.New("Twitter is unavailable")
	}
	if resp.StatusCode >= 400 {
		return nil, errors.New("Twitter request is invalid")
	}
	user := &Account{}
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

