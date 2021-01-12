package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/mdesson/CatFactsForever/factmanager"
	"github.com/mdesson/CatFactsForever/scheduler"
	"github.com/mdesson/CatFactsForever/sms"
)

func main() {
	// Begin logging to file
	f, err := os.OpenFile("catfacts-logs", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

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

	// Initialize database
	db, err := factmanager.Init(dbHost, dbUser, dbPass, dbName, dbPort)
	if err != nil {
		log.Fatalf("Error opening db connection:\n%v", err)
	}

	// populate if there are no facts
	var facts []factmanager.Fact
	if err := db.Find(&facts).Error; err != nil {
		log.Fatalf("Error getting facts on startup: %v", err)
	}
	if len(facts) == 0 {
		log.Println("No facts in database. Running populate on all tables")
		if err := factmanager.Populate(db, "cat", "facts.csv"); err != nil {
			log.Fatalf("Error populating database on startup: %v", err)
		}
		time.Sleep(1 * time.Second)
		log.Println("Completed database population")
	}

	msg := factmanager.MakeFactMessage("cat", db)
	log.Println(msg)

	schedules := []factmanager.Subscription{}
	if err := db.Find(&schedules).Error; err != nil {
		log.Printf("error listing subscriptions: %v", err)
		log.Fatalf("Error occurred fetching schedules")
	}

	// Add the fact sms job to the scheduler
	for _, s := range schedules {
		subscription := s
		jobFunc := func(ctx context.Context) error {
			users := []factmanager.CatEnthusiast{}
			if err := db.Where("subscription_id = ?", subscription.ID).Find(&users).Error; err != nil {
				return fmt.Errorf("Error fetching users that have subscriptionID %v: %v", subscription.ID, err)
			}
			for _, user := range users {
				if user.Active {
					msg := factmanager.MakeFactMessage(user.FactCategory, db)
					respCode := sms.SendText(msg, sid, token, user.PhoneNumber, from)
					// If http response from Twilio is other than 201, register error
					if respCode != 201 {
						return fmt.Errorf("Error sending text message to %v with code %v", user.Name, respCode)
					}
					// If no error occurred, update the total messages sent to the user and the total number of thanks
					if err := db.Model(&user).Updates(&factmanager.CatEnthusiast{TotalSent: (user.TotalSent + 1), TotalSentSession: (user.TotalSentSession + 1)}).Error; err != nil {
						return fmt.Errorf("Error updating user %v's stats: %v", user.Name, err)
					}
				}
			}
			return nil
		}

		if err := scheduler.AddJob(fmt.Sprint(subscription.ID), subscription.Cron, subscription.Description, true, true, jobFunc); err != nil {
			log.Fatalf("Error registering cat facts job with scheduler:\n%v", err)
		}
	}
	go scheduler.Start()

	r := mux.NewRouter()
	r.HandleFunc("/sms", sms.MakeResponseHandler(db)).Methods("POST")
	http.Handle("/", r)
	if err = http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error starting on server on ':8080':\n%v\n", err)
	}
}
