package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Response struct {
	Message []string `xml:Message>Body`
}

func main() {
	// Load .env file for account credentials
	if err := godotenv.Load(); err != nil {
		log.Fatal("Please include an .env file with SID and TOKEN values from Twilio")
	}
	// sid := os.Getenv("SID")
	// token := os.Getenv("TOKEN")
	// to := os.Getenv("TO")
	// from := os.Getenv("FROM")
	r := mux.NewRouter()
	r.HandleFunc("/sms", exampleHandler).Methods("POST")
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

func sendText(msg, sid, token, to, from string) int {
	// Config for text message
	data := url.Values{}
	data.Set("To", to)
	data.Set("From", from)
	data.Set("Body", "test123")

	msgURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", sid)

	// Set up request
	r, _ := http.NewRequest(http.MethodPost, msgURL, strings.NewReader(data.Encode()))
	r.SetBasicAuth(sid, token)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	// Send Request
	client := &http.Client{}
	resp, _ := client.Do(r)
	fmt.Println(resp.StatusCode)

	return resp.StatusCode
}

// Handlers
func exampleHandler(w http.ResponseWriter, r *http.Request) {
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

	responseMsg := Response{[]string{"no u"}}
	x, _ := xml.Marshal(responseMsg)

	w.Header().Set("Content-Type", "application/xml")
	w.Write(x)
}
