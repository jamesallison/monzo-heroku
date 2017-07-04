package main

import (
	"log"
	"net/http"
	"os"
	"fmt"
	"encoding/json"
	"strings"
	"errors"
	"time"
)

var (
	accessToken string
	slackWebhookUrl string
)


type SlackMessagePayload struct {
	Emoji    string `json:"icon_emoji"`
	Username string `json:"username"`
	Text     string `json:"text"`
}

// split this into another file, stupid go imports 😡
type WebhookPayload struct {
	Type string `json:"type"` // transaction.created
	Data struct {
		ID string `json:"id"`
		Created time.Time `json:"created"`
		Description string `json:"description"`
		Amount int `json:"amount"`
		Currency string `json:"currency"`
		Merchant interface{} `json:"merchant"`
		Notes string `json:"notes"`
		Metadata struct {
			Notes string `json:"notes"`
			P2PTransferID string `json:"p2p_transfer_id"`
		} `json:"metadata"`
		AccountBalance int `json:"account_balance"`
		Attachments interface{} `json:"attachments"`
		Category string `json:"category"`
		IsLoad bool `json:"is_load"`
		Settled time.Time `json:"settled"`
		LocalAmount int `json:"local_amount"`
		LocalCurrency string `json:"local_currency"`
		Updated time.Time `json:"updated"`
		AccountID string `json:"account_id"`
		Counterparty struct {
			Number string `json:"number"`
			UserID string `json:"user_id"`
		} `json:"counterparty"`
		Scheme string `json:"scheme"`
		DedupeID string `json:"dedupe_id"`
		Originator bool `json:"originator"`
		IncludeInSpending bool `json:"include_in_spending"`
	} `json:"data"`
}

func main() {
	// load up environment vars
	port := os.Getenv("PORT")
	accessToken := os.Getenv("ACCESSTOKEN")
	slackWebhook := os.Getenv("SLACKWEBHOOK")

	if port == "" {
		log.Fatal("$PORT must be set")
	}
	if accessToken == "" {
		log.Fatal("$ACCESSTOKEN must be set")
	}
	if slackWebhook == "" {
		log.Fatal("$SLACKWEBHOOK must be set")
	}

	http.HandleFunc("/webhook", webhookHandler)

	http.ListenAndServe(":" + port, nil)
}

func webhookHandler(w http.ResponseWriter, req *http.Request) {
	log.Print("received /webhook")

	// parse payload
	decoder := json.NewDecoder(req.Body)
	var monzoWebhookPayload WebhookPayload
	errParse := decoder.Decode(&monzoWebhookPayload)
	if errParse != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Could not decode")
		log.Print("ERROR: Could not decode payload from Monzo webhook")
		log.Print(errParse.Error())
		return
	}

	// check if type is transaction
	if monzoWebhookPayload.Type == "transaction.created" {
		// send webhook of transaction to slack
		webhookErr := sendSlackWebhook(SlackMessagePayload{
				Emoji: "money_bag",
				Username: "MoneyBot",
				Text: "You have recieved a new transaction. Description: " + monzoWebhookPayload.Data.Description,
			})
		if webhookErr != nil {
			log.Print("ERROR: could not post message to Slack")
			log.Print(webhookErr.Error())
		}
	} else {
		log.Print("Received webhook was not transaction.created, it was:" + monzoWebhookPayload.Type)
	}

	// tell Monzo we have received the webhook all okay
	fmt.Fprintf(w, "OK")
}

func sendSlackWebhook(payload SlackMessagePayload) error {
	// parse the Monzo payload
	sp, _ := json.Marshal(payload)
	p := strings.NewReader(string(sp))
	req, _ := http.NewRequest("POST", slackWebhookUrl, p)

	req.Header.Add("content-type", "application/json")
	req.Header.Add("cache-control", "no-cache")

	res, httpErr := http.DefaultClient.Do(req)

	// check Slack is happy
	if res.StatusCode != 200 || httpErr != nil {
		return errors.New("Slack error code " + res.Status)
	}

	return nil
}