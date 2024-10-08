package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"regexp"
	"strings"

	//"invcatter/cataas"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
)

func main() {
    botOptions := []bot.Option{
	//bot.WithDefaultHandler(handlerSendPhoto),
	bot.WithMessageTextHandler("/cat", bot.MatchTypeExact, handlerSendPhoto),
	bot.WithMessageTextHandler("/tag", bot.MatchTypePrefix, handlerSendPhotoByTag),
	bot.WithMessageTextHandler("/help", bot.MatchTypeExact, handlerHelp),
	bot.WithMessageTextHandler("/start", bot.MatchTypeExact, handlerHelp),
	//bot.WithMessageTextHandler("/tag", bot.MatchTypeExact, handlerHelp),
	//bot.WithMessageTextHandler("/id", bot.MatchTypeExact, handlerHelp),
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
        log.Fatal("-- Couldn't load .env file; error: ", err.Error())
    }

    telegramApiToken, found := os.LookupEnv("API_TOKEN") 
    if !found {
        log.Fatal("-- Couldn't load API_TOKEN variable from .env; exiting now.")
    }

    b, err := bot.New(telegramApiToken, options...)
    if err != nil {
        log.Fatal("-- Couldn't construct Telegram bot object; error: ", err.Error())
    }

    return b
}

func handlerHelp(ctx context.Context, telegramBot *bot.Bot, update *models.Update) {
    if update.Message.Text == "/start" {
	log.Printf(
	    "-- Chat ID: %s; /start command received",
	    strconv.FormatInt(update.Message.Chat.ID, 10),
	)
    }

    if _, err := telegramBot.SendMessage(ctx, &bot.SendMessageParams{
	ChatID:	update.Message.Chat.ID,
	Text:	"Heya! This bot is mostly made to send cats to people.\n" +
		"'twas but made by @fecupacufacu; feel free to reach 'em out" +
		" in Telegram.\n" +
		"Try some of the following commands:\n" +
		"/cat (Will send a random cat.)\n" +
		"/tag type_cat_here (Will search for a cat by the tag you specify.)",
    }); err != nil {
	log.Print("Couldn't reply to /help or /start command; error: ", err.Error())
    }
}

func handlerSendPhoto(ctx context.Context, tgBot *bot.Bot, update *models.Update) {
    // TODO: Set actual logic whenever the message is a reply lmao
    /*
    if update.Message.ReplyToMessage != nil {
	handlerSendPhoto(ctx, bot, update)
    } else {

    }
    */
    apiResponse, err := http.Get("https://cataas.com/cat")
    if err != nil {
	log.Print("-- Failed to fetch cat; error: ", err.Error())
	return
    }
    defer apiResponse.Body.Close()


    if _, err := tgBot.SendPhoto(ctx, &bot.SendPhotoParams{
	ChatID:	update.Message.Chat.ID,
	Photo: &models.InputFileUpload{
	    Data: apiResponse.Body,
	},
    }); err != nil {
	apiResponse.Body.Close()
	if _, nestedErr := tgBot.SendMessage(ctx, &bot.SendMessageParams{
	    ChatID:	update.Message.Chat.ID,
	    Text:	fmt.Sprint("Couldn't send cat pic; error was:\n", err.Error(), "\nLength of response in bytes: ", apiResponse.ContentLength),
	}); nestedErr != nil {
	    log.Print("Couldn't give more info on error; error: ", err.Error())
	}
    } else {
	log.Print("-- Sent cat successfully!")
    }

}

func handlerSendPhotoByTag(ctx context.Context, tgBot *bot.Bot, update *models.Update) {
    // TODO: Set actual logic whenever the message is a reply lmao
    /*
    if update.Message.ReplyToMessage != nil {
	handlerSendPhoto(ctx, bot, update)
    } else {

    }
    */

    tagToFetch := strings.TrimSpace(regexp.MustCompile(`^\/tag `).ReplaceAllString(update.Message.Text, `${1}`))
    if tagToFetch == "" {
	if _, nestedErr := tgBot.SendMessage(ctx, &bot.SendMessageParams{
	    ChatID:	update.Message.Chat.ID,
	    Text:	"What are you, nuts?\nYou didn't type in a tag.",
	}); nestedErr != nil {
	    log.Print("Couldn't give more info on error; error: ", nestedErr.Error())
	}
	return
    }

    log.Print(fmt.Sprintf("-- Fetching cat by tag: %s", tagToFetch))
    apiResponse, err := http.Get(
	fmt.Sprintf("https://cataas.com/cat/%s", tagToFetch),
    )
    if err != nil {
	apiResponse.Body.Close()
	log.Print("-- Failed to fetch cat; error: ", err.Error())
	return
    }
    defer apiResponse.Body.Close()


    if _, err := tgBot.SendPhoto(ctx, &bot.SendPhotoParams{
	ChatID:	update.Message.Chat.ID,
	Photo: &models.InputFileUpload{
	    Data: apiResponse.Body,
	},
    }); err != nil {
	if _, nestedErr := tgBot.SendMessage(ctx, &bot.SendMessageParams{
	    ChatID:	update.Message.Chat.ID,
	    Text:	fmt.Sprint("Couldn't send cat pic; error was:\n", err.Error(), "\nLength of response in bytes: ", apiResponse.ContentLength),
	}); nestedErr != nil {
	    log.Print("Couldn't give more info on error; error: ", err.Error())
	}
    } else  {
	log.Print(fmt.Sprintf("-- Sent cat by tag '%s' successfully!", tagToFetch))
    }

}
