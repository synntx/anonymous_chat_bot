package main

import (
	"github.com/harshyadavone/tgx"
	"github.com/harshyadavone/tgx/models"
)

const (
	MessageQuickGuide = `Quick Guide:
/connect - Find someone to chat with.
/stop - End the current chat session.
/next - Skip to the next chat partner.
/gender - Set your gender and matching preferences.
/interests - Set your interests to find better matches.
/status - Check your chat connection status.
/block - Block the current user and end the chat.
/report - Report the current user.

Be respectful and stay anonymous! 🤝

Need help or have suggestions?
Feel free to reach out anytime via @harsh_693.`

	MessageNotConnected            = "❌ You are not connected to anyone right now. Use /connect to start chatting."
	MessagePartnerNotAvailable     = "👤 Your partner has left the chat. Use /connect to find a new partner."
	MessageAlreadyConnected        = "⚠️ You are already connected to someone. If you'd like to end this chat, type /stop."
	MessageConnected               = "✨ You’re connected! Say hi to your chat partner. Type /stop if you’d like to end the chat."
	MessageLookingForPartner       = "🔍 Searching for a partner... I’ll let you know as soon as someone is ready to chat!"
	MessageConnectWithSomeoneFirst = "⚠️ Please connect with someone first! Use /connect to get started."
	MessagePartnerLeftChat         = "👋 Your chat partner has left the chat. Use /connect to find a new partner."
	MessageChatEnded               = "✅ The chat has ended. Type /connect to start a new chat!"

	MessageNotConnectedStatus = "❌ You are not connected to anyone right now. Type /connect to start chatting!"
	MessageCurrentlyChatting  = "✅ You are currently chatting with someone. Say hi! 👋"
	MessageInWaitingList      = "⌛ You are in the waiting list. I'm searching for a partner for you. Hang tight!"

	MessageErrSomethingWentWrong = "⚠️ Oops! Something went wrong on my end. Please try again in a moment. If the issue persists, contact support."

	MessageNotInChat    = "You are not in a chat. There is no one to block or report."
	MessageUserBlocked  = "🚫 User has been blocked. You will not be matched with them again. The chat has been ended."
	MessageReportThanks = "🙏 Thank you for your report. We will review the case. The chat has been ended."

	// User states
	StateDefault            = ""
	StateAwaitingGender     = "awaiting_gender"
	StateAwaitingPreference = "awaiting_preference"
	StateAwaitingInterests  = "awaiting_interests"

	// Gender options
	GenderMale   = "male"
	GenderFemale = "female"
	GenderOther  = "other"
	PrefAny      = "any"
)

var Commands = []tgx.BotCommand{
	{
		Command:     "/start",
		Description: "Get started with the bot and see the welcome message.",
	},
	{
		Command:     "/connect",
		Description: "Find a chat partner to start chatting.",
	},
	{
		Command:     "/stop",
		Description: "End the current chat session.",
	},
	{
		Command:     "/next",
		Description: "Skip to the next chat partner.",
	},
	{
		Command:     "/gender",
		Description: "Set your gender and matching preferences.",
	},
	{
		Command:     "/interests",
		Description: "Set your interests to find better matches.",
	},
	{
		Command:     "/block",
		Description: "Block the current user and find a new one.",
	},
	{
		Command:     "/report",
		Description: "Report the current user to the administrators.",
	},
	{
		Command:     "/help",
		Description: "Get a quick guide on how to use the bot.",
	},
}

var inlineKeyboardButton = [][]models.InlineKeyboardButton{
	{
		{
			Text:         "Connect",
			CallbackData: "connect",
		},
		{
			Text:         "Check Status",
			CallbackData: "status",
		},
	},
}

var genderKeyboard = models.InlineKeyboardMarkup{
	InlineKeyboard: [][]models.InlineKeyboardButton{
		{
			{Text: "Male", CallbackData: "gender_male"},
			{Text: "Female", CallbackData: "gender_female"},
			{Text: "Other", CallbackData: "gender_other"},
		},
	},
}

var preferenceKeyboard = models.InlineKeyboardMarkup{
	InlineKeyboard: [][]models.InlineKeyboardButton{
		{
			{Text: "Male", CallbackData: "pref_male"},
			{Text: "Female", CallbackData: "pref_female"},
			{Text: "Anyone", CallbackData: "pref_any"},
		},
	},
}
