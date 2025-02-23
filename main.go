package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

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
		bot.WithMessageTextHandler("", bot.MatchTypeContains, handlerGroupMessage),
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
		ChatID: update.Message.Chat.ID,
		Text: "Heya! This bot is mostly made to send cats to people.\n" +
			"'twas but made by @effygp; feel free to reach 'em out" +
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
		ChatID: update.Message.Chat.ID,
		Photo: &models.InputFileUpload{
			Data: apiResponse.Body,
		},
	}); err != nil {
		apiResponse.Body.Close()
		if _, nestedErr := tgBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprint("Couldn't send cat pic; error was:\n", err.Error(), "\nLength of response in bytes: ", apiResponse.ContentLength),
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
	if tagToFetch == "/tag" {
		if _, nestedErr := tgBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "What are you, nuts?\nYou didn't type in a tag.",
		}); nestedErr != nil {
			log.Print("Couldn't give more info on error; error: ", nestedErr.Error())
		}
		return
	}

	log.Print(fmt.Sprintf("-- Fetching cat by tag: %s", tagToFetch))
	apiResponse, err := http.Get(
		fmt.Sprintf("https://cataas.com/cat/%s", tagToFetch),
	)
	if apiResponse.StatusCode != http.StatusOK {
		statusCatResponse, _ := http.Get(
			fmt.Sprintf("https://http.cat/%s", strconv.Itoa(apiResponse.StatusCode)),
		)
		tgBot.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID: update.Message.Chat.ID,
			Photo: &models.InputFileUpload{
				Data: statusCatResponse.Body,
			},
		})
		apiResponse.Body.Close()
		statusCatResponse.Body.Close()
		log.Print("-- Failed to fetch cat; status wasn't 200 OK but: ", apiResponse.Status)
		return
	} else if err != nil {
		apiResponse.Body.Close()
		log.Print("-- Failed to fetch cat; error: ", err.Error())
		return
	}
	defer apiResponse.Body.Close()

	if _, err := tgBot.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID: update.Message.Chat.ID,
		Photo: &models.InputFileUpload{
			Data: apiResponse.Body,
		},
	}); err != nil {
		if _, nestedErr := tgBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprint("Couldn't send cat pic; error was:\n", err.Error(), "\nLength of response in bytes: ", apiResponse.ContentLength),
		}); nestedErr != nil {
			log.Print("Couldn't give more info on error; error: ", err.Error())
		}
	} else {
		log.Print(fmt.Sprintf("-- Sent cat by tag '%s' successfully!", tagToFetch))
	}

}

func handlerGroupMessage(ctx context.Context, tgBot *bot.Bot, update *models.Update) {
	// Regular expression to match x.com URLs
	twitterRegex := regexp.MustCompile(`https?://(?:www\.)?x\.com/[^/]+/status/[0-9]+`)
	// Find all matches in the message
	matches := twitterRegex.FindAllString(update.Message.Text, -1)

	// Check if the message is from a group chat and contains the specified chat ID
	if update.Message.Chat.Type == "group" || update.Message.Chat.Type == "supergroup" {
		// Get the chat ID from the environment variable
		specifiedChatID := os.Getenv("CENSORED_CHAT_ID")
		if specifiedChatID == "" {
			log.Print("SPECIFIED_CHAT_ID environment variable is not set.")
			return
		}


		// Check if the chat ID matches the specified chat ID
		if strconv.FormatInt(update.Message.Chat.ID, 10) == specifiedChatID {
			// Prepare the JSON payload
			messagePayload := fmt.Sprintf(`{"message": "%s"}`, update.Message.Text)

			// Send the message to the specified localhost endpoint
			endpoint := os.Getenv("CUMCEN_ENDPOINT")
			if !strings.HasPrefix(endpoint, "http://") {
				endpoint = "http://" + endpoint
			}
			resp, err := http.Post(endpoint, "application/json", strings.NewReader(messagePayload))
			if err != nil {
				log.Print("Failed to send message to localhost; error: ", err.Error())
				return
			}
			defer resp.Body.Close()

			// Decode the response
			var response struct {
				ThreatProbability float64 `json:"threat_probability"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				log.Print("Failed to decode response; error: ", err.Error())
				return
			}

			// Check the threat probability
			if response.ThreatProbability >= 85 {
				// Notify the chat about the message deletion
				tgBot.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Warning: This message will be deleted in 15 seconds due to high threat probability.",
				})

				// Run the deletion in a goroutine
				go func() {
					time.Sleep(15 * time.Second)

					// Delete the message
					tgBot.DeleteMessage(ctx, &bot.DeleteMessageParams{
						ChatID:    update.Message.Chat.ID,
						MessageID: update.Message.ID,
					})
				}()
			}
			return
		}
	}

	if len(matches) > 0 {
		replacedLinks := make([]string, len(matches))
		for i, match := range matches {
			replacedLinks[i] = strings.Replace(match, "x.com/", "girlcockx.com/", 1)
		}

		response := "Here's your tweet with better embedding:\n" + strings.Join(replacedLinks, "\n")

		if _, err := tgBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   response,
			ReplyParameters: &models.ReplyParameters{
				MessageID: update.Message.ID,
				ChatID:    update.Message.Chat.ID,
			},
		}); err != nil {
			log.Printf("Failed to send vxtwitter link: %v", err)
		}
	}
}
