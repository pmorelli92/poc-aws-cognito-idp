package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func homeHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		c, err := r.Cookie("token")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		keyID, err := getKeyIDFromToken(c.Value)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// This should be cached
		validKeys, err := getKeysIDFromJWKS()
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		validSignature := false
		for _, vk := range validKeys {
			if vk == keyID {
				validSignature = true
				break
			}
		}

		if !validSignature {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		_, _ = w.Write([]byte("VALID TOKEN, PRINTING FOR DEBUG: " + c.Value))
	}
}

func getKeysIDFromJWKS() ([]string, error) {
	r, err := url.Parse(os.Getenv("COGNITO_JKWS_URL"))
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(r.String())
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS response with status %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	rsBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type keysRs struct {
		Keys []struct {
			KeyID string `json:"kid"`
		} `json:"keys"`
	}

	var keys keysRs
	if err := json.Unmarshal(rsBytes, &keys); err != nil {
		return nil, err
	}

	k := make([]string, len(keys.Keys))
	for i, v := range keys.Keys {
		k[i] = v.KeyID
	}

	return k, nil
}

func getKeyIDFromToken(token string) (string, error) {
	headerStr := strings.Split(token, ".")
	if len(headerStr) == 0 {
		return "", fmt.Errorf("the token has an invalid format")
	}

	b, err := base64.RawStdEncoding.DecodeString(headerStr[0])
	if err != nil {
		return "", err
	}

	type headerJWT struct {
		KeyID string `json:"kid"`
	}

	var header headerJWT
	if err := json.Unmarshal(b, &header); err != nil {
		return "", err
	}

	return header.KeyID, nil
}
