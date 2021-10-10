package main

import (
	"log"
	"os"
	"time"

	"github.com/actuallycabbage/telegram-rep-bot/internal/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	DB               *db.DB
	Bot              *tgbotapi.BotAPI
	UserRepCooldowns = make(map[int64]map[int64]time.Time) // NOTE: This doesn't work if multiple instances.
)

func main() {
	var err error

	// Connect to databse
	DB, err = db.Connect(&db.Config{
		Type:             "sqlite",
		ConnectionString: "/data/testing.db",
	})
	check(err)

	// Start up the bot
	Bot, err = tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	check(err)

	log.Printf("Authorized on bot: %s", Bot.Self.UserName)

	// Configure polling settings
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Start updating polling goroutine
	updates := Bot.GetUpdatesChan(u)

	// Wait for updates & handle
	for update := range updates {
		updateHandler(&update)

	}
}

// Panic if error is not nil
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
