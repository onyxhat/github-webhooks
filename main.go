package main

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.Println("server started")
	http.HandleFunc("/webhook", handleWebhook)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
