package main

import (
	"log"
	"time"

	"github.com/actuallycabbage/telegram-rep-bot/internal/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Process telegram update events
func updateHandler(update *tgbotapi.Update) {

	// Fetch chat settings
	var settings *db.AccountSettings
	var err error

	if err != nil {
		log.Println(err)
	}

	// There's a few different types of updates. We're only interested in messages at the moment.
	if update.Message != nil {
		settings, err = DB.GetChatSettings(update.Message.Chat.ID)
		if update.Message.IsCommand() {
			go commandHandler(update.Message, settings)
		} else {
			go messageHandler(update.Message, settings)
		}
	}

}

// Process commands
func commandHandler(msg *tgbotapi.Message, settings *db.AccountSettings) error {
	switch msg.Command() {
	case "toprep":
		toprepCommandHandler(msg, settings)
		break
	case "bottomrep":
		bottomrepCommandHandler(msg, settings)
		break
	default:
		break
	}
	return nil
}

// Process message events
func messageHandler(msg *tgbotapi.Message, settings *db.AccountSettings) error {

	// === CHECK FOR REP EVENTS ===
	go repHandler(msg, settings)

	return nil
}

func repHandler(msg *tgbotapi.Message, settings *db.AccountSettings) error {
	// We've got a few conditions where we don't handle rep.

	// 1: The user must be replying to a message
	if msg.ReplyToMessage == nil {
		return nil
	}

	// 2: The user must not be replying to this bot
	if msg.ReplyToMessage.From.ID == Bot.Self.ID {
		return nil
	}

	// 3: The user must not be replying to themselves
	//if msg.From.ID == msg.ReplyToMessage.From.ID {
	//	return nil
	//}

	// 4: Does the user have a rep cooldown.
	if val, exists := UserRepCooldowns[msg.Chat.ID][msg.From.ID]; exists {
		// Has it expired? Remove it.
		if time.Now().After(val) {
			delete(UserRepCooldowns[msg.Chat.ID], msg.From.ID)
		} else {
			log.Printf("User `%d` still has a cooldown", msg.From.ID)
			return nil
		}
	}

	// 5: Rep must be enabled for the account the chat is linked to.
	// This sits last as it requires a DB operation.
	if settings.Rep.Enabled == false {
		return nil
	}

	var err error

	var m = map[string]interface{}{} // Metadata of the event
	var positiveFind bool = false    // Positive rep event found
	var negativeFind bool = false    // Negative rep event found
	var repchange int = 0            // How much rep to adjust

	if msg.Sticker != nil {
		// Check if sticker is in our positive stickers list
		positiveFind = arrayContains(msg.Sticker.FileUniqueID, settings.Rep.PositiveStickers)

		// Check if sticker is in our negative stickers list
		if positiveFind != true {
			negativeFind = arrayContains(msg.Sticker.FileUniqueID, settings.Rep.NegativeStickers)
		}

		m["trigger"] = "chat.sticker"
		m["sticker_id"] = msg.Sticker.FileUniqueID
		m["sticker_emoji"] = msg.Sticker.Emoji

	} else if msg.Text != "" {

		// Any positive chat trigger rep changes?
		positiveFind, err = regexMatchArray(&settings.Rep.PositiveTriggers, &msg.Text)
		if err != nil {
			log.Println(err.Error())
		}

		// Any negative chat trigger rep changes?
		if positiveFind != true {
			negativeFind, err = regexMatchArray(&settings.Rep.NegativeTriggers, &msg.Text)
			if err != nil {
				log.Println(err.Error())
			}
		}

		m["trigger"] = "chat.message"

	}

	m["rep_message_id"] = msg.ReplyToMessage.MessageID
	m["origin_message_id"] = msg.MessageID

	if positiveFind {
		repchange = 1
	} else if negativeFind {
		repchange = -1
	}

	if repchange != 0 {
		DB.CreateRepEvent(msg.Chat.ID, msg.ReplyToMessage.From.ID, msg.From.ID, repchange, m)
		log.Printf("Rep change (%d) of '%s' type for user %d triggered by %d on chat %d", repchange, m["trigger"], msg.ReplyToMessage.From.ID, msg.From.ID, msg.Chat.ID)

		// Fetch the duration from the settings
		length, err := time.ParseDuration(settings.Rep.Cooldown.Duration)
		if err != nil {
			log.Println(err.Error())
		}

		// Add a cooldown for the user
		if UserRepCooldowns[msg.Chat.ID] == nil {
			UserRepCooldowns[msg.Chat.ID] = make(map[int64]time.Time)
		}
		UserRepCooldowns[msg.Chat.ID][msg.From.ID] = time.Now().Add(length)
	}
	return nil
}
