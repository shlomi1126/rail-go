package bot

import (
	"context"
	"fmt"
	"sync"

	"rail-go/internal/config"
	"rail-go/internal/logger"
	"rail-go/internal/train"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Emoji constants
const (
	trainEmoji   = "🚆"
	homeEmoji    = "🏠"
	workEmoji    = "🏢"
	searchEmoji  = "🔍"
	warningEmoji = "⚠️"
	successEmoji = "✅"
	stationEmoji = "🚉"
	steakEmoji   = "🥩"
	cheeseEmoji  = "🧀"
)

type Service struct {
	bot          *tgbotapi.BotAPI
	config       *config.Config
	logger       *logger.Logger
	userState    *sync.Map
	trainService *train.Service
}

func NewService(cfg *config.Config, log *logger.Logger) (*Service, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.Bot.Token)
	if err != nil {
		return nil, err
	}

	bot.Debug = cfg.Bot.Debug

	trainService := train.NewService(cfg, log)

	return &Service{
		bot:          bot,
		config:       cfg,
		logger:       log,
		userState:    &sync.Map{},
		trainService: trainService,
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("bot service starting", "debug_mode", s.bot.Debug)

	if err := s.sendStartupMessage(ctx); err != nil {
		s.logger.Error("failed to send startup message", "error", err)
	}

	updates := s.bot.GetUpdatesChan(tgbotapi.UpdateConfig{
		Timeout: s.config.Bot.Timeout,
	})

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("bot service shutting down")
			return ctx.Err()
		case update := <-updates:
			if err := s.handleUpdate(ctx, update); err != nil {
				s.logger.Error("failed to handle update", "error", err)
			}
		}
	}
}

// sendStartupMessage sends a notification message when the service starts
func (s *Service) sendStartupMessage(ctx context.Context) error {
	if s.config.Bot.AdminChatID == 0 {
		return nil
	}

	startupMsg := fmt.Sprintf("%s Bot Service Started\nVersion: %s\nDebug Mode: %v",
		successEmoji,
		s.config.Version,
		s.bot.Debug)

	return s.SendMessage(ctx, s.config.Bot.AdminChatID, startupMsg)
}

func (s *Service) handleUpdate(ctx context.Context, update tgbotapi.Update) error {
	if update.CallbackQuery != nil {
		s.logger.Info("received callback query",
			"user_id", update.CallbackQuery.From.ID,
			"callback_data", update.CallbackQuery.Data)
		return s.handleCallbackQuery(ctx, update.CallbackQuery)
	}
	if update.Message != nil {
		s.logger.Info("received message",
			"user_id", update.Message.From.ID,
			"text", update.Message.Text)
		return s.handleMessage(ctx, update.Message)
	}
	return nil
}

func (s *Service) handleMessage(ctx context.Context, msg *tgbotapi.Message) error {
	state, _ := s.userState.Load(fmt.Sprintf("state_%d", msg.Chat.ID))
	currentState, ok := state.(string)
	if !ok {
		currentState = "default"
	}

	s.logger.Info("handling message",
		"user_id", msg.From.ID,
		"chat_id", msg.Chat.ID,
		"state", currentState,
		"text", msg.Text)

	switch currentState {
	case "awaitingFromStation", "awaitingToStation":
		return s.handleAwaitingInput(ctx, msg)
	default:
		return s.handleDefaultState(ctx, msg)
	}
}

func (s *Service) handleCallbackQuery(ctx context.Context, query *tgbotapi.CallbackQuery) error {
	callback := tgbotapi.NewCallback(query.ID, "")
	if _, err := s.bot.Request(callback); err != nil {
		s.logger.Error("failed to acknowledge callback", "error", err)
	}

	s.logger.Info("handling callback query",
		"user_id", query.From.ID,
		"chat_id", query.Message.Chat.ID,
		"data", query.Data)

	switch query.Data {
	case "home":
		return s.handleHomeRoute(ctx, query)
	case "work":
		return s.handleWorkRoute(ctx, query)
	case "other":
		return s.handleOtherRoute(ctx, query)
	case "search_train":
		return s.handleSearchTrain(ctx, query)
	default:
		return s.handleStationSelection(ctx, query)
	}
}

func (s *Service) handleHomeRoute(ctx context.Context, query *tgbotapi.CallbackQuery) error {
	s.logger.Info("handling home route request",
		"user_id", query.From.ID,
		"chat_id", query.Message.Chat.ID)

	schedule, err := s.trainService.GetSchedule(ctx, "4600", "8700")
	if err != nil {
		s.logger.Error("failed to get home route schedule", "error", err)
		return fmt.Errorf("failed to get home route schedule: %w", err)
	}

	s.logger.Info("home route schedule retrieved successfully",
		"user_id", query.From.ID,
		"chat_id", query.Message.Chat.ID)

	return s.SendMessages(ctx, query.Message.Chat.ID, schedule)
}

func (s *Service) handleWorkRoute(ctx context.Context, query *tgbotapi.CallbackQuery) error {
	s.logger.Info("handling work route request",
		"user_id", query.From.ID,
		"chat_id", query.Message.Chat.ID)

	schedule, err := s.trainService.GetSchedule(ctx, "8700", "4600")
	if err != nil {
		s.logger.Error("failed to get work route schedule", "error", err)
		return fmt.Errorf("failed to get work route schedule: %w", err)
	}

	s.logger.Info("work route schedule retrieved successfully",
		"user_id", query.From.ID,
		"chat_id", query.Message.Chat.ID)

	return s.SendMessages(ctx, query.Message.Chat.ID, schedule)
}

func (s *Service) handleOtherRoute(ctx context.Context, query *tgbotapi.CallbackQuery) error {
	s.logger.Info("handling other route request",
		"user_id", query.From.ID,
		"chat_id", query.Message.Chat.ID)

	// Initialize the route selection process
	route := make(map[string]string)
	s.userState.Store(fmt.Sprintf("route_%d", query.Message.Chat.ID), route)
	s.userState.Store(fmt.Sprintf("state_%d", query.Message.Chat.ID), "awaitingFromStation")

	s.logger.Info("route selection started",
		"user_id", query.From.ID,
		"chat_id", query.Message.Chat.ID,
		"state", "awaitingFromStation")

	return s.SendMessage(ctx, query.Message.Chat.ID,
		fmt.Sprintf("%s אנא הקלד את האותיות הראשונות של תחנת המוצא.", stationEmoji))
}

func (s *Service) handleSearchTrain(ctx context.Context, query *tgbotapi.CallbackQuery) error {
	s.logger.Info("handling search train request",
		"user_id", query.From.ID,
		"chat_id", query.Message.Chat.ID)

	route, _ := s.userState.Load(fmt.Sprintf("route_%d", query.Message.Chat.ID))
	routeMap, ok := route.(map[string]string)
	if !ok {
		s.logger.Error("invalid route type",
			"user_id", query.From.ID,
			"chat_id", query.Message.Chat.ID,
			"route", route)
		return fmt.Errorf("invalid route type")
	}

	schedule, err := s.trainService.GetSchedule(ctx, routeMap["from"], routeMap["to"])
	if err != nil {
		s.logger.Error("failed to get schedule", "error", err)
		return fmt.Errorf("failed to get schedule: %w", err)
	}

	s.logger.Info("train schedule retrieved successfully",
		"user_id", query.From.ID,
		"chat_id", query.Message.Chat.ID,
		"to", routeMap["to"])

	return s.SendMessages(ctx, query.Message.Chat.ID, schedule)
}

func (s *Service) handleStationSelection(ctx context.Context, query *tgbotapi.CallbackQuery) error {
	s.logger.Info("handling station selection",
		"user_id", query.From.ID,
		"chat_id", query.Message.Chat.ID,
		"station_id", query.Data)

	// Get the current state
	state, _ := s.userState.Load(fmt.Sprintf("state_%d", query.Message.Chat.ID))
	currentState, ok := state.(string)
	if !ok {
		s.logger.Error("invalid state type",
			"user_id", query.From.ID,
			"chat_id", query.Message.Chat.ID,
			"state", state)
		return s.SendMessage(ctx, query.Message.Chat.ID,
			fmt.Sprintf("%s שגיאה במצב. אנא נסה שוב.", warningEmoji))
	}

	s.logger.Info("current state",
		"user_id", query.From.ID,
		"chat_id", query.Message.Chat.ID,
		"state", currentState)

	switch currentState {
	case "awaitingFromStationSelection":
		// Get the route map
		route, _ := s.userState.Load(fmt.Sprintf("route_%d", query.Message.Chat.ID))
		routeMap, ok := route.(map[string]string)
		if !ok {
			s.logger.Error("invalid route type",
				"user_id", query.From.ID,
				"chat_id", query.Message.Chat.ID,
				"route", route)
			return s.SendMessage(ctx, query.Message.Chat.ID,
				fmt.Sprintf("%s שגיאה בבחירת התחנות. אנא נסה שוב.", warningEmoji))
		}
		routeMap["from"] = query.Data
		// Store the updated route map
		s.userState.Store(fmt.Sprintf("route_%d", query.Message.Chat.ID), routeMap)
		// Set the state to awaitingToStation
		s.userState.Store(fmt.Sprintf("state_%d", query.Message.Chat.ID), "awaitingToStation")

		s.logger.Info("from station selected, now awaiting to station input",
			"user_id", query.From.ID,
			"chat_id", query.Message.Chat.ID,
			"station_id", query.Data)

		return s.SendMessage(ctx, query.Message.Chat.ID,
			fmt.Sprintf("%s אנא הקלד את האותיות הראשונות של תחנת היעד.", stationEmoji))
	case "awaitingToStationSelection":
		// Get the route map
		route, _ := s.userState.Load(fmt.Sprintf("route_%d", query.Message.Chat.ID))
		routeMap, ok := route.(map[string]string)
		if !ok {
			s.logger.Error("invalid route type",
				"user_id", query.From.ID,
				"chat_id", query.Message.Chat.ID,
				"route", route)
			return s.SendMessage(ctx, query.Message.Chat.ID,
				fmt.Sprintf("%s שגיאה בבחירת התחנות. אנא נסה שוב.", warningEmoji))
		}
		routeMap["to"] = query.Data
		// Store the updated route map
		s.userState.Store(fmt.Sprintf("route_%d", query.Message.Chat.ID), routeMap)

		s.logger.Info("to station selected",
			"user_id", query.From.ID,
			"chat_id", query.Message.Chat.ID,
			"from_station", routeMap["from"],
			"to_station", routeMap["to"])

		btn := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("חפש %s", searchEmoji), "search_train")
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{btn})
		msg := tgbotapi.NewMessage(query.Message.Chat.ID,
			fmt.Sprintf("%s רכבת מתחנת %v לתחנת %v",
				trainEmoji,
				s.trainService.GetStationName(routeMap["from"]),
				s.trainService.GetStationName(routeMap["to"])))
		msg.ReplyMarkup = inlineKeyboard
		_, err := s.bot.Send(msg)
		return err
	default:
		s.logger.Error("unknown state",
			"user_id", query.From.ID,
			"chat_id", query.Message.Chat.ID,
			"state", currentState)
		return s.SendMessage(ctx, query.Message.Chat.ID,
			fmt.Sprintf("%s מצב לא ידוע. אנא נסה שוב.", warningEmoji))
	}
}

func (s *Service) handleAwaitingInput(ctx context.Context, msg *tgbotapi.Message) error {
	s.logger.Info("handling awaiting input",
		"user_id", msg.From.ID,
		"chat_id", msg.Chat.ID,
		"input", msg.Text)

	// Get the current state
	state, _ := s.userState.Load(fmt.Sprintf("state_%d", msg.Chat.ID))
	currentState, ok := state.(string)
	if !ok {
		s.logger.Error("invalid state type",
			"user_id", msg.From.ID,
			"chat_id", msg.Chat.ID,
			"state", state)
		return s.SendMessage(ctx, msg.Chat.ID,
			fmt.Sprintf("%s שגיאה במצב. אנא נסה שוב.", warningEmoji))
	}

	s.logger.Info("current state",
		"user_id", msg.From.ID,
		"chat_id", msg.Chat.ID,
		"state", currentState)

	switch currentState {
	case "awaitingFromStation":
		suggestions := s.trainService.GetStationSuggestions(msg.Text)
		if len(suggestions) == 0 {
			s.logger.Info("no stations found for input",
				"user_id", msg.From.ID,
				"chat_id", msg.Chat.ID,
				"input", msg.Text)
			return s.SendMessage(ctx, msg.Chat.ID,
				fmt.Sprintf("%s לא נמצאו תחנות תואמות. נסה להקליד אותיות אחרות.", warningEmoji))
		}

		var buttons [][]tgbotapi.InlineKeyboardButton
		for stationName, stationID := range suggestions {
			btn := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s %s", stationEmoji, stationName), stationID)
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(btn))
		}

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
		message := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("%s אנא בחר תחנת מוצא:", stationEmoji))
		message.ReplyMarkup = inlineKeyboard
		_, err := s.bot.Send(message)
		if err != nil {
			s.logger.Error("failed to send message", "error", err)
			return fmt.Errorf("failed to send message: %w", err)
		}

		// Set the state to awaitingFromStationSelection
		s.userState.Store(fmt.Sprintf("state_%d", msg.Chat.ID), "awaitingFromStationSelection")
		return nil
	case "awaitingToStation":
		suggestions := s.trainService.GetStationSuggestions(msg.Text)
		if len(suggestions) == 0 {
			s.logger.Info("no stations found for input",
				"user_id", msg.From.ID,
				"chat_id", msg.Chat.ID,
				"input", msg.Text)
			return s.SendMessage(ctx, msg.Chat.ID,
				fmt.Sprintf("%s לא נמצאו תחנות תואמות. נסה להקליד אותיות אחרות.", warningEmoji))
		}

		var buttons [][]tgbotapi.InlineKeyboardButton
		for stationName, stationID := range suggestions {
			btn := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s %s", stationEmoji, stationName), stationID)
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(btn))
		}

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
		message := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("%s אנא בחר תחנת יעד:", stationEmoji))
		message.ReplyMarkup = inlineKeyboard
		_, err := s.bot.Send(message)
		if err != nil {
			s.logger.Error("failed to send message", "error", err)
			return fmt.Errorf("failed to send message: %w", err)
		}

		// Set the state to awaitingToStationSelection
		s.userState.Store(fmt.Sprintf("state_%d", msg.Chat.ID), "awaitingToStationSelection")
		return nil
	default:
		s.logger.Error("unknown state",
			"user_id", msg.From.ID,
			"chat_id", msg.Chat.ID,
			"state", currentState)
		return s.SendMessage(ctx, msg.Chat.ID,
			fmt.Sprintf("%s מצב לא ידוע. אנא נסה שוב.", warningEmoji))
	}
}

func (s *Service) handleDefaultState(ctx context.Context, msg *tgbotapi.Message) error {
	s.logger.Info("handling default state",
		"user_id", msg.From.ID,
		"chat_id", msg.Chat.ID,
		"text", msg.Text)

	switch msg.Text {
	case trainEmoji:
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("בית %s", homeEmoji), "home"),
				tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("עבודה %s", workEmoji), "work"),
				tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("אחר %s", searchEmoji), "other"),
			),
		)
		msg := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("%s בחר יעד:", trainEmoji))
		msg.ReplyMarkup = keyboard
		_, err := s.bot.Send(msg)
		return err
	case steakEmoji:
		return s.SendMessage(ctx, msg.Chat.ID, meat())
	case cheeseEmoji:
		return s.SendMessage(ctx, msg.Chat.ID, cheese())
	default:
		return s.SendMessage(ctx, msg.Chat.ID,
			fmt.Sprintf("%s אנא בחר אופציה או הקלד את הפקודה הרצויה %s", warningEmoji, trainEmoji))
	}
}

func (s *Service) getStationSuggestions(query string) map[string]string {
	// This is a placeholder. You should implement the actual station lookup logic
	suggestions := make(map[string]string)
	suggestions["תל אביב - סבידור מרכז"] = "4600"
	suggestions["הרצליה"] = "8700"
	return suggestions
}

func (s *Service) SendMessage(ctx context.Context, chatID int64, message string) error {
	msg := tgbotapi.NewMessage(chatID, message)
	_, err := s.bot.Send(msg)
	return err
}

func (s *Service) SendMessages(ctx context.Context, chatID int64, messages []string) error {
	for i, message := range messages {
		msg := tgbotapi.NewMessage(chatID, message)
		if i > 0 {
			msg.Text = fmt.Sprintf("(חלק %d מתוך %d)\n%s", i+1, len(messages), message)
		}
		if _, err := s.bot.Send(msg); err != nil {
			return err
		}
	}
	return nil
}
