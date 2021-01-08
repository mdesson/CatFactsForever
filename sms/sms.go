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
)

// response is a Twilio sms response to be sent as xml
// It can contain any number of text Messages
type response struct {
	message []string `xml:Message>Body`
}

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

// ResponseHandler is an http handler that sends responses to sms messages as they come in
func ResponseHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("Error decoding request body:\n%v", err)
	}

	bodyMap, err := url.ParseQuery(string(body))
	if err != nil {
		log.Fatalf("Error converting body to map:\n%v", err)
	}

	msg := bodyMap["Body"][0]

	log.Println(msg)

	responseMsg := response{[]string{"no u", "and u"}}
	x, _ := xml.Marshal(responseMsg)

	w.Header().Set("Content-Type", "application/xml")
	w.Write(x)
}
