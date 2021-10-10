package main

import (
	"strconv"
	"strings"

	"github.com/actuallycabbage/telegram-rep-bot/internal/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	LeaderboardAscending  = iota + 1
	LeaderboardDescending = iota + 1
)

func repCommandHandler(msg *tgbotapi.Message, settings *db.AccountSettings, leaderboardDirection int) error {
	var leaderboardLimit = 10

	// Is there a specific limit on leaderboard size
	arguments := strings.Fields(msg.CommandArguments())
	if len(arguments) > 0 {
		i, err := strconv.Atoi(arguments[0])
		if err == nil {
			leaderboardLimit = i
		}
	}

	// Bot message config
	m := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           msg.Chat.ID,
			ReplyToMessageID: 0,
		},
		ParseMode: "MarkdownV2",
	}

	switch leaderboardDirection {
	case LeaderboardAscending:
		m.Text = renderLeaderboard(DB.GetChatRep(msg.Chat.ID, "asc", leaderboardLimit), msg.Chat.ID)
		break
	case LeaderboardDescending:
		m.Text = renderLeaderboard(DB.GetChatRep(msg.Chat.ID, "desc", leaderboardLimit), msg.Chat.ID)
		break
	}

	Bot.Send(m)

	return nil

}

func toprepCommandHandler(msg *tgbotapi.Message, settings *db.AccountSettings) error {
	repCommandHandler(msg, settings, LeaderboardDescending)
	return nil
}

func bottomrepCommandHandler(msg *tgbotapi.Message, settings *db.AccountSettings) error {
	repCommandHandler(msg, settings, LeaderboardDescending)
	return nil
}
