package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/harshyadavone/anonymous_chat/store"
	"github.com/harshyadavone/tgx"
	"github.com/harshyadavone/tgx/models"
	"github.com/harshyadavone/tgx/pkg/logger"
)

var (
	bot       *tgx.Bot
	userStore *store.DynamoDBStore
)

func init() {
	token := os.Getenv("BOT_TOKEN")
	tableName := os.Getenv("DYNAMODB_TABLE")
	if token == "" || tableName == "" {
		log.Fatal("FATAL: BOT_TOKEN and DYNAMODB_TABLE environment variables must be set")
	}

	logger := logger.NewDefaultLogger(logger.INFO)

	var err error
	userStore, err = store.New(context.Background(), tableName)
	if err != nil {
		log.Fatalf("FATAL: failed to initialize DynamoDB store: %v", err)
	}

	bot = tgx.NewBot(token, "", logger)

	bot.OnError(func(ctx *tgx.Context, err error) {
		log.Printf("ERROR: An error occurred in an update: %v", err)
		ctx.Reply(MessageErrSomethingWentWrong)
	})

	bot.SetMyCommands(Commands)

	bot.OnCommand("start", func(ctx *tgx.Context) error {
		log.Println("LOG: Handling /start command")
		req := &tgx.SendMessageRequest{
			ChatId:      ctx.ChatID,
			Text:        "👋 Welcome! Chat anonymously with random people here. Type /connect to start or /help for commands!",
			ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: inlineKeyboardButton},
		}
		return bot.SendMessageWithOpts(req)
	})

	bot.OnCommand("help", func(ctx *tgx.Context) error {
		return ctx.Reply(MessageQuickGuide)
	})

	bot.OnCommand("connect", func(ctx *tgx.Context) error {
		return HandleConnect(bot, ctx.ChatID)
	})

	bot.OnCommand("stop", func(ctx *tgx.Context) error {
		return HandleStop(bot, ctx.ChatID)
	})

	bot.OnCommand("next", func(ctx *tgx.Context) error {
		return HandleNext(bot, ctx.ChatID)
	})

	bot.OnCommand("status", func(ctx *tgx.Context) error {
		return HandleStatus(bot, ctx.ChatID)
	})

	bot.OnCallback("connect", func(ctx *tgx.CallbackContext) error {
		err := HandleConnect(bot, ctx.GetChatID())
		if err != nil {
			log.Printf("ERROR: HandleConnect from callback failed: %v", err)
		}
		return ctx.AnswerCallback(&tgx.CallbackAnswerOptions{})
	})

	bot.OnCallback("status", func(ctx *tgx.CallbackContext) error {
		err := HandleStatus(bot, ctx.GetChatID())
		if err != nil {
			log.Printf("ERROR: HandleStatus from callback failed: %v", err)
		}
		return ctx.AnswerCallback(&tgx.CallbackAnswerOptions{})
	})

	bot.OnMessage("Text", func(ctx *tgx.Context) error {
		partnerChatId, errMsg := CheckAndGetPartner(ctx.ChatID)
		if errMsg != "" {
			return ctx.Reply(errMsg)
		}
		return bot.SendMessage(partnerChatId, ctx.Text)
	})

	bot.OnMessage("Animation", func(ctx *tgx.Context) error {
		partnerChatId, errMsg := CheckAndGetPartner(ctx.ChatID)
		if errMsg != "" {
			return ctx.Reply(errMsg)
		}
		req := &tgx.SendAnimationRequest{
			Animation:        ctx.Animation.FileId,
			BaseMediaRequest: tgx.BaseMediaRequest{ChatId: partnerChatId},
		}
		return bot.SendAnimation(req)
	})

	bot.OnMessage("Photo", func(ctx *tgx.Context) error {
		partnerChatId, errMsg := CheckAndGetPartner(ctx.ChatID)
		if errMsg != "" {
			return ctx.Reply(errMsg)
		}
		req := &tgx.SendPhotoRequest{
			Photo:            ctx.Photo[0].FileId,
			BaseMediaRequest: tgx.BaseMediaRequest{ChatId: partnerChatId},
		}
		return bot.SendPhoto(req)
	})

	bot.OnMessage("Voice", func(ctx *tgx.Context) error {
		partnerChatId, errMsg := CheckAndGetPartner(ctx.ChatID)
		if errMsg != "" {
			return ctx.Reply(errMsg)
		}
		req := &tgx.SendVoiceRequest{
			Voice:            ctx.Voice.FileId,
			BaseMediaRequest: tgx.BaseMediaRequest{ChatId: partnerChatId},
		}
		return bot.SendVoice(req)
	})

	bot.OnMessage("Document", func(ctx *tgx.Context) error {
		partnerChatId, errMsg := CheckAndGetPartner(ctx.ChatID)
		if errMsg != "" {
			return ctx.Reply(errMsg)
		}
		req := &tgx.SendDocumentRequest{
			Document:         ctx.Document.FileId,
			BaseMediaRequest: tgx.BaseMediaRequest{ChatId: partnerChatId},
		}
		return bot.SendDocument(req)
	})

	bot.OnMessage("Video", func(ctx *tgx.Context) error {
		partnerChatId, errMsg := CheckAndGetPartner(ctx.ChatID)
		if errMsg != "" {
			return ctx.Reply(errMsg)
		}
		req := &tgx.SendVideoRequest{
			Video:            ctx.Video.FileId,
			BaseMediaRequest: tgx.BaseMediaRequest{ChatId: partnerChatId},
		}
		return bot.SendVideo(req)
	})

	bot.OnMessage("Sticker", func(ctx *tgx.Context) error {
		partnerChatId, errMsg := CheckAndGetPartner(ctx.ChatID)
		if errMsg != "" {
			return ctx.Reply(errMsg)
		}
		req := &tgx.SendStickerRequest{
			Sticker:          ctx.Sticker.FileId,
			BaseMediaRequest: tgx.BaseMediaRequest{ChatId: partnerChatId},
		}
		return bot.SendSticker(req)
	})

	log.Println("--- BOT INITIALIZED SUCCESSFULLY ---")
}

func HandleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Handler invoked! Request Body: %s", req.Body)
	httpRequest, err := http.NewRequest("POST", "/", strings.NewReader(req.Body))
	if err != nil {
		log.Printf("ERROR: Could not create new HTTP request: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()
	bot.HandleWebhook(responseRecorder, httpRequest)
	return events.APIGatewayProxyResponse{StatusCode: responseRecorder.Code, Body: responseRecorder.Body.String()}, nil
}

func main() {
	lambda.Start(HandleRequest)
}

func HandleConnect(b *tgx.Bot, chatId int64) error {
	log.Printf("LOG: HandleConnect called for ChatID: %d", chatId)
	ctx := context.Background()

	user, err := userStore.GetUser(ctx, chatId)
	if err != nil {
		log.Printf("LOG: User %d not found in DB, creating new user object.", chatId)
		user = &store.User{ChatId: chatId}
	}

	if user.IsConnected {
		log.Printf("LOG: User %d is already connected. Aborting connect.", chatId)
		return b.SendMessage(chatId, MessageAlreadyConnected)
	}

	partner, err := userStore.FindAndConnectPartner(ctx, user)
	if err != nil {
		log.Printf("ERROR: FindAndConnectPartner failed for %d: %v", chatId, err)
		return b.SendMessage(chatId, MessageErrSomethingWentWrong)
	}

	if partner != nil {
		log.Printf("LOG: Match found! %d is now connected with %d.", user.ChatId, partner.ChatId)
		b.SendMessage(user.ChatId, MessageConnected)
		b.SendMessage(partner.ChatId, MessageConnected)
		return nil // Success!
	}

	log.Printf("LOG: No partner found for %d. Attempting to put user in queue.", chatId)
	user.IsConnecting = 1
	if err := userStore.UpdateUser(ctx, user); err != nil {
		log.Printf("ERROR: Failed to put user %d into queue: %v", chatId, err)
		return b.SendMessage(chatId, MessageErrSomethingWentWrong)
	}

	log.Printf("LOG: Successfully put user %d in queue.", chatId)
	return b.SendMessage(chatId, MessageLookingForPartner)
}

func HandleStop(b *tgx.Bot, chatId int64) error {
	log.Printf("LOG: HandleStop called for ChatID: %d", chatId)
	ctx := context.Background()

	user, err := userStore.GetUser(ctx, chatId)
	if err != nil || (!user.IsConnected && user.IsConnecting == 0) {
		log.Printf("LOG: User %d tried to stop but was not in a chat or queue.", chatId)
		return b.SendMessage(chatId, MessageConnectWithSomeoneFirst)
	}

	if user.IsConnected {
		log.Printf("LOG: User %d is disconnecting from partner %d.", chatId, user.Partner)
		partner, err := userStore.GetUser(ctx, user.Partner)
		if err == nil {
			partner.IsConnected = false
			partner.Partner = 0
			userStore.UpdateUser(ctx, partner)
			b.SendMessage(partner.ChatId, MessagePartnerLeftChat)
		} else {
			log.Printf("WARN: Could not find partner %d to notify about disconnection for user %d.", user.Partner, chatId)
		}
	}

	log.Printf("LOG: Resetting status for user %d.", chatId)
	user.IsConnected = false
	user.IsConnecting = 0
	user.Partner = 0
	if err := userStore.UpdateUser(ctx, user); err != nil {
		log.Printf("ERROR: Failed to update user %d on stop: %v", chatId, err)
		return b.SendMessage(chatId, MessageErrSomethingWentWrong)
	}

	return b.SendMessage(chatId, MessageChatEnded)
}

func HandleNext(b *tgx.Bot, chatId int64) error {
	log.Printf("LOG: HandleNext called for ChatID: %d", chatId)
	HandleStop(b, chatId)
	return HandleConnect(b, chatId)
}

func HandleStatus(b *tgx.Bot, chatId int64) error {
	log.Printf("LOG: HandleStatus called for ChatID: %d", chatId)
	user, err := userStore.GetUser(context.Background(), chatId)
	if err != nil {
		return b.SendMessage(chatId, MessageNotConnectedStatus)
	}
	if user.IsConnected {
		return b.SendMessage(chatId, MessageCurrentlyChatting)
	}
	if user.IsConnecting == 1 {
		return b.SendMessage(chatId, MessageInWaitingList)
	}
	return b.SendMessage(chatId, MessageNotConnectedStatus)
}

func CheckAndGetPartner(chatId int64) (int64, string) {
	log.Printf("LOG: Checking for partner for ChatID %d", chatId)

	user, err := userStore.GetUser(context.Background(), chatId)
	if err != nil {
		log.Printf("WARN: User %d not found in DB for partner check.", chatId)
		return 0, MessageNotConnected
	}
	if !user.IsConnected || user.Partner == 0 {
		log.Printf("LOG: User %d is not currently connected to a partner.", chatId)
		return 0, MessageNotConnected
	}

	log.Printf("LOG: Found partner %d for user %d.", user.Partner, chatId)
	return user.Partner, ""
}

