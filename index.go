package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"512b.it/godss/src/chart"
	"512b.it/godss/src/dss"
	"512b.it/godss/src/server"
	"512b.it/godss/src/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type ReplyData struct {
	Head   string
	Values []string
	Repeat bool
	After  *time.Time
}

var replyManager = make(map[int64]*ReplyData) // stores state for each user (chat_id)
var counter = dss.NewDss(1000, 24*time.Hour)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Load environment variable for TOKEN
	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("TOKEN not found in environment")
	}

	port := utils.IntOrDefault(os.Getenv("PORT"), 8080)

	fmt.Printf("Token %s...%s\n", token[:3], token[len(token)-3:])

	go initBot(token)
	initHttp(port)
}

func initBot(token string) {
	// Initialize the bot
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	// Set bot to poll updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// Handle each update (message, command, etc.)
	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := update.Message.Text

		switch {
		case strings.HasPrefix(text, "/help"):
			handleStart(bot, chatID)
		case strings.HasPrefix(text, "/start"):
			handleStart(bot, chatID)
		case strings.HasPrefix(text, "/pie"):
			handlePieStart(bot, chatID, nil)
		case strings.HasPrefix(text, "/after"):
			handlePieStart(bot, chatID, &text)
		case replyManager[chatID] != nil:
			handleReply(bot, chatID, text)
		}
	}
}

func initHttp(port int) {
	server := server.NewServer(counter, "data/words.txt")

	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/api/challenge", server.HandleChallenge)
	r.GET("/api/popularity", server.HandlePopularity)

	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		panic(err)
	}
}

func handleStart(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Hello, I'm a @dsspiebot!\n"+
		"Send me a phrase and some different conclusion of that phrase and I will tell you the online popularity.\nExample:\n"+
		"/pie\nI like\n foods\n cats\n/after 2021-01-01\nI like\n foods\n cats")
	bot.Send(msg)
}

func handlePieStart(bot *tgbotapi.BotAPI, chatID int64, afterString *string) {
	msg := tgbotapi.NewMessage(chatID, "Please send the statement.")
	bot.Send(msg)

	var after *time.Time
	if afterString != nil {
		timeStr := strings.TrimPrefix(*afterString, "/after ")
		afterParse, err := time.Parse("2006-01-02", timeStr)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "Invalid date format. Please use YYYY-MM-DD.")
			bot.Send(msg)
			return
		}

		after = &afterParse
	}

	// Initialize state for this chat
	replyManager[chatID] = &ReplyData{
		Values: []string{},
		Repeat: false,
		After:  after,
	}
}

func handleReply(bot *tgbotapi.BotAPI, chatID int64, text string) {
	data := replyManager[chatID]

	if data.Head == "" { // First reply
		data.Head = text
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Send me the first option for (<i>\"%s\"</i>)", data.Head))
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return
	}

	if text != "/done" { // Collecting options
		data.Values = append(data.Values, text)
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Send me the next option for (<i>\"%s\"</i>)\nSend /done when you are done.", data.Head))
		msg.ParseMode = "HTML"
		bot.Send(msg)
	} else { // User finished input
		samples := "\n- " + strings.Join(data.Values, "\n- ")
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Elaborating plot for <i>\"%s\"</i>\n%s", data.Head, samples))
		msg.ParseMode = "HTML"
		bot.Send(msg)

		// Call DSS function (replace with actual logic)
		go generateChart(bot, chatID, data.Head, data.Values, data.After)
		delete(replyManager, chatID) // Remove state after completion
	}
}

func generateChart(bot *tgbotapi.BotAPI, chatID int64, head string, values []string, after *time.Time) {
	var err error
	// Simulate chart creation and response (replace with real implementation)
	bot.Send(tgbotapi.NewChatAction(chatID, "upload_photo"))

	println("Generating chart for", head, strings.Join(values, ", "))
	var results map[string]int64
	if results, err = counter.CountEvents(head, values, nil); err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Unable to complete the operation for <i>\"%s\"</i>.", head))
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return
	}

	orderResults := []int64{}
	for _, value := range values {
		orderResults = append(orderResults, results[value])
	}

	if after != nil {
		head = fmt.Sprintf("%s after %s", head, after.Format("2006-01-02"))
	}

	println("Creating chart for", head, strings.Join(values, ", "))
	// Create the pie chart
	var pie []byte
	if pie, err = chart.CreatePieChart(head, values, orderResults); err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Unable to complete the operation for <i>\"%s\"</i>.", head))
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return
	}

	photo := tgbotapi.FileBytes{
		Name:  "chart.png",
		Bytes: pie,
	}

	println("Sending chart for", head, strings.Join(values, ", "))
	msg := tgbotapi.NewPhoto(chatID, photo)
	bot.Send(msg)
}
