package main

import (
	"log"
	"net/http"
	"os"

	"github.com/harshyadavone/anonymous_chat/queue"
	"github.com/harshyadavone/anonymous_chat/store"

	"github.com/harshyadavone/tgx"
	"fmt"
	"strings"

	"github.com/harshyadavone/tgx/models"
	"github.com/harshyadavone/tgx/pkg/logger"
)

const (
	MaxSelectedInterests   = 3
	MaxInterestSearchDepth = 5 // How many users to check from the queue for an interest match

	CallbackInterestViewEdit      = "interest_view_edit"
	CallbackInterestSelectPrefix  = "interest_select_"
	CallbackInterestClear         = "interest_clear"
	CallbackInterestCancel        = "interest_cancel"
	CallbackInterestDoneEditing   = "interest_done_editing"

	MessageManageInterests        = "Manage your anonymous interests. Selecting interests can help us find more compatible chat partners for you. You can select up to 3 interests. Your interests are not directly shared with your chat partners."
	MessageInterestsCleared       = "Your anonymous interests have been cleared."
	MessageInterestsNoChanges     = "No changes made to your interests."
	MessageMaxInterestsReached    = "Max 3 interests allowed. Please deselect one if you wish to choose another."
	MessageInterestsUpdatedPrefix = "Your interests have been updated to: "
	MessageNoInterestsSelected  = "You currently have no interests selected."
	MessageConnectedWithPotentialSharedInterest = "✨ You’re connected! Say hi to your chat partner. You might discover you have something in common! Type /stop if you’d like to end the chat."
)

var PredefinedInterests = []string{
	"Movies & TV", "Music", "Gaming", "Books & Writing",
	"Sports & Fitness", "Tech & Science", "Travel & Outdoors", "Food & Cooking",
}

// Assuming Commands is defined in this file as it's not in utils.go for this task
var Commands = []models.BotCommand{
	{Command: "/start", Description: "Get started with the bot"},
	{Command: "/connect", Description: "Find someone to chat with"},
	{Command: "/stop", Description: "End the current chat session"},
	{Command: "/status", Description: "Check your chat connection status"},
	{Command: "/help", Description: "Get a quick guide on how to use the bot"},
	{Command: "/next", Description: "Find a new chat partner (Not Implemented)"},
	{Command: "/myinterests", Description: "Set/view your anonymous chat interests."},
}

var userStore = &store.UserStore{
	Users: make(map[int64]*store.User),
}

var waitingQueue = queue.NewQueue()

func main() {
	token := os.Getenv("BOT_TOKEN")
	webhookURL := os.Getenv("WEBHOOK_URL")

	logger := logger.NewDefaultLogger(logger.DEBUG)

	bot := tgx.NewBot(token, webhookURL, logger)

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
		return HandleConnect(bot, ctx.ChatID)
	})

	bot.OnCommand("status", func(ctx *tgx.Context) error {
		return HandleStatus(bot, ctx.ChatID)
	})

	bot.OnCommand("next", func(ctx *tgx.Context) error {
		return ctx.Reply(MessageFeatureNotImplemented)
	})

	bot.OnCommand("stop", func(ctx *tgx.Context) error {
		return HandleStop(bot, ctx.ChatID)
	})

	bot.OnCommand("myinterests", HandleMyInterestsCommand)

	bot.SetMyCommands(Commands)

	bot.OnCallback(CallbackInterestViewEdit, HandleInterestViewEditCallback)
	bot.OnCallbackPrefix(CallbackInterestSelectPrefix, HandleInterestSelectCallback)
	bot.OnCallback(CallbackInterestClear, HandleInterestClearCallback)
	bot.OnCallback(CallbackInterestCancel, HandleInterestCancelCallback)
	bot.OnCallback(CallbackInterestDoneEditing, HandleInterestDoneEditingCallback)

	bot.OnCallback("connect", func(ctx *tgx.CallbackContext) error {
		if err := HandleConnect(bot, ctx.GetChatID()); err != nil {
			return err
		}
		return ctx.AnswerCallback(&tgx.CallbackAnswerOptions{})
	})

	bot.OnCallback("status", func(ctx *tgx.CallbackContext) error {
		if err := HandleStatus(bot, ctx.GetChatID()); err != nil {
			return err
		}
		return ctx.AnswerCallback(&tgx.CallbackAnswerOptions{})
	})

	bot.OnMessage("Text", func(ctx *tgx.Context) error {

		partner, errMsg := CheckAndGetPartner(ctx.ChatID, userStore)
		if errMsg != "" {
			return ctx.Reply(errMsg)
		}

		return bot.SendMessage(partner.ChatId, ctx.Text)
	})

	bot.OnMessage("Animation", func(ctx *tgx.Context) error {

		partner, errMsg := CheckAndGetPartner(ctx.ChatID, userStore)
		if errMsg != "" {
			return ctx.Reply(errMsg)
		}

		req := &tgx.SendAnimationRequest{
			Animation: ctx.Animation.FileId,
			BaseMediaRequest: tgx.BaseMediaRequest{
				ChatId: partner.ChatId,
			},
		}
		return bot.SendAnimation(req)
	})

	bot.OnMessage("Photo", func(ctx *tgx.Context) error {

		partner, errMsg := CheckAndGetPartner(ctx.ChatID, userStore)
		if errMsg != "" {
			return ctx.Reply(errMsg)
		}

		req := &tgx.SendPhotoRequest{
			Photo: ctx.Photo[0].FileId,
			BaseMediaRequest: tgx.BaseMediaRequest{
				ChatId: partner.ChatId,
			},
		}
		return bot.SendPhoto(req)
	})

	bot.OnMessage("Voice", func(ctx *tgx.Context) error {

		partner, errMsg := CheckAndGetPartner(ctx.ChatID, userStore)
		if errMsg != "" {
			return ctx.Reply(errMsg)
		}

		req := &tgx.SendVoiceRequest{
			Voice: ctx.Voice.FileId,
			BaseMediaRequest: tgx.BaseMediaRequest{
				ChatId: partner.ChatId,
			},
		}
		return bot.SendVoice(req)
	})

	bot.OnMessage("Document", func(ctx *tgx.Context) error {

		partner, errMsg := CheckAndGetPartner(ctx.ChatID, userStore)
		if errMsg != "" {
			return ctx.Reply(errMsg)
		}

		req := &tgx.SendDocumentRequest{
			Document: ctx.Document.FileId,
			BaseMediaRequest: tgx.BaseMediaRequest{
				ChatId: partner.ChatId,
			},
		}

		return bot.SendDocument(req)
	})

	bot.OnMessage("Video", func(ctx *tgx.Context) error {

		partner, errMsg := CheckAndGetPartner(ctx.ChatID, userStore)
		if errMsg != "" {
			return ctx.Reply(errMsg)
		}

		req := &tgx.SendVideoRequest{
			Video: ctx.Video.FileId,
			BaseMediaRequest: tgx.BaseMediaRequest{
				ChatId: partner.ChatId,
			},
		}

		return bot.SendVideo(req)
	})

	bot.OnMessage("Sticker", func(ctx *tgx.Context) error {

		partner, errMsg := CheckAndGetPartner(ctx.ChatID, userStore)
		if errMsg != "" {
			return ctx.Reply(errMsg)
		}

		req := &tgx.SendStickerRequest{
			Sticker: ctx.Sticker.FileId,
			BaseMediaRequest: tgx.BaseMediaRequest{
				ChatId: partner.ChatId,
			},
		}

		return bot.SendSticker(req)
	})

	if err := bot.SetWebhook(); err != nil {
		log.Fatal("Failed to set webhook:", err)
	}

	logger.Info("Starting server on :8080")
	http.HandleFunc("/webhook", bot.HandleWebhook)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server error:", err)
	}
}

func HandleConnect(b *tgx.Bot, chatId int64) error {
	userA, userAExists := userStore.GetUser(chatId)
	if !userAExists {
		userA = &store.User{ChatId: chatId, SelectedInterests: make([]string, 0)}
		userStore.AddUser(userA)
	}

	if userA.IsConnected {
		return b.SendMessage(chatId, MessageAlreadyConnected)
	}

	// Attempt to remove userA from queue if they were already waiting.
	// The bug in RemoveNode for tail element might be an edge case here,
	// but generally, it's good to try to remove.
	_ = waitingQueue.RemoveNode(chatId) // Error ignored as user might not be in queue

	userAInterests, err := userStore.GetUserInterests(chatId)
	if err != nil {
		// Should not happen if user is added correctly
		userAInterests = make([]string, 0)
	}

	// Logic for users with interests
	if len(userAInterests) > 0 {
		pendingRequeue := make([]int64, 0)
		var matchedPartner *store.User

		for i := 0; i < MaxInterestSearchDepth; i++ {
			partnerChatID, dequeueErr := waitingQueue.Dequeue()
			if dequeueErr != nil { // Queue is empty or became empty
				break
			}

			userB, userBExists := userStore.GetUser(partnerChatID)
			if !userBExists {
				continue // Skip if dequeued user is not in store (should be rare)
			}
			if userB.IsConnected { // Skip if userB is already connected (should be rare if queue is managed well)
				continue
			}


			userBInterests, _ := userStore.GetUserInterests(partnerChatID)

			if len(userBInterests) > 0 && shareCommonInterest(userAInterests, userBInterests) {
				matchedPartner = userB
				break // Found a match
			} else {
				pendingRequeue = append(pendingRequeue, partnerChatID)
			}
		}

		// Re-enqueue users who were dequeued but not matched, in reverse order to maintain relative order at the front
		for i := len(pendingRequeue) - 1; i >= 0; i-- {
			waitingQueue.EnqueueAtFront(pendingRequeue[i])
		}

		if matchedPartner != nil {
			// Connect userA and matchedPartner
			userA.IsConnected = true
			userA.IsConnecting = false
			userA.Partner = matchedPartner.ChatId

			matchedPartner.IsConnected = true
			matchedPartner.IsConnecting = false
			matchedPartner.Partner = userA.ChatId

			// Use the new message for interest-based matches
			if err := b.SendMessage(userA.ChatId, MessageConnectedWithPotentialSharedInterest); err != nil {
				// Log error, attempt to notify other user too
				b.SendMessage(matchedPartner.ChatId, MessageConnectedWithPotentialSharedInterest) // Attempt to notify partner
				return err
			}
			return b.SendMessage(matchedPartner.ChatId, MessageConnectedWithPotentialSharedInterest)
		} else {
			// No interest-based match found within depth, User A waits
			userA.IsConnecting = true
			userA.IsConnected = false
			waitingQueue.Enqueue(chatId)
			return b.SendMessage(chatId, MessageLookingForPartner)
		}
	} else {
		// Fallback to original logic for users without interests
		partnerChatID, dequeueErr := waitingQueue.Dequeue()
		if dequeueErr != nil { // Queue is empty
			userA.IsConnecting = true
			userA.IsConnected = false
			waitingQueue.Enqueue(chatId)
			return b.SendMessage(chatId, MessageLookingForPartner)
		}

		userB, userBExists := userStore.GetUser(partnerChatID)
		if userBExists && !userB.IsConnected {
			userA.IsConnected = true
			userA.IsConnecting = false
			userA.Partner = userB.ChatId

			userB.IsConnected = true
			userB.IsConnecting = false
			userB.Partner = userA.ChatId

			if err := b.SendMessage(userA.ChatId, MessageConnected); err != nil {
				b.SendMessage(userB.ChatId, MessageConnected) // Attempt to notify partner
				return err
			}
			return b.SendMessage(userB.ChatId, MessageConnected)
		} else {
			// Dequeued partner doesn't exist or is already connected, User A waits
			if userBExists && userB.IsConnected { // If userB was valid but connected, re-enqueue them.
				waitingQueue.Enqueue(partnerChatID)
			}
			userA.IsConnecting = true
			userA.IsConnected = false
			waitingQueue.Enqueue(chatId)
			return b.SendMessage(chatId, MessageLookingForPartner)
		}
	}
}


func shareCommonInterest(interests1 []string, interests2 []string) bool {
	if len(interests1) == 0 || len(interests2) == 0 {
		return false
	}
	set1 := make(map[string]struct{})
	for _, interest := range interests1 {
		set1[interest] = struct{}{}
	}
	for _, interest := range interests2 {
		if _, exists := set1[interest]; exists {
			return true
		}
	}
	return false
}

func HandleStop(b *tgx.Bot, chatId int64) error {
	user, exists := userStore.GetUser(chatId)

	if !exists || (!user.IsConnected && !user.IsConnecting) {
		return b.SendMessage(chatId, MessageConnectWithSomeoneFirst)
	}

	if user.IsConnecting {
		if err := waitingQueue.RemoveNode(chatId); err != nil {
			return b.SendMessage(chatId, "Error removing you from the queue. Please try again.")
		}
		user.IsConnecting = false
		return b.SendMessage(chatId, "You have been removed from the connection queue.")
	}

	if user.IsConnected {
		partner, partnerExists := userStore.GetUser(user.Partner)
		if partnerExists {
			partner.IsConnected = false
			partner.Partner = 0

			if err := b.SendMessage(partner.ChatId, MessagePartnerLeftChat); err != nil {
				return err
			}
		}

		user.IsConnected = false
		user.IsConnecting = false
		user.Partner = 0

		return b.SendMessage(chatId, MessageChatEnded)
	}

	// fallback (should not reach here)
	return b.SendMessage(chatId, MessageConnectWithSomeoneFirst)
}

func HandleStatus(b *tgx.Bot, chatId int64) error {
	user, exists := userStore.GetUser(chatId)
	if !exists {
		return b.SendMessage(chatId, MessageNotConnectedStatus)
	}

	if user.IsConnected {
		return b.SendMessage(chatId, MessageCurrentlyChatting)
	}

	if user.IsConnecting {
		return b.SendMessage(chatId, MessageInWaitingList)
	}

	return b.SendMessage(chatId, MessageNotConnectedStatus)
}

func CheckAndGetPartner(chatId int64, userStore *store.UserStore) (*store.User, string) {
	user, exists := userStore.GetUser(chatId)
	if !exists || !user.IsConnected {
		return nil, MessageNotConnected
	}

	partner, exists := userStore.GetUser(user.Partner)
	if !exists {
		return nil, MessagePartnerNotAvailable
	}
	return partner, ""
}

// HandleMyInterestsCommand handles the /myinterests command.
func HandleMyInterestsCommand(ctx *tgx.Context) error {
	chatID := ctx.ChatID
	_, exists := userStore.GetUser(chatID)
	if !exists {
		// Ensure user exists, add if not (safeguard)
		userStore.AddUser(&store.User{ChatId: chatID, SelectedInterests: make([]string, 0)})
	}

	buttons := [][]models.InlineKeyboardButton{
		{
			{Text: "View/Edit My Interests", CallbackData: CallbackInterestViewEdit},
		},
		{
			{Text: "Clear My Interests", CallbackData: CallbackInterestClear},
		},
		{
			{Text: "Cancel", CallbackData: CallbackInterestCancel},
		},
	}
	replyMarkup := models.InlineKeyboardMarkup{InlineKeyboard: buttons}

	return ctx.ReplyWithOpts(&tgx.SendMessageRequest{
		ChatId:      chatID,
		Text:        MessageManageInterests,
		ReplyMarkup: replyMarkup,
	})
}

func isInterestSelected(interest string, selectedInterests []string) bool {
	for _, si := range selectedInterests {
		if si == interest {
			return true
		}
	}
	return false
}

func generateInterestSelectionMessageAndKeyboard(chatID int64) (string, models.InlineKeyboardMarkup, error) {
	currentUserInterests, err := userStore.GetUserInterests(chatID)
	if err != nil {
		// This case should ideally not happen if user is created in HandleMyInterestsCommand
		// but handle defensively.
		currentUserInterests = make([]string, 0)
	}

	var messageTextBuilder strings.Builder
	messageTextBuilder.WriteString("Your current interests: ")
	if len(currentUserInterests) == 0 {
		messageTextBuilder.WriteString("None. ")
	} else {
		messageTextBuilder.WriteString(strings.Join(currentUserInterests, ", ") + ". ")
	}
	messageTextBuilder.WriteString(fmt.Sprintf("You can select up to %d.", MaxSelectedInterests))

	var keyboard [][]models.InlineKeyboardButton
	row := []models.InlineKeyboardButton{}
	for i, interest := range PredefinedInterests {
		buttonText := interest
		if isInterestSelected(interest, currentUserInterests) {
			buttonText = interest + " ✅"
		} else {
			buttonText = interest + " ➕"
		}
		row = append(row, models.InlineKeyboardButton{
			Text:         buttonText,
			CallbackData: CallbackInterestSelectPrefix + interest,
		})
		if (i+1)%2 == 0 || i == len(PredefinedInterests)-1 { // 2 buttons per row
			keyboard = append(keyboard, row)
			row = []models.InlineKeyboardButton{}
		}
	}
	keyboard = append(keyboard, []models.InlineKeyboardButton{{Text: "Done Editing", CallbackData: CallbackInterestDoneEditing}})

	return messageTextBuilder.String(), models.InlineKeyboardMarkup{InlineKeyboard: keyboard}, nil
}

// HandleInterestViewEditCallback handles the "View/Edit My Interests" button.
func HandleInterestViewEditCallback(ctx *tgx.CallbackContext) error {
	chatID := ctx.GetChatID()
	messageText, replyMarkup, err := generateInterestSelectionMessageAndKeyboard(chatID)
	if err != nil {
		ctx.AnswerCallback(&tgx.CallbackAnswerOptions{Text: "Error fetching your interests.", ShowAlert: true})
		return err
	}

	err = ctx.EditMessageText(messageText, &tgx.EditMessageTextOptions{ReplyMarkup: replyMarkup})
	if err != nil {
		// Log error, potentially answer callback with error for user
		fmt.Printf("Error editing message for View/Edit: %v\n", err)
	}
	return ctx.AnswerCallback(&tgx.CallbackAnswerOptions{})
}

// HandleInterestSelectCallback handles selection/deselection of an interest.
func HandleInterestSelectCallback(ctx *tgx.CallbackContext) error {
	chatID := ctx.GetChatID()
	interestName := strings.TrimPrefix(ctx.CallbackQuery.Data, CallbackInterestSelectPrefix)

	currentUserInterests, err := userStore.GetUserInterests(chatID)
	if err != nil {
		currentUserInterests = make([]string, 0) // Initialize if error, though user should exist
	}

	var newSelectedInterests []string
	found := false
	for _, selectedInterest := range currentUserInterests {
		if selectedInterest == interestName {
			found = true
		} else {
			newSelectedInterests = append(newSelectedInterests, selectedInterest)
		}
	}

	alertMessage := ""
	if !found { // Interest was not selected, try to add it
		if len(currentUserInterests) < MaxSelectedInterests {
			newSelectedInterests = append(newSelectedInterests, interestName)
		} else {
			alertMessage = MessageMaxInterestsReached
			newSelectedInterests = currentUserInterests // No change
		}
	}
	// If found, it's already removed from newSelectedInterests

	err = userStore.SetUserInterests(chatID, newSelectedInterests)
	if err != nil {
		ctx.AnswerCallback(&tgx.CallbackAnswerOptions{Text: "Error saving your interests.", ShowAlert: true})
		return err
	}

	// Re-render the message and keyboard
	messageText, replyMarkup, genErr := generateInterestSelectionMessageAndKeyboard(chatID)
	if genErr != nil {
		ctx.AnswerCallback(&tgx.CallbackAnswerOptions{Text: "Error updating display.", ShowAlert: true})
		return genErr
	}

	err = ctx.EditMessageText(messageText, &tgx.EditMessageTextOptions{ReplyMarkup: replyMarkup})
	if err != nil {
		fmt.Printf("Error editing message for Select: %v\n", err)
	}

	return ctx.AnswerCallback(&tgx.CallbackAnswerOptions{Text: alertMessage})
}

// HandleInterestClearCallback handles clearing all selected interests.
func HandleInterestClearCallback(ctx *tgx.CallbackContext) error {
	chatID := ctx.GetChatID()
	err := userStore.ClearUserInterests(chatID)
	if err != nil {
		ctx.AnswerCallback(&tgx.CallbackAnswerOptions{Text: "Error clearing interests.", ShowAlert: true})
		return err
	}
	err = ctx.EditMessageText(MessageInterestsCleared, &tgx.EditMessageTextOptions{ReplyMarkup: models.InlineKeyboardMarkup{}}) // Remove keyboard
	if err != nil {
		fmt.Printf("Error editing message for Clear: %v\n", err)
	}
	return ctx.AnswerCallback(&tgx.CallbackAnswerOptions{})
}

// HandleInterestDoneEditingCallback handles finishing the editing process.
func HandleInterestDoneEditingCallback(ctx *tgx.CallbackContext) error {
	chatID := ctx.GetChatID()
	currentUserInterests, err := userStore.GetUserInterests(chatID)
	if err != nil {
		currentUserInterests = make([]string, 0)
	}

	var replyText string
	if len(currentUserInterests) == 0 {
		replyText = MessageNoInterestsSelected
	} else {
		replyText = MessageInterestsUpdatedPrefix + strings.Join(currentUserInterests, ", ") + "."
	}

	err = ctx.EditMessageText(replyText, &tgx.EditMessageTextOptions{ReplyMarkup: models.InlineKeyboardMarkup{}}) // Remove keyboard
	if err != nil {
		fmt.Printf("Error editing message for Done: %v\n", err)
	}
	return ctx.AnswerCallback(&tgx.CallbackAnswerOptions{})
}

// HandleInterestCancelCallback handles cancelling the interest editing.
func HandleInterestCancelCallback(ctx *tgx.CallbackContext) error {
	err := ctx.EditMessageText(MessageInterestsNoChanges, &tgx.EditMessageTextOptions{ReplyMarkup: models.InlineKeyboardMarkup{}}) // Remove keyboard
	if err != nil {
		fmt.Printf("Error editing message for Cancel: %v\n", err)
	}
	return ctx.AnswerCallback(&tgx.CallbackAnswerOptions{})
}
