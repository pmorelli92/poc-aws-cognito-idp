package main

import (
	"log"
	"net/http"
)

func main() {
	router := http.NewServeMux()

	// Cognito will redirect here
	router.HandleFunc("/login/callback", callbackHandler())

	// The callback will redirect to /home with a token cookie
	router.HandleFunc("/home", homeHandler())

	log.Fatal(http.ListenAndServe(":8000", router))
}
