package admin

import (
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/mdesson/CatFactsForever/factmanager"
	"github.com/mdesson/CatFactsForever/scheduler"
	"gorm.io/gorm"
)

// Help displays the list of admin commands
func Help() string {
	return `Admin commands are:
add name +1XXXYYYZZZZ suscriptionID category - add user

start name - enables sms on user

stop name - disables sms on user

info name - details about user

update name subscriptionID - change user's schedule

list users - lists all users

list schedules - lists all schedules

reset confirm - deletes all data [DANGER]

populate confirm - puts in starter data [DANGER]
`
}

// ListUsers will display either a list of users or jobs
func ListUsers(db *gorm.DB) string {
	var users []factmanager.CatEnthusiast
	output := ""
	if err := db.Find(&users).Error; err != nil {
		log.Printf("error listing users: %v", err)
		return "an error occurred fetching users"
	}
	if len(users) == 0 {
		return "no users, use 'add' to add some"
	}
	for _, user := range users {
		var status string
		if user.Active {
			status = "active"
		} else {
			status = "inactive"
		}
		output = fmt.Sprintf("%v%v is %v with category %v\n", output, user.Name, status, user.FactCategory)
	}
	return output
}

// ListJobs displays a list of all jobs and their status
func ListJobs() string {
	statuses := scheduler.Statuses()
	output := ""
	for _, status := range statuses {
		output = fmt.Sprintf("%v%v\n\n", output, status)
	}
	return output
}

// ListSubscriptions displays the name and ID of each available subscription type
func ListSubscriptions(db *gorm.DB) string {
	output := "Schedule IDs and names:\n"

	schedules := []factmanager.Subscription{}
	if err := db.Find(&schedules).Error; err != nil {
		log.Printf("error listing subscriptions: %v", err)
		return "an error occurred fetching schedules"
	}

	if len(schedules) == 0 {
		return "no schedules found, your admin can add some for you"
	}

	for _, schedule := range schedules {
		output = fmt.Sprintf("%v%v: %v\n", output, schedule.ID, schedule.Frequency)
	}
	return output
}

// Start will set the user to active
func Start(name string, db *gorm.DB) string {
	if err := db.Model(&factmanager.CatEnthusiast{}).Where("name = ?", name).Update("active", true).Error; err != nil {
		log.Printf("error setting user %v to active: %v", name, err)
		return "an error occurred setting user to active"
	}
	return "done"
}

// Stop will set the user to inactive
func Stop(name string, db *gorm.DB) string {
	if err := db.Model(&factmanager.CatEnthusiast{}).Where("name = ?", name).Update("active", false).Error; err != nil {
		log.Printf("error setting user %v to inactive: %v", name, err)
		return "an error occurred setting user to inactive"
	}
	return "done"
}

// Info displays all details on given user
func Info(name string, db *gorm.DB) string {
	user := &factmanager.CatEnthusiast{}
	if err := db.Where("name = ?", name).First(user).Error; err != nil {
		return "user not found"

	}
	userInfo := fmt.Sprintf(`Name: %v
Phone: %v
Active: %v
Category: %v
SubscriptionID: %v
Total cat facts: %v`,
		user.Name,
		user.PhoneNumber,
		user.Active,
		user.FactCategory,
		user.SubscriptionID,
		user.TotalSent)

	return userInfo
}

// Add will add a new user
func Add(userName, phoneNumber, subID, category string, db *gorm.DB) (reply, frequency string, ok bool) {
	// Validate phone number format
	r := regexp.MustCompile(`\+1\d{10}`)
	if !r.MatchString(phoneNumber) {
		return "phone number should be format +1XXXYYYZZZZ", "", false
	}

	// Validate subscription ID is int
	uintSubID, err := strconv.ParseUint(subID, 10, 32)
	if err != nil {
		return "make sure the subscription ID is a number", "", false
	}

	// Validate subscription ID exists
	sub := &factmanager.Subscription{}
	if err := db.Where("id = ?", subID).First(sub).Error; err != nil {
		return "subscription id not found. try 'list subscriptions'", "", false
	}

	// Validate unique name and phone number
	// gorm will not return error if unique constraint is violated
	u := &factmanager.CatEnthusiast{}
	if err := db.Where("name = ? OR phone_number = ?", userName, phoneNumber).First(u).Error; err == nil {
		return "user with name or phone number already exists", "", false
	}

	user := &factmanager.CatEnthusiast{
		Name:             userName,
		PhoneNumber:      phoneNumber,
		Active:           true,
		FactCategory:     category,
		SubscriptionID:   uint(uintSubID),
		TotalSentSession: 0,
		TotalSent:        0,
	}
	if err := db.Create(user).Error; err != nil {
		return "", "something went wrong, it's probably not your fault", false
	}
	return fmt.Sprintf("%v was added with the subscription %v", user.Name, sub.Frequency), sub.Frequency, true
}

// Update will alter the user's subscription
func Update(userName, subID string, db *gorm.DB) string {
	// Validate subscription ID is int
	uID, err := strconv.ParseUint(subID, 10, 32)
	if err != nil {
		return "make sure the subscription ID is a number"
	}
	user := &factmanager.CatEnthusiast{}
	sub := &factmanager.Subscription{}

	// Fetch user and subscription, ensure both exist
	if err := db.Where("name = ?", userName).First(user).Error; err != nil {
		return "user not found. try 'list users'"
	}
	if err := db.Where("id = ?", subID).First(sub).Error; err != nil {
		return "subscription id not found. try 'list subscriptions'"
	}

	// Update data, notify on error
	user.SubscriptionID = uint(uID)
	if err := db.Save(user).Error; err != nil {
		return "error saving new user's subscription.\nNot your fault, user and subscription both exist"
	}

	return fmt.Sprintf("%v's subscripion is now %v", user.Name, sub.Frequency)
}
