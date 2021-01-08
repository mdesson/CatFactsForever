package factmanager

import "gorm.io/gorm"

// CatEnthusiast represents users of CatFacts
type CatEnthusiast struct {
	gorm.Model
	Name             string
	PhoneNumber      string // Format +1XXXXXXXXXX
	Active           bool   // Send facts to active user
	FactCategory     string
	SubscriptionID   uint
	TotalSentSession int // Total messages sent to user during current subscription
	TotalSent        int // Total messages sent to user over all time
}

// Fact is a simple fact on any category, such as "cat"
type Fact struct {
	gorm.Model
	Body     string
	Category string
}

// Greeting is prepended to facts sent to the user
type Greeting struct {
	gorm.Model
	Category string
	Body     string
}

// ThanksMessage passive-aggressively urges the user to say thanks, the secret unsubscribe word
type ThanksMessage struct {
	gorm.Model
	Category string
	Body     string
}

// ReplyMessage is prepended to a fact any time the user trieds to reply to the text message
// The only time this will not trigger is if the user sends the correct unsubscribe keyword
type ReplyMessage struct {
	gorm.Model
	Category string
	Body     string
}

// Category is a category of fact, such as cat
type Category struct {
	gorm.Model
	Name           string
	SubscribeMsg   string
	UnsubscribeMsg string
}

// Subscription describes the frequency with which text messages are sent, and how soon unsubscribe hints begin
type Subscription struct {
	gorm.Model
	Frequency       string // Descriptive name such as "daily" or "every fifteen minutes"
	Description     string // Short description of the subscription
	Cron            string // cron string, only ints and special characters *,- accepted
	ThanksThreshold int    // Number of messages sent prior to beginning of say thanks hints
}
