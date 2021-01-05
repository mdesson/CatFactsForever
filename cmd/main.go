package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/mdesson/CatFactsForever/factManager"
	"github.com/mdesson/CatFactsForever/scheduler"
)

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
	// to := os.Getenv("TO")
	from := os.Getenv("FROM")
	dbUser := os.Getenv("DB_USER")
	dbHost := os.Getenv("DB_HOST")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	// Initialize database
	db, err := factManager.Init(dbHost, dbUser, dbPass, dbName, dbPort)
	if err != nil {
		log.Fatalf("Error opening db connection:\n%v", err)
	}

	// Fetch cat enthusiasts and facts
	users := make([]factManager.CatEnthusiast, 0)
	facts := make([]factManager.Fact, 0)
	db.Where("fact_category = ?", "cat").Find(&users)
	db.Where("category = ?", "cat").Find(&facts)

	// Add the fact sms job to the scheduler
	jobFunc := func(ctx context.Context) error {
		for _, user := range users {
			seed := rand.NewSource(time.Now().UnixNano())
			fact := facts[rand.New(seed).Intn(len(facts))]
			sendText(fact.Body, sid, token, user.PhoneNumber, from)
		}
		return nil
	}
	if err := scheduler.AddJob("catFacts", "* * * * *", "Sends cat facts to cat enthusiasts", true, true, jobFunc); err != nil {
		log.Fatalf("Error registering cat facts job with scheduler:\n%v", err)
	}
	go scheduler.Start()

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
