package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/harshyadavone/anonymous_chat/store"
	"github.com/harshyadavone/tgx"
	"github.com/harshyadavone/tgx/models"
	"github.com/harshyadavone/tgx/pkg/logger"
)

type App struct {
	bot       *tgx.Bot
	userStore *store.UserStore
}

func main() {
	token := os.Getenv("BOT_TOKEN")
	webhookURL := os.Getenv("WEBHOOK_URL")

	logger := logger.NewDefaultLogger(logger.DEBUG)

	bot := tgx.NewBot(token, webhookURL, logger)

	app := &App{
		bot: bot,
		userStore: &store.UserStore{
			Users: make(map[int64]*store.User),
		},
	}

	logger.Info("Starting the bot...")

	bot.OnError(func(ctx *tgx.Context, err error) {
		payload := &tgx.SendMessageRequest{
			ChatId: ctx.ChatID,
			Text:   MessageErrSomethingWentWrong,
		}
		ctx.ReplyWithOpts(payload)
	})

	bot.OnCommand("start", func(ctx *tgx.Context) error {
		req := &tgx.SendMessageRequest{
			ChatId: ctx.ChatID,
			Text:   "👋 Welcome! Chat anonymously with random people here. Type /connect to start or /help for commands!",
			ReplyMarkup: models.InlineKeyboardMarkup{
				InlineKeyboard: inlineKeyboardButton,
			},
		}
		return bot.SendMessageWithOpts(req)
	})

	bot.OnCommand("help", func(ctx *tgx.Context) error {
		return ctx.Reply(MessageQuickGuide)
	})

	bot.OnCommand("connect", func(ctx *tgx.Context) error {
		return app.HandleConnect(ctx.ChatID)
	})

	bot.OnCommand("status", func(ctx *tgx.Context) error {
		return app.HandleStatus(ctx.ChatID)
	})

	bot.OnCommand("next", func(ctx *tgx.Context) error {
		return app.HandleNext(ctx.ChatID)
	})

	bot.OnCommand("stop", func(ctx *tgx.Context) error {
		return app.HandleStop(ctx.ChatID)
	})

	bot.OnCommand("gender", func(ctx *tgx.Context) error {
		return app.HandleGender(ctx)
	})

	bot.OnCommand("interests", func(ctx *tgx.Context) error {
		return app.HandleInterests(ctx)
	})

	bot.OnCommand("block", func(ctx *tgx.Context) error {
		return app.HandleBlock(ctx.ChatID)
	})

	bot.OnCommand("report", func(ctx *tgx.Context) error {
		return app.HandleReport(ctx.ChatID)
	})

	bot.SetMyCommands(Commands)

	// Specific callbacks that are not part of a conversation
	bot.OnCallback("connect", func(ctx *tgx.CallbackContext) error {
		if err := app.HandleConnect(ctx.GetChatID()); err != nil {
			return err
		}
		return ctx.AnswerCallback(&tgx.CallbackAnswerOptions{})
	})
	bot.OnCallback("status", func(ctx *tgx.CallbackContext) error {
		if err := app.HandleStatus(ctx.GetChatID()); err != nil {
			return err
		}
		return ctx.AnswerCallback(&tgx.CallbackAnswerOptions{})
	})

	// Generic callback handler for conversations
	bot.OnCallback(func(ctx *tgx.CallbackContext) error {
		return app.handleCallback(ctx)
	})

	// Generic message handler
	messageHandler := func(ctx *tgx.Context) error {
		return app.handleMediaMessage(ctx)
	}
	bot.OnMessage("Text", messageHandler)
	bot.OnMessage("Animation", messageHandler)
	bot.OnMessage("Photo", messageHandler)
	bot.OnMessage("Voice", messageHandler)
	bot.OnMessage("Document", messageHandler)
	bot.OnMessage("Video", messageHandler)
	bot.OnMessage("Sticker", messageHandler)

	if err := bot.SetWebhook(); err != nil {
		log.Fatal("Failed to set webhook:", err)
	}

	logger.Info("Starting server on :8080")
	http.HandleFunc("/webhook", bot.HandleWebhook)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server error:", err)
	}
}
func (app *App) HandleBlock(chatId int64) error {
	user, exists := app.userStore.GetUser(chatId)
	if !exists || !user.IsConnected {
		return app.bot.SendMessage(chatId, MessageNotInChat)
	}

	partnerId := user.Partner
	if user.BlockedUsers == nil {
		user.BlockedUsers = make(map[int64]struct{})
	}
	user.BlockedUsers[partnerId] = struct{}{}

	app.bot.SendMessage(chatId, MessageUserBlocked)
	return app.HandleStop(chatId)
}

func (app *App) HandleReport(chatId int64) error {
	user, exists := app.userStore.GetUser(chatId)
	if !exists || !user.IsConnected {
		return app.bot.SendMessage(chatId, MessageNotInChat)
	}

	log.Printf("REPORT: User %d reported user %d", user.ChatId, user.Partner)

	app.bot.SendMessage(chatId, MessageReportThanks)
	return app.HandleStop(chatId)
}
func (app *App) HandleGender(ctx *tgx.Context) error {
	user, exists := app.userStore.GetUser(ctx.ChatID)
	if !exists {
		user = &store.User{ChatId: ctx.ChatID}
		app.userStore.AddUser(user)
	}

	user.State = StateAwaitingGender
	req := &tgx.SendMessageRequest{
		ChatId:      ctx.ChatID,
		Text:        "Please select your gender.",
		ReplyMarkup: genderKeyboard,
	}
	return app.bot.SendMessageWithOpts(req)
}

func (app *App) HandleInterests(ctx *tgx.Context) error {
	user, exists := app.userStore.GetUser(ctx.ChatID)
	if !exists {
		user = &store.User{ChatId: ctx.ChatID}
		app.userStore.AddUser(user)
	}

	user.State = StateAwaitingInterests
	return ctx.Reply("Please send me a list of your interests, separated by commas (e.g., movies, music, coding).")
}

func (app *App) handleCallback(ctx *tgx.CallbackContext) error {
	user, exists := app.userStore.GetUser(ctx.GetChatID())
	if !exists {
		return ctx.AnswerCallback(&tgx.CallbackAnswerOptions{Text: "Please type /start first."})
	}

	// This is part of the gender selection conversation
	if user.State == StateAwaitingGender && strings.HasPrefix(ctx.Data, "gender_") {
		gender := strings.TrimPrefix(ctx.Data, "gender_")
		user.Gender = gender
		user.State = StateAwaitingPreference

		req := &tgx.EditMessageTextRequest{
			ChatId:      ctx.GetChatID(),
			MessageId:   ctx.GetMessageId(),
			Text:        "Great! Now, who would you like to connect with?",
			ReplyMarkup: &preferenceKeyboard,
		}
		app.bot.EditMessageText(req)
		return ctx.AnswerCallback(&tgx.CallbackAnswerOptions{})
	}

	if user.State == StateAwaitingPreference && strings.HasPrefix(ctx.Data, "pref_") {
		preference := strings.TrimPrefix(ctx.Data, "pref_")
		user.Preference = preference
		user.State = StateDefault

		req := &tgx.EditMessageTextRequest{
			ChatId:    ctx.GetChatID(),
			MessageId: ctx.GetMessageId(),
			Text:      "Your preferences have been saved! Type /connect to find a partner.",
		}
		app.bot.EditMessageText(req)
		return ctx.AnswerCallback(&tgx.CallbackAnswerOptions{})
	}

	return ctx.AnswerCallback(&tgx.CallbackAnswerOptions{})
}

func (app *App) HandleConnect(chatId int64) error {
	user, exists := app.userStore.GetUser(chatId)
	if !exists {
		user = &store.User{ChatId: chatId}
		app.userStore.AddUser(user)
	}

	if user.IsConnected {
		return app.bot.SendMessage(chatId, MessageAlreadyConnected)
	}

	if user.Gender == "" || user.Preference == "" {
		return app.bot.SendMessage(chatId, "Please set your gender and partner preference using /gender before connecting.")
	}

	// Try to find a match
	if partner, found := app.userStore.FindMatch(user); found {
		// Match found!
		user.IsConnected = true
		user.IsConnecting = false
		user.Partner = partner.ChatId

		partner.IsConnected = true
		partner.IsConnecting = false
		partner.Partner = user.ChatId

		if err := app.bot.SendMessage(chatId, MessageConnected); err != nil {
			return err
		}
		return app.bot.SendMessage(partner.ChatId, MessageConnected)
	}

	// No match found, put user in waiting state
	user.IsConnecting = true
	return app.bot.SendMessage(chatId, MessageLookingForPartner)
}

func (app *App) HandleNext(chatId int64) error {
	user, exists := app.userStore.GetUser(chatId)
	if !exists || !user.IsConnected {
		return app.bot.SendMessage(chatId, "You need to be in a chat to use /next. Use /stop to leave the waiting queue.")
	}

	if err := app.HandleStop(chatId); err != nil {
		return err
	}

	return app.HandleConnect(chatId)
}

func (app *App) HandleStop(chatId int64) error {
	user, exists := app.userStore.GetUser(chatId)
	if !exists || (!user.IsConnected && !user.IsConnecting) {
		return app.bot.SendMessage(chatId, MessageConnectWithSomeoneFirst)
	}

	if user.IsConnecting {
		user.IsConnecting = false
		return app.bot.SendMessage(chatId, "You have been removed from the connection queue.")
	}

	if user.IsConnected {
		partner, partnerExists := app.userStore.GetUser(user.Partner)
		if partnerExists {
			partner.IsConnected = false
			partner.Partner = 0
			app.bot.SendMessage(partner.ChatId, MessagePartnerLeftChat)
		}

		user.IsConnected = false
		user.Partner = 0
		return app.bot.SendMessage(chatId, MessageChatEnded)
	}

	return app.bot.SendMessage(chatId, MessageConnectWithSomeoneFirst)
}

func (app *App) HandleStatus(chatId int64) error {
	user, exists := app.userStore.GetUser(chatId)
	if !exists {
		return app.bot.SendMessage(chatId, MessageNotConnectedStatus)
	}

	if user.IsConnected {
		return app.bot.SendMessage(chatId, MessageCurrentlyChatting)
	}

	if user.IsConnecting {
		return app.bot.SendMessage(chatId, MessageInWaitingList)
	}

	return app.bot.SendMessage(chatId, MessageNotConnectedStatus)
}

func (app *App) CheckAndGetPartner(chatId int64) (*store.User, string) {
	user, exists := app.userStore.GetUser(chatId)
	if !exists || !user.IsConnected {
		return nil, MessageNotConnected
	}
	if user.State != StateDefault {
		return nil, "You are in the middle of a conversation with the bot. Please complete it first."
	}

	partner, exists := app.userStore.GetUser(user.Partner)
	if !exists {
		return nil, MessagePartnerNotAvailable
	}
	return partner, ""
}

func (app *App) handleMediaMessage(ctx *tgx.Context) error {
	user, exists := app.userStore.GetUser(ctx.ChatID)
	if exists {
		// Handle conversation states
		switch user.State {
		case StateAwaitingGender, StateAwaitingPreference:
			return ctx.Reply("Please use the buttons to make your selection.")
		case StateAwaitingInterests:
			// Parse and save interests
			interests := strings.Split(ctx.Text, ",")
			for i, interest := range interests {
				interests[i] = strings.TrimSpace(interest)
			}
			user.Interests = interests
			user.State = StateDefault
			return ctx.Reply("Your interests have been saved!")
		}
	}

	partner, errMsg := app.CheckAndGetPartner(ctx.ChatID)
	if errMsg != "" {
		return ctx.Reply(errMsg)
	}

	// Send a "typing" action to the partner to make the chat feel more responsive.
	app.bot.SendChatAction(&tgx.SendChatActionRequest{
		ChatId: partner.ChatId,
		Action: "typing",
	})

	switch {
	case ctx.Text != "":
		return app.bot.SendMessage(partner.ChatId, ctx.Text)
	case ctx.Animation != nil:
		req := &tgx.SendAnimationRequest{
			Animation: ctx.Animation.FileId,
			BaseMediaRequest: tgx.BaseMediaRequest{
				ChatId: partner.ChatId,
			},
		}
		return app.bot.SendAnimation(req)
	case ctx.Photo != nil:
		req := &tgx.SendPhotoRequest{
			Photo: ctx.Photo[0].FileId,
			BaseMediaRequest: tgx.BaseMediaRequest{
				ChatId: partner.ChatId,
			},
		}
		return app.bot.SendPhoto(req)
	case ctx.Voice != nil:
		req := &tgx.SendVoiceRequest{
			Voice: ctx.Voice.FileId,
			BaseMediaRequest: tgx.BaseMediaRequest{
				ChatId: partner.ChatId,
			},
		}
		return app.bot.SendVoice(req)
	case ctx.Document != nil:
		req := &tgx.SendDocumentRequest{
			Document: ctx.Document.FileId,
			BaseMediaRequest: tgx.BaseMediaRequest{
				ChatId: partner.ChatId,
			},
		}
		return app.bot.SendDocument(req)
	case ctx.Video != nil:
		req := &tgx.SendVideoRequest{
			Video: ctx.Video.FileId,
			BaseMediaRequest: tgx.BaseMediaRequest{
				ChatId: partner.ChatId,
			},
		}
		return app.bot.SendVideo(req)
	case ctx.Sticker != nil:
		req := &tgx.SendStickerRequest{
			Sticker: ctx.Sticker.FileId,
			BaseMediaRequest: tgx.BaseMediaRequest{
				ChatId: partner.ChatId,
			},
		}
		return app.bot.SendSticker(req)
	}

	return nil
}
