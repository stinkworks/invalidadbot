package main

import (
	"fmt"
    "context"
	"log"
	"os"

	"github.com/go-telegram/bot"
	"github.com/joho/godotenv"
)

func main() {
    TelegramBot := initializeBot()
    telegramContext, cancelContext := context.WithCancel(context.Background())
    if telegramContext == nil {
        log.Fatal(`Just how the hell did you mess up?
            There is no context for the bot; exiting now.`)
    }
    TelegramBot.Start(telegramContext)
    defer cancelContext()
}

func initializeBot() *bot.Bot {
    if err := godotenv.Load(); err != nil {
        log.Fatal("Couldn't load .env file; error:", err)
    }

    telegramApiToken, found := os.LookupEnv("API_TOKEN") 
    if !found {
        log.Fatal("Couldn't load API_TOKEN variable from .env; exiting now.")
    }

    b, err := bot.New(telegramApiToken)
    if err != nil {
        log.Fatal("Couldn't construct Telegram bot object; error:", err)
    }

    return b
}
