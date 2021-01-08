package factmanager

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// GetRandomFact provides a random fact from the given category
func GetRandomFact(db *gorm.DB, category string) string {
	facts := make([]Fact, 0)
	db.Where("category = ?", category).Find(&facts)
	seed := rand.NewSource(time.Now().UnixNano())
	fact := facts[rand.New(seed).Intn(len(facts))]
	return fact.Body
}

// GetRandomThanks provides a random passive-aggressive thanks message
func GetRandomThanks(db *gorm.DB, category string) string {
	allThanks := make([]ThanksMessage, 0)
	db.Where("category = ?", category).Find(&allThanks)
	seed := rand.NewSource(time.Now().UnixNano())
	thanks := allThanks[rand.New(seed).Intn(len(allThanks))]
	return thanks.Body
}

// MakeFactMessage generates a fact for the given category
func MakeFactMessage(category string, db *gorm.DB) string {
	// Fetch the fact
	fact := GetRandomFact(db, category)

	// Select a random greeting
	greetings := make([]Greeting, 0)
	db.Where("category = ?", category).Find(&greetings)
	seed := rand.NewSource(time.Now().UnixNano())
	greeting := greetings[rand.New(seed).Intn(len(greetings))]
	msg := fmt.Sprintf("%s\n\n%s", greeting.Body, fact)

	return msg
}

// MakeThanksMessage generates a reply message for the given category
func MakeThanksMessage(category string, db *gorm.DB) string {
	// Fetch the fact
	fact := GetRandomFact(db, category)

	// Select a random greeting
	replies := make([]ReplyMessage, 0)
	db.Where("category = ?", category).Find(&replies)
	seed := rand.NewSource(time.Now().UnixNano())
	reply := replies[rand.New(seed).Intn(len(replies))]
	msg := fmt.Sprintf("%s\n\n%s", reply.Body, fact)

	return msg
}

// Init establishes a postgresql database connection
func Init(host, user, pass, name, port string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=America/Toronto", host, user, pass, name, port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Perform migrations any migrations
	db.AutoMigrate(&Category{})
	db.AutoMigrate(&Greeting{})
	db.AutoMigrate(&Subscription{})
	db.AutoMigrate(&Fact{})
	db.AutoMigrate(&ThanksMessage{})
	db.AutoMigrate(&ReplyMessage{})
	db.AutoMigrate(&CatEnthusiast{})

	return db, nil
}

// ResetAndPopulate drops all tables and populates them with default data about cats
func ResetAndPopulate(db *gorm.DB, adminName1, adminPhone1, adminName2, adminPhone2, categoryName, factCSV string) error {
	// Empty all tables
	db.Migrator().DropTable(&Greeting{})
	db.Migrator().DropTable(&Fact{})
	db.Migrator().DropTable(&ThanksMessage{})
	db.Migrator().DropTable(&ReplyMessage{})
	db.Migrator().DropTable(&CatEnthusiast{})
	db.Migrator().DropTable(&Subscription{})
	db.Migrator().DropTable(&Category{})

	db.Migrator().CreateTable(&Greeting{})
	db.Migrator().CreateTable(&Fact{})
	db.Migrator().CreateTable(&ThanksMessage{})
	db.Migrator().CreateTable(&ReplyMessage{})
	db.Migrator().CreateTable(&CatEnthusiast{})
	db.Migrator().CreateTable(&Category{})
	db.Migrator().CreateTable(&Subscription{})

	greetings := []Greeting{
		{Category: categoryName, Body: "CAT FACT ATTACK!"},
		{Category: categoryName, Body: "Here's your CAT FACT!"},
		{Category: categoryName, Body: "Cat. 😻"},
		{Category: categoryName, Body: "CAT FACTS here with another purrrrrrfect feline fact!"},
	}
	db.CreateInBatches(greetings, 3)

	// Populate starter data
	category := &Category{
		Name:           "cat",
		SubscribeMsg:   "Thank you for subscribing to CAT FACTS, the best source of fun facts about cool kitties and famous felines!\nReply UNSUBSCRIBE if you do not want to receive more facts.",
		UnsubscribeMsg: "You're very welcome! As a true Cat Enthusiast you clearly are no longer in need of more cat facts. If you ever want to resubscribe just reply START",
	}
	db.Create(category)

	thanks := []ThanksMessage{
		{Category: categoryName, Body: "Would it be so hard to say thanks? We're working hard over here at CAT FACTS to provide you with the highest quality facts."},
		{Category: categoryName, Body: "A little bit of gratefulness for our great CAT FACTS would go a long way over here."},
		{Category: categoryName, Body: "I want chicken, I want liver, meow-thanks, meow-thanks, be thanks-giver."},
		{Category: categoryName, Body: "It took us a decade to prepare these Fresh Feline Facts, and we're finding you just a little ungrateful overe here."},
	}
	db.CreateInBatches(thanks, 4)

	replies := []ReplyMessage{
		{Category: categoryName, Body: "Glad to hear you're enjoying CAT FACTS! Here's another one:"},
		{Category: categoryName, Body: "Thank you for subscribing to CAT FACTS! This next fact is a meowthful:"},
		{Category: categoryName, Body: "We love hearing from a CAT FACTS FAN! Here's a bonus feline fact since you lov us so much:"},
		{Category: categoryName, Body: "Unsubscribe successful. We hoped you enjoyed your time with - Just kittying! Here's another CAT FACT:"},
	}
	db.CreateInBatches(replies, 4)

	// Populate facts from csv
	file, err := os.Open(factCSV)
	if err != nil {
		return err
	}
	r := csv.NewReader(file)
	facts, err := r.ReadAll()
	if err != nil {
		return err
	}
	catFacts := make([]Fact, len(facts))

	for _, fact := range facts {
		fact := Fact{Category: categoryName, Body: fact[0]}
		catFacts = append(catFacts, fact)
	}
	db.CreateInBatches(catFacts, len(facts))

	// cron to send every 15 minutes during the day
	subscription := &Subscription{
		Frequency:       "every fifteen minutes",
		Description:     "Will send at X:00, X:15, X:30, and X:45 between 9am and 10pm",
		Cron:            "0,15,30,45 9-21 * * *",
		ThanksThreshold: 10,
	}
	db.Create(subscription)

	// Add addmin as catenthusiast
	adminUsers := []CatEnthusiast{
		{
			Name:           adminName1,
			PhoneNumber:    adminPhone1,
			Active:         true,
			FactCategory:   categoryName,
			SubscriptionID: subscription.ID,
			TotalSent:      0,
		},
		{
			Name:           adminName2,
			PhoneNumber:    adminPhone2,
			Active:         true,
			FactCategory:   categoryName,
			SubscriptionID: subscription.ID,
			TotalSent:      0,
		},
	}
	db.Create(adminUsers)

	return nil
}
