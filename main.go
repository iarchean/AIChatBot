package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	gogpt "github.com/sashabaranov/go-gpt3"
)

var apiKey = ""

var startSequence = "\nAI: "
var restartSqeuence = "\nHuman: "

// make a map to store user session
var userSession = make(map[int64]bool)

// make a map to store user's message and bot response
var userMessages = make(map[int64]string)

// create locks since userSession and userMessages are gonig to be shared by handleUpdate concurrently to avoid race
var userSessionLock = new(sync.RWMutex)
var userMessagesLock = new(sync.RWMutex)

func main() {
	// load env vars
	godotenv.Load()
	apiKey = os.Getenv("GPT3_API_KEY")
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	// check if env vars are set
	if apiKey == "" || botToken == "" {
		log.Fatal("GPT3_API_KEY and TELEGRAM_BOT_TOKEN must be set")
	}

	// init telegram bot
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}
	// bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	// set updater to get all updates
	updater := tgbotapi.NewUpdate(0)
	updater.Timeout = 60

	updates := bot.GetUpdatesChan(updater)

	// loop through updates
	for update := range updates {
		go handleUpdate(update, bot)
	}
}

func handleUpdate(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if update.Message == nil {
		log.Print("[BOT] empty message")
		return
	}
	// set user info
	userId := update.Message.From.ID
	userName := update.Message.From.UserName
	userMessage := update.Message.Text
	userChatID := update.Message.Chat.ID

	// log user message
	log.Printf("[BOT] From %s [id:%d]: %s", userName, userId, userMessage)

	// check if user is in session
	if update.Message.IsCommand() {
		// handle commands
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		switch update.Message.Command() {
		case "start":
			msg.Text = "Hello " + update.Message.From.FirstName + " " + update.Message.From.LastName + ", just type /chat to start chatting with me"
		case "chat":
			msg.Text = "Hi, I'm a chatbot. Ask me anything!, if you want to end the chat, just type /endchat"
			userSessionLock.Lock()
			userSession[userId] = true
			userSessionLock.Unlock()
		case "endchat":
			msg.Text = "Bye " + update.Message.From.FirstName + " " + update.Message.From.LastName + ", if you want to chat again, just type /chat"
			userSessionLock.Lock()
			userSession[userId] = false
			userSessionLock.Unlock()

			userMessagesLock.Lock()
			delete(userMessages, userId)
			userMessagesLock.Unlock()
		}
		bot.Send(msg)

	} else if userSession[userId] {
		// TODO: not sure whether this check would require a read lock

		// send typing action
		sendTypingAction(bot, userChatID)

		// if user is in session, send message to GPT-3 API
		userMessagesLock.Lock()
		userMessages[userId] += userMessage + startSequence
		var resp = makeCompletionRequest(apiKey, userMessages[userId])
		userMessagesLock.Unlock()

		// send response to user
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, resp)
		bot.Send(msg)
		userMessagesLock.Lock()
		userMessages[userId] += resp + restartSqeuence
		userMessagesLock.Unlock()

		// log the response
		log.Printf("[BOT] To %s [id:%d]: %s", update.Message.From.UserName, userId, resp)
	} else {
		// if user is not in session, send message to user
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Please type /chat to start chatting with me")
		bot.Send(msg)
	}
}

// function to make request to GPT-3 API
func makeCompletionRequest(apiKey string, prompt string) string {
	c := gogpt.NewClient(apiKey)
	ctx := context.Background()

	req := gogpt.CompletionRequest{
		Model:       "text-davinci-003",
		MaxTokens:   500,
		Temperature: 0.9,
		TopP:        1,
		Prompt:      prompt,
		Stop:        []string{" Human:", " AI:"},
	}
	resp, err := c.CreateCompletion(ctx, req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return resp.Choices[0].Text
}

// function to send typing action
func sendTypingAction(bot *tgbotapi.BotAPI, chatID int64) error {
	// Create a new ChatAction object with the "typing" action
	typingAction := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)

	// Use the bot's SendChatAction method to send the typing action
	// _, err := bot.SendChatAction(typingAction)
	_, err := bot.Send(typingAction)
	return err
}
