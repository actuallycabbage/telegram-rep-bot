package main

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/actuallycabbage/telegram-rep-bot/internal/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Attempts to match the target string with any of the regex expressions in the conditions slice
//
// Will complain if it cannot compile the regex
func regexMatchArray(conditions *[]string, target *string) (bool, error) {
	for _, condition := range *conditions {
		// Attempt to match
		result, err := regexp.MatchString(condition, *target)

		// Issues
		if err != nil {
			return false, errors.New("Regex compile failed for " + condition)
		}

		// Found?
		if result {
			return true, nil
		}

	}

	return false, nil
}

// Check if slice contains item
func arrayContains(item string, array []string) bool {
	for _, val := range array {
		if strings.Compare(val, item) == 0 {
			return true
		}
	}

	return false
}

func renderLeaderboard(board []db.LeaderboardEntry, chatID int64) string {
	// Find padding so they're all in line.
	var padding int

	padding = len(fmt.Sprintf("%d", board[0].Rep))

	p := len(fmt.Sprintf("%d", board[len(board)-1].Rep))

	if p > padding {
		padding = p
	}

	var entries []string

	for _, val := range board {
		// TODO: Seriously need to cache this!!!!
		user, err := Bot.GetChatMember(tgbotapi.GetChatMemberConfig{
			ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
				ChatID: chatID,
				UserID: val.UserID,
			},
		})

		if err != nil {
			log.Fatal(err)
		}

		s := fmt.Sprintf("`%"+fmt.Sprintf("%d", padding+1)+"v `", val.Rep) + fmt.Sprintf("[%s %s](tg://user?id=%d)", user.User.FirstName, user.User.LastName, val.UserID)

		entries = append(entries, s)
	}

	return strings.Join(entries, "\n")
}
