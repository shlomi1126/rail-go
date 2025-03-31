package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const TOKEN = "1402814094:AAHRSU0i38o83OESiRKKrjCqLqfMxug4kRA"

var chatID int64 = 519614625
var from, to string

type Bot struct {
	*tgbotapi.BotAPI
}

var userState = make(map[int64]string)

func main() {

	botApi, _ := tgbotapi.NewBotAPI(TOKEN)
	bot := &Bot{botApi}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 6000
	updates := bot.GetUpdatesChan(u)

	runChan := make(chan time.Time)
	stopChan := make(chan struct{})
	doneChan := make(chan struct{})
	go scheduler(runChan, stopChan)
	go monthlyTask(bot, runChan, doneChan)

	for update := range updates {
		if update.CallbackQuery != nil {
			handleCallbackQuery(bot, update)
		} else if update.Message != nil {
			handleMessage(bot, update)
		}
	}
}

func getStationSuggestions(query string, maxResults int) map[string]string {
	query = strings.ToLower(query)

	suggestions := make(map[string]string)
	for stationName, station := range STATION_INDEX {
		if strings.HasPrefix(strings.ToLower(stationName), query) {
			suggestions[stationName] = station
			if len(suggestions) >= maxResults {
				break
			}
		}
	}
	return suggestions
}

func (bot *Bot) send(msg tgbotapi.MessageConfig) {
	if _, err := bot.Send(msg); err != nil {
		panic(err)
	}
}
func handleMessage(bot *Bot, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userText := update.Message.Text
	state := userState[chatID]

	log.Printf("Received message from user %d: %s", update.Message.From.ID, userText)
	log.Printf("User %d state: %s", chatID, state)

	switch state {
	case awaitingInput:
		suggestions := getStationSuggestions(userText, 5)
		if len(suggestions) == 0 {
			msg := tgbotapi.NewMessage(chatID, "לא נמצאו תחנות תואמות. נסה להקליד אותיות אחרות.")
			bot.Send(msg)
			return
		}
		// Create inline keyboard with suggestions
		var buttons [][]tgbotapi.InlineKeyboardButton
		for suggestionName, suggestion := range suggestions {
			btn := tgbotapi.NewInlineKeyboardButtonData(suggestionName, suggestion)
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(btn))
		}
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
		msg := tgbotapi.NewMessage(chatID, "אנא בחר תחנת מוצא :")
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
		userState[chatID] = awaitingSelection
		log.Printf("Set user state to 'awaitingSelection' for chatID %d", chatID)

	default:
		// Handle other messages
		switch userText {
		case "🥩":
			msg := tgbotapi.NewMessage(chatID, meat())
			bot.Send(msg)
		case "🧀":
			msg := tgbotapi.NewMessage(chatID, cheese())
			bot.Send(msg)
		case "🚆":
			msg := tgbotapi.NewMessage(chatID, "בחר יעד:")
			msg.ReplyMarkup = trainKeyBoard
			bot.Send(msg)
		default:
			msg := tgbotapi.NewMessage(chatID, "אנא בחר אופציה או הקלד את הפקודה הרצויה.")
			bot.Send(msg)
		}
	}
}

func handleCallbackQuery(bot *Bot, update tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	data := update.CallbackQuery.Data

	// Acknowledge the callback query
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("Error acknowledging callback: %v", err)
	}

	switch data {
	case "home":
		userId := strconv.FormatInt(update.CallbackQuery.From.ID, 10)
		msg := tgbotapi.NewMessage(chatID, getRailSchedule(userId, "4600", "8700"))
		bot.Send(msg)

	case "work":
		userId := strconv.FormatInt(update.CallbackQuery.From.ID, 10)
		msg := tgbotapi.NewMessage(chatID, getRailSchedule(userId, "8700", "4600"))
		bot.Send(msg)

	case "other":
		from, to = "", ""
		msg := tgbotapi.NewMessage(chatID, "אנא הקלד את האותיות הראשונות של שם התחנה.")
		bot.Send(msg)
		userState[chatID] = awaitingInput
		log.Printf("Set user state to 'awaiting_input' for chatID %d", chatID)
	case "search_train":
		userId := strconv.FormatInt(update.CallbackQuery.From.ID, 10)
		msg := tgbotapi.NewMessage(chatID, getRailSchedule(userId, from, to))
		bot.Send(msg)

	default:
		if userState[chatID] == awaitingSelection {
			selectedStation := data
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("בחרת את התחנה: %s", STATIONS[selectedStation]["Heb"]))
			bot.Send(msg)
			if from == "" || to == "" {
				if from == "" {
					from = selectedStation
				}
				if from != "" && to == "" {
					to = selectedStation
				}
				msg = tgbotapi.NewMessage(chatID, "אנא הקלד את האותיות הראשונות של שם תחנת היעד.")
				bot.Send(msg)
				userState[chatID] = awaitingInput
			} else {
				to = selectedStation
				btn := tgbotapi.NewInlineKeyboardButtonData("חפש", "search_train")
				inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{btn})
				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("רכבת מתחנת %v לתחנת %v", STATIONS[from]["Heb"], STATIONS[to]["Heb"]))
				msg.ReplyMarkup = inlineKeyboard
				userState[chatID] = "done"
				bot.Send(msg)

			}
			log.Printf("change user state for chatID %d to %v", chatID, userState[chatID])
		} else {
			msg := tgbotapi.NewMessage(chatID, "פעולה לא מוכרת.")
			bot.Send(msg)
		}
	}
}
