package sms

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/mdesson/CatFactsForever/factmanager"
	"gorm.io/gorm"
)

// Response is a Twilio sms response to be sent as xml
// It can contain any number of text Messages
type Response struct {
	Message []string `xml:Message>Body`
}

// SendText sends an sms message to the specified number
func SendText(msg, sid, token, to, from string) int {
	// Config for text message
	data := url.Values{}
	data.Set("To", to)
	data.Set("From", from)
	data.Set("Body", msg)

	msgURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", sid)

	// Set up request
	r, _ := http.NewRequest(http.MethodPost, msgURL, strings.NewReader(data.Encode()))
	r.SetBasicAuth(sid, token)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	// Send Request
	client := &http.Client{}
	resp, _ := client.Do(r)

	return resp.StatusCode
}

// MakeResponseHandler generates an http handler that sends responses to sms messages as they come in
func MakeResponseHandler(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatalf("Error decoding request body:\n%v", err)
		}

		bodyMap, err := url.ParseQuery(string(body))
		if err != nil {
			log.Fatalf("Error converting body to map:\n%v", err)
		}

		// Get the income message and phone number of user
		incomingMsg := bodyMap["Body"][0]
		phoneNumber := bodyMap["From"]

		// declarations of user, their subscription, and the xml to be marshalled
		user := factmanager.CatEnthusiast{}
		subscription := factmanager.Subscription{}
		var x []byte

		// populate user and subscription
		db.Where("phone_number = ?", phoneNumber).Find(&user)
		db.Where("id = ?", user.SubscriptionID).Find(&subscription)

		if strings.ToLower(incomingMsg) == "thanks" || strings.ToLower(incomingMsg) == "stop" {
			db.Model(&user).Update("total_sent_session", 0)
			return
		}

		// fetch outgoing message
		outgoingMsg := factmanager.MakeReplyMessage("cat", db)

		// Inlcude a thanks message if user has reached their subscription's threshold
		if user.TotalSentSession >= subscription.ThanksThreshold {
			thanks := factmanager.GetRandomThanks(db, user.FactCategory)
			x, _ = xml.Marshal(Response{[]string{outgoingMsg, thanks}})

		} else {
			x, _ = xml.Marshal(Response{[]string{outgoingMsg}})
		}

		// Increment total messages sent to user by one
		db.Model(&user).Updates(&factmanager.CatEnthusiast{TotalSent: (user.TotalSent + 1), TotalSentSession: (user.TotalSentSession + 1)})

		w.Header().Set("Content-Type", "application/xml")
		w.Write(x)
	}
}
