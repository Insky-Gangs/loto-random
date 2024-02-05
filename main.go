package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type BotTask struct {
	mu         sync.Mutex
	active     bool
	chatID     int64
	combat     bool
	nameCombat string
	time       int
	numbers    sync.Map
}

func main() {
	// Set your Telegram bot token
	botToken := "token"
	envToken := os.Getenv("TELE_TOKEN")
	fmt.Println("envToken",envToken)
	if len(envToken) > 0 {
		botToken = envToken
	}
	
	// Create a new bot API instance
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize a BotTask
	task := &BotTask{}

	// Set up a message handler
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates, err := bot.GetUpdatesChan(updateConfig)

	// Set up a signal handler for graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start the task
	go runTask(task, bot)
	go runComabat(task, bot)
	// Process incoming messages
	for {
		select {
		case update := <-updates:
			if update.Message == nil {
				// Ignore any non-Message updates
				continue
			}

			if update.Message.IsCommand() {
				// Handle the received command
				handleCommand(bot, task, update.Message)
			}
		case <-shutdown:
			// Stop the task and exit
			stopTask(task)
			os.Exit(0)
		}
	}
}

func runTask(task *BotTask, bot *tgbotapi.BotAPI) {
	task.numbers = sync.Map{}
	rand.Seed(time.Now().UnixNano())
	for {
		// Perform your task here
		// task.mu.Lock()
		if task.active {
			fmt.Println("Task is running...")
			// Perform your task logic here

			rand.Seed(time.Now().UnixNano())

			number := rand.Intn(89) + 1
			for {
				_, ok := task.numbers.Load(number)
				if !ok {
					task.numbers.Store(number, true)
					break
				}

				number = rand.Intn(89) + 1
			}

			reply := tgbotapi.NewMessage(task.chatID, strconv.Itoa(number))
			bot.Send(reply)

			// Simulate the task running for a period
			time.Sleep(time.Duration(task.time) * time.Second)
		}
		// task.mu.Unlock()
	}
}

func runComabat(task *BotTask, bot *tgbotapi.BotAPI) {
	for {
		// Perform your task here
		// task.mu.Lock()
		if task.combat {
			// Perform your task logic here
			msg := "Khua " + task.nameCombat + " ngu"
			reply := tgbotapi.NewMessage(task.chatID, msg)
			bot.Send(reply)

			// Simulate the task running for a period
			time.Sleep(10 * time.Second)
		}
		// task.mu.Unlock()
	}
}

func stopTask(task *BotTask) {
	task.mu.Lock()
	defer task.mu.Unlock()

	// Stop the task
	task.active = false
	task.combat = false
	fmt.Println("Bingo!")
}

func handleCommand(bot *tgbotapi.BotAPI, task *BotTask, msg *tgbotapi.Message) {
	command := msg.Command()
	args := msg.CommandArguments()

	switch command {
	case "start":
		task.mu.Lock()
		task.active = true
		task.chatID = msg.Chat.ID
		time, _ := strconv.Atoi(args)
		task.time = time
		task.mu.Unlock()

		reply := tgbotapi.NewMessage(msg.Chat.ID, "Task started. Use /bingo to stop the task")
		bot.Send(reply)
	case "bingo":
		stopTask(task)

		reply := tgbotapi.NewMessage(msg.Chat.ID, "Bingooooooooooooooooooo.")
		bot.Send(reply)

	case "combat":
		task.mu.Lock()
		task.combat = true
		task.chatID = msg.Chat.ID
		task.nameCombat = args
		if strings.Contains(args, "đ") || strings.Contains(args, "Đ") {
			task.nameCombat = ""
		}
		task.mu.Unlock()

		reply := tgbotapi.NewMessage(msg.Chat.ID, "Game started. Use /bingo to stop the task")
		bot.Send(reply)

	case "reset":
		task.mu.Lock()
		task.numbers = sync.Map{}
		task.mu.Unlock()

	case "check":
		nums := strings.Split(args, " ")
		for _, num := range nums {
			numInt, _ := strconv.Atoi(num)
			if _, ok := task.numbers.Load(numInt); !ok {
				fmt.Println(num)
				reply := tgbotapi.NewMessage(msg.Chat.ID, "Invalid")
				bot.Send(reply)
				return
			}
		}

		task.numbers.Range(func(key, value any) bool {
			fmt.Println(key, value)
			return true
		})

		reply := tgbotapi.NewMessage(msg.Chat.ID, "OK")
		bot.Send(reply)

	default:
		task.mu.Lock()
		task.combat = true
		task.chatID = msg.Chat.ID
		task.nameCombat = args
		if strings.Contains(args, "đ") || strings.Contains(args, "Đ") {
			task.nameCombat = "Thinh"
		}
		task.mu.Unlock()
	}
}
