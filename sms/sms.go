package sms

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/mdesson/CatFactsForever/admin"
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
		phoneNumber := bodyMap["From"][0]

		// declarations of user, their subscription, and the xml to be marshalled
		user := factmanager.CatEnthusiast{}
		subscription := factmanager.Subscription{}
		var x []byte

		if phoneNumber == os.Getenv("ADMIN_PHONE_1") || phoneNumber == os.Getenv("ADMIN_PHONE_2") {
			// all input is case insensitive, all db data is stored in lower case
			incomingMsg = strings.ToLower(incomingMsg)

			// trim any leading/trailing whitespace
			incomingMsg = strings.TrimSpace(incomingMsg)

			// Get admin command and its arguments
			words := strings.Split(incomingMsg, " ")
			cmd := words[0]
			args := words[1:]

			// Declare reply to admin
			var reply string

			// Parse command and its arguments
			if cmd == "help" {
				reply = admin.Help()
			} else if cmd == "add" {
				if len(args) != 4 {
					reply = "bad format for adding. see help"
				} else {
					var ok bool
					var freq string
					reply, freq, ok = admin.Add(args[0], args[1], args[2], args[3], db)
					if ok {
						// welcome user to cat facts with their first fact
						fact := factmanager.GetRandomFact(db, args[3])
						msg := "Welcome to CAT FACTS! We deliver purrfectly accurate feline friend facts and sometimes pawful puns straight to your smartphone!"
						msg = fmt.Sprintf("%v You will receive a CAT FACT <%v>. Reply UNSUBSCRIBE to unsubscribe.\n%v", msg, freq, fact)
						SendText(msg, os.Getenv("SID"), os.Getenv("TOKEN"), args[1], os.Getenv("FROM"))
					}
				}
			} else if cmd == "start" {
				if len(args) != 1 {
					reply = "bad format for start. see help"
				} else {
					reply = admin.Start(args[0], db)
				}
			} else if cmd == "stop" {
				if len(args) != 1 {
					reply = "bad format for stop. see help"
				} else {
					reply = admin.Stop(args[0], db)
				}
			} else if cmd == "info" {
				if len(args) != 1 {
					reply = "bad format for info. see help"
				} else {
					reply = admin.Info(args[0], db)
				}
			} else if cmd == "update" {
				if len(args) != 2 {
					reply = "bad format for update. see help"
				} else {
					reply = admin.Update(args[0], args[1], db)
				}
			} else if cmd == "list" {
				if len(args) != 1 {
					reply = "bad format for list. see help"
				} else if args[0] == "users" {
					reply = admin.ListUsers(db)
				} else if args[0] == "schedules" {
					reply = admin.ListSubscriptions(db)
				} else if args[0] == "jobs" {
					reply = admin.ListJobs()
				} else {
					reply = "can't list that. See help"
				}
			} else if cmd == "reset" {
				if len(args) == 1 && args[0] == "confirm" {
					factmanager.Reset(db)
					reply = "deleted all tables in database"
				} else {
					reply = "invalid reset command. be careful, this command is destructive"
				}
			} else if cmd == "populate" {
				if len(args) == 1 && args[0] == "confirm" {
					if err := factmanager.Populate(db, "cat", "facts.csv"); err != nil {
						reply = fmt.Sprintf("Error populating: %v", err)
					} else {
						reply = "repopulated database with starter data"
					}
				} else {
					reply = "invalid populate command. be careful, this command is destructive"
				}
			} else {
				reply = "don't know that one. type help to see available options"
			}
			x, _ = xml.Marshal(Response{[]string{reply}})
		} else {
			// populate user and subscription
			if err := db.Where("phone_number = ?", phoneNumber).First(&user).Error; err != nil {
				log.Printf("Error looking up incoming text user: %v", err)
				return
			}
			if err := db.Where("id = ?", user.SubscriptionID).First(&subscription).Error; err != nil {
				log.Printf("Error looking up user subscription: %v", err)
				return
			}

			// If user unsubscribes, clear total messages sent during current subscription cycle
			if strings.ToLower(incomingMsg) == "thanks" || strings.ToLower(incomingMsg) == "stop" {
				if err := db.Model(&user).Update("total_sent_session", 0).Error; err != nil {
					log.Printf("Error looking up incoming text user: %v", err)
				}
				return
			}

			// Don't send facts for help messages
			if strings.ToLower(incomingMsg) == "help" {
				return
			}

			// fetch outgoing message
			outgoingMsg := factmanager.MakeReplyMessage(user.FactCategory, db)

			// Inlcude a thanks message if user has reached their subscription's threshold
			if user.TotalSentSession >= subscription.ThanksThreshold {
				thanks := factmanager.GetRandomThanks(db, user.FactCategory)
				x, _ = xml.Marshal(Response{[]string{outgoingMsg, thanks}})

			} else {
				x, _ = xml.Marshal(Response{[]string{outgoingMsg}})
			}

			// Increment total messages sent to user by one
			if err := db.Model(&user).Updates(&factmanager.CatEnthusiast{TotalSent: (user.TotalSent + 1), TotalSentSession: (user.TotalSentSession + 1)}).Error; err != nil {
				log.Printf("Error updating user %v's stats: %v", user.Name, err)
				return
			}
		}

		w.Header().Set("Content-Type", "application/xml")
		w.Write(x)
	}
}
