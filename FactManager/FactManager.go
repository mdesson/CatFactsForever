package FactManager

import (
	"encoding/csv"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Current status of user's cat fact campaign
type UserStatus int

const (
	PENDING UserStatus = iota
	ACTIVE
	STOPPED
)

func (s UserStatus) String() string {
	return [...]string{"PENDING", "ACTIVE", "STOPPED"}[s]
}

type CatEnthusiast struct {
	gorm.Model
	Name          string
	PhoneNumber   string
	Status        UserStatus
	FactCategory  int
	TotalSent     int
	ThanksCounter int
}

type Fact struct {
	gorm.Model
	Body     string
	Category string
}

func Init(host, user, pass, name, port string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=America/Toronto", host, user, pass, name, port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Create tables and populate with facts from facts.csv
func Populate(db *gorm.DB) error {
	// Perform migrations to create tables
	db.AutoMigrate(&Fact{})
	db.AutoMigrate(&CatEnthusiast{})

	// Populate table with facts
	file, err := os.Open("facts.csv")
	if err != nil {
		return err
	}
	r := csv.NewReader(file)
	facts, err := r.ReadAll()
	if err != nil {
		return err
	}
	for _, pair := range facts {
		fact := &Fact{Body: pair[0], Category: pair[1]}
		db.Create(fact)
	}
	return nil
}
