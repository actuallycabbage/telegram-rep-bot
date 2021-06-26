package main

import (
	"errors"
	"log"
	"os"
	"telegram_rep_tracker/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	DB         *db.DB
	TeleramBot *tgbotapi.BotAPI
)

func main() {
	var err error

	// Connect to databse
	DB, err = db.Connect(&db.Config{
		Type:             "sqlite",
		ConnectionString: "/data/testing.db",
	})

	if err != nil {
		log.Fatal(err.Error())
	}

	TeleramBot, err = tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	TeleramBot.Debug = false

	log.Printf("Authorized on bot: %s", TeleramBot.Self.UserName)

	// Configure polling settings
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Start updating polling goroutine
	updates, err := TeleramBot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			go messageHandler(update.Message)
		}

	}
}

func messageHandler(msg *tgbotapi.Message) error {
	// BUG: Forwarded messages might give issues.

	// Search for rep triggers
	if msg.ReplyToMessage != nil {
		repchange, err := determineRep(msg)
		if err != nil {
			log.Fatal(err)
		}
		if repchange != 0 {
			m := map[string]interface{}{
				"trigger": "chat.message",
			}

			DB.CreateRepEvent(msg.Chat.ID, msg.ReplyToMessage.From.ID, msg.From.ID, repchange, m)
		}
	}

	return errors.New("Not implemented")
}

func determineRep(msg *tgbotapi.Message) (repChange int, err error) {
	// Constants until chat settings are implemented
	positiveRep := []string{"^(\\+)$", "^(kek)$", "^(fucking\\skek)$", "^(holy\\skek)$", "^(ty)$", "^.*(thank\\syou)$", "^(happy\\sbirthday)$", "^(991019)$"}
	negativeRep := []string{"^(-)$", "^(repost)$", "^(slug)$", "^(/ruin)$", "^(no\\sthank\\syou)$"}

	// Any positive rep changes?
	positiveFind, err := regexMatchArray(&positiveRep, &msg.Text)
	if err != nil {
		log.Println(err.Error())
	}

	if positiveFind {
		return 1, nil
	}

	// Any negative rep changes?
	negativeFind, err := regexMatchArray(&negativeRep, &msg.Text)
	if err != nil {
		log.Println(err.Error())
	}

	if negativeFind {
		return -1, nil
	}

	// Okay didn't find anything
	return 0, nil
}
