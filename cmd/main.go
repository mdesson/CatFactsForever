package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/mdesson/CatFactsForever/factmanager"
	"github.com/mdesson/CatFactsForever/scheduler"
	"github.com/mdesson/CatFactsForever/sms"
)

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
			respCode := sms.SendText(msg, sid, token, user.PhoneNumber, from)
			// If http response from Twilio is other than 201, register error
			if respCode != 201 {
				return fmt.Errorf("Error sending text message to %v with code %v", user.Name, respCode)
			}
			// If no error occurred, update the total messages sent to the user and the total number of thanks
			db.Model(&user).Updates(map[string]int{"total_sent": (user.TotalSent + 1), "total_sent_session": (user.TotalSentSession + 1)})
		}
		return nil
	}

	if err := scheduler.AddJob(fmt.Sprint(subscription.ID), subscription.Cron, subscription.Description, true, true, jobFunc); err != nil {
		log.Fatalf("Error registering cat facts job with scheduler:\n%v", err)
	}
	go scheduler.Start()

	r := mux.NewRouter()
	r.HandleFunc("/sms", sms.MakeResponseHandler(db)).Methods("POST")
	http.Handle("/", r)
	if err = http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error starting on server on ':8080':\n%v\n", err)
	}
}
