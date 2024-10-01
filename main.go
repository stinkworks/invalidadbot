package main

import (
	"context"
	"log"
	"os"
	"net/http"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
)

func main() {
    botOptions := []bot.Option{
	//bot.WithDefaultHandler(handlerSendPhoto),
	bot.WithMessageTextHandler("/help", bot.MatchTypeExact, handlerHelp),
	bot.WithMessageTextHandler("/start", bot.MatchTypeExact, handlerHelp),
    }

    TelegramBot := initializeBot(botOptions)
    telegramContext, cancelContext := context.WithCancel(context.Background())
    if telegramContext == nil {
        log.Fatal(`Just how the hell did you mess up?
            There is no context for the bot; exiting now.`)
    }

    TelegramBot.Start(telegramContext)
    defer cancelContext()
}


func initializeBot(options []bot.Option) *bot.Bot {
    if err := godotenv.Load(); err != nil {
        log.Fatal("Couldn't load .env file; error:", err.Error())
    }

    telegramApiToken, found := os.LookupEnv("API_TOKEN") 
    if !found {
        log.Fatal("Couldn't load API_TOKEN variable from .env; exiting now.")
    }

    b, err := bot.New(telegramApiToken, options...)
    if err != nil {
        log.Fatal("Couldn't construct Telegram bot object; error:", err.Error())
    }

    return b
}

func handlerHelp(ctx context.Context, telegramBot *bot.Bot, update *models.Update) {
    if _, err := telegramBot.SendMessage(ctx, &bot.SendMessageParams{
	ChatID:	update.Message.ID,
	Text:	`Heya! This bot is mostly made to send cats to people.
		'twas but made by @fecupacufacu; feel free to reach 'em out
		in Telegram.`,
    }); err != nil {
	log.Print("Couldn't reply to /help or /start command; error:", err.Error())
    }
}
/*
func handlerSendPhoto(ctx context.Context, bot *bot.Bot, update *models.Update) {
    
}
*/
