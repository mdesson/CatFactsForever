package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/mdesson/CatFactsForever/factmanager"
	"github.com/mdesson/CatFactsForever/scheduler"
)

// Response is a Twilio Response
type Response struct {
	Message []string `xml:Message>Body`
}

func main() {
	// Load .env file for account credentials
	if err := godotenv.Load(); err != nil {
		log.Fatal("Please include an .env file with SID and TOKEN values from Twilio")
	}
	sid := os.Getenv("SID")
	token := os.Getenv("TOKEN")
	from := os.Getenv("FROM")
	dbUser := os.Getenv("DB_USER")
	dbHost := os.Getenv("DB_HOST")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	adminName1 := os.Getenv("ADMIN_NAME_1")
	adminPhone1 := os.Getenv("ADMIN_PHONE_1")
	adminName2 := os.Getenv("ADMIN_NAME_2")
	adminPhone2 := os.Getenv("ADMIN_PHONE_2")

	// Initialize database
	db, err := factmanager.Init(dbHost, dbUser, dbPass, dbName, dbPort)
	if err != nil {
		log.Fatalf("Error opening db connection:\n%v", err)
	}
	factmanager.ResetAndPopulate(db, adminName1, adminPhone1, adminName2, adminPhone2, "cat", "facts.csv")
	msg := factmanager.MakeFactMessage("cat", db)
	fmt.Println(msg)

	subscription := &factmanager.Subscription{}
	users := []factmanager.CatEnthusiast{}
	db.Where("id = ?", 1).Find(subscription)
	db.Where("subscription_id = ?", subscription.ID).Find(&users)

	// Add the fact sms job to the scheduler
	jobFunc := func(ctx context.Context) error {
		for _, user := range users {
			msg := factmanager.MakeFactMessage(user.FactCategory, db)
			respCode := sendText(msg, sid, token, user.PhoneNumber, from)
			if respCode != 201 {
				return fmt.Errorf("Error sending text message to %v with code %v", user.Name, respCode)
			}
		}
		return nil
	}

	if err := scheduler.AddJob(fmt.Sprint(subscription.ID), subscription.Cron, subscription.Description, true, true, jobFunc); err != nil {
		log.Fatalf("Error registering cat facts job with scheduler:\n%v", err)
	}
	go scheduler.Start()

	r := mux.NewRouter()
	r.HandleFunc("/sms", exampleHandler).Methods("POST")
	http.Handle("/", r)
	if err = http.ListenAndServe(":8000", nil); err != nil {
		log.Fatalf("Error starting on server on ':8080':\n%v\n", err)
	}
}

func sendText(msg, sid, token, to, from string) int {
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

	responseMsg := Response{[]string{"no u", "and u"}}
	x, _ := xml.Marshal(responseMsg)

	w.Header().Set("Content-Type", "application/xml")
	w.Write(x)
}
