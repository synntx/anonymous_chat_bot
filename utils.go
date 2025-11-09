package main

import (
	"github.com/harshyadavone/tgx"
	"github.com/harshyadavone/tgx/models"
)

const (
	MessageQuickGuide = `Quick Guide:
/connect - Find someone to chat with.
/stop - End the current chat session.
/next - Find a new partner (not available yet).
/status - Check your chat connection status.
/report - Report your current chat partner.
/mygender - Set your gender (e.g., /mygender female).
/partnergender - Set your preferred partner gender (e.g., /partnergender male).

Be respectful and stay anonymous! 🤝

Need help or have suggestions?
Feel free to reach out anytime via @harsh_693.`

	MessageFeatureNotImplemented   = "🚧 The /next feature is not available yet. Instead, you can type /stop to end your current chat and /connect to find a new partner."
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

	MessageReportConfirmation   = "Thank you for your report. The user has been reported, and your chat has been disconnected."
	MessageNotInChat            = "You can't perform this action because you are not in a chat. Use /connect to find a partner."
	MessagePartnerReportWarning = "⚠️ Be advised: This user has been reported multiple times for their behavior. Please be cautious."

	MessageGenderSet            = "Your gender has been set to: %s."
	MessagePartnerGenderSet     = "Your preferred partner gender has been set to: %s."
	MessageInvalidGender        = "Invalid gender. Please use one of: male, female, other."
	MessageInvalidPartnerGender = "Invalid preference. Please use one of: male, female, any."

	CallbackGenderPrefix        = "gender_"
	CallbackPartnerGenderPrefix = "pgender_"
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
		Command:     "/report",
		Description: "Report your chat partner for inappropriate behavior.",
	},
	{
		Command:     "/mygender",
		Description: "Set your gender (e.g., /mygender female).",
	},
	{
		Command:     "/partnergender",
		Description: "Set your preferred partner gender (e.g., /partnergender male).",
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

var inlineKeyboardGender = [][]models.InlineKeyboardButton{
	{
		{Text: "Male", CallbackData: CallbackGenderPrefix + "male"},
		{Text: "Female", CallbackData: CallbackGenderPrefix + "female"},
		{Text: "Other", CallbackData: CallbackGenderPrefix + "other"},
	},
}

var inlineKeyboardPartnerGender = [][]models.InlineKeyboardButton{
	{
		{Text: "Male", CallbackData: CallbackPartnerGenderPrefix + "male"},
		{Text: "Female", CallbackData: CallbackPartnerGenderPrefix + "female"},
		{Text: "Any", CallbackData: CallbackPartnerGenderPrefix + "any"},
	},
}
