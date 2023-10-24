package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func callbackHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		token, err := getToken(r.Context(), code)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		cookie := http.Cookie{
			Name:     "token",
			Value:    token,
			MaxAge:   3600,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		}

		http.SetCookie(w, &cookie)
		homeURL := fmt.Sprintf("%s/home", os.Getenv("API_URL"))
		http.Redirect(w, r, homeURL, http.StatusSeeOther)
	}
}

func getToken(ctx context.Context, code string) (string, error) {
	OAuthURL := os.Getenv("COGNITO_OAUTH_URL")
	redirectURL := fmt.Sprintf("%s/login/callback", os.Getenv("API_URL"))

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURL)
	d := data.Encode()

	clientID := os.Getenv("COGNITO_USER_POOL_CLIENT_ID")
	clientSecret := os.Getenv("COGNITO_USER_POOL_CLIENT_SECRET_ID")
	basicAuth := fmt.Sprintf("%s:%s", clientID, clientSecret)
	sEnc := base64.StdEncoding.EncodeToString([]byte(basicAuth))

	r, err := http.NewRequestWithContext(
		ctx, http.MethodPost, OAuthURL, strings.NewReader(d))
	if err != nil {
		return "", err
	}

	r.Header.Add("Authorization", fmt.Sprintf("Basic %s", sEnc))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	rs, err := http.DefaultClient.Do(r)
	if err != nil {
		return "", err
	}

	defer rs.Body.Close()
	rsBytes, err := io.ReadAll(rs.Body)
	if err != nil {
		return "", err
	}

	if rs.StatusCode != http.StatusOK {
		return "", fmt.Errorf("response code is %d", rs.StatusCode)
	}

	var token tokenRs
	if err := json.Unmarshal(rsBytes, &token); err != nil {
		return "", err
	}

	return token.IDToken, nil
}

// Other properties abbreviated, do not need for this use case
type tokenRs struct {
	IDToken string `json:"id_token"`
}
