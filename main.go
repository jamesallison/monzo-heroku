package main

import (
	"log"
	"net/http"
	"os"
	"fmt"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	http.HandleFunc("/webhook", webhookHandler)

	http.ListenAndServe(":" + port, nil)
}

func webhookHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello from webhook!")
}