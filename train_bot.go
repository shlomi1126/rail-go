package main

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

var trainKeyBoard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("home", "home"),
		tgbotapi.NewInlineKeyboardButtonData("work", "work"),
		tgbotapi.NewInlineKeyboardButtonData("other", "other"),
	),
)

var trainStations = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("home", "home"),
		tgbotapi.NewInlineKeyboardButtonData("work", "work"),
		tgbotapi.NewInlineKeyboardButtonData("other", "other"),
	),
)
