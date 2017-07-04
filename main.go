package main

import (
	"log"
	"net/http"
	"os"
	"fmt"
	"encoding/json"
	"time"
	"strings"
	"errors"
	"strconv"
)

var (
	accessToken string
	slackWebhookUrl string
)


type SlackMessagePayload struct {
	Emoji    string `json:"icon_emoji"`
	Username string `json:"username"`
	Text     string `json:"text"`
	Attachments []SlackAttachment `json:"attachments"`
}

type SlackAttachment struct {
	Fallback string  `json:"fallback"`
	Color    string  `json:"color"`
	Title    string  `json:"title"`
	Text     string  `json:"text"`
	Fields   []SlackField `json:"fields"`
}

type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type MonzoWebhookPayload struct {
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
	accessToken = os.Getenv("ACCESSTOKEN")
	slackWebhookUrl = os.Getenv("SLACKWEBHOOK")

	if port == "" {
		log.Fatal("$PORT must be set")
	}
	if accessToken == "" {
		log.Fatal("$ACCESSTOKEN must be set")
	}
	if slackWebhookUrl == "" {
		log.Fatal("$SLACKWEBHOOK must be set")
	}

	http.HandleFunc("/webhook", webhookHandler)

	http.ListenAndServe(":" + port, nil)
}

func webhookHandler(w http.ResponseWriter, req *http.Request) {
	log.Print("received /webhook")

	// parse payload
	decoder := json.NewDecoder(req.Body)
	var monzoWebhookPayload MonzoWebhookPayload
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
		// send message to slack with transaction details
		webhookErr := sendSlackMessage(SlackMessagePayload{
				Emoji: "moneybag",
				Username: "MoneyBot",
				Text: "You have recieved a new transaction. Description: " + monzoWebhookPayload.Data.Description,
				Attachments: generateAttachments(monzoWebhookPayload),
			})
		if webhookErr != nil {
			log.Print("ERROR: could not post message to Slack")
			log.Print(webhookErr.Error())
		} else {
			log.Print("successfully posted to Slack!")
		}
	} else {
		log.Print("Webhook from Monzo was not transaction.created, it was:" + monzoWebhookPayload.Type)
	}

	// tell Monzo we have received the webhook all okay
	fmt.Fprintf(w, "OK")
}

func sendSlackMessage(payload SlackMessagePayload) error {
	// parse the Monzo payload
	sp, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	p := strings.NewReader(string(sp))
	req, err := http.NewRequest("POST", slackWebhookUrl, p)
	if err != nil {
		return err
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("cache-control", "no-cache")

	res, httpErr := http.DefaultClient.Do(req)

	// check Slack is happy
	if res.StatusCode != 200 || httpErr != nil {
		return errors.New("Slack error code " + res.Status)
	}

	return nil
}

func generateAttachments(payload MonzoWebhookPayload) []SlackAttachment {
	var attachments []SlackAttachment

	// create fields
	amountField := SlackField{
		Title: "Amount",
		Value: strconv.Itoa(payload.Data.Amount) + " pennies",
		Short: true,
	}
	balanceField := SlackField{
		Title: "Current Balance",
		Value: strconv.Itoa(payload.Data.AccountBalance) + " pennies",
		Short: true,
	}
	categoryField := SlackField{
		Title: "Category",
		Value: payload.Data.Category,
		Short: true,
	}
	fields := []SlackField{amountField, balanceField, categoryField}

	// create the attachment
	attachment := SlackAttachment{
		Fallback: payload.Data.Description,
		Color: "#36a64f",
		Title: "New Transaction!",
		Text: payload.Data.Description,
		Fields: fields,
	}

	attachments = append(attachments, attachment)

	return attachments
}