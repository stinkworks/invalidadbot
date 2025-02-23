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
		log.Fatal(throwError(3, fmt.Errorf("missing context for bot")))
	}
	TelegramBot.Start(telegramContext)
	defer cancelContext()
}


func initializeBot(options []bot.Option) *bot.Bot {
	if err := godotenv.Load(); err != nil {
		log.Fatal(throwError(1, err))
	}

	telegramApiToken, found := os.LookupEnv("API_TOKEN")
	if !found {
		log.Fatal(throwError(2, fmt.Errorf("API_TOKEN not found")))
	}

	b, err := bot.New(telegramApiToken, options...)
	if err != nil {
		log.Fatal(throwError(3, err))
	}

	return b
}

func handlerHelp(ctx context.Context, telegramBot *bot.Bot, update *models.Update) {
	if update.Message.Text == "/start" {
		log.Printf("-- Chat ID: %s; /start command received", strconv.FormatInt(update.Message.Chat.ID, 10))
	}

	if _, err := telegramBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: "Heya! This bot is mostly made to send cats to people.\n" +
			"'twas but made by @effygp; feel free to reach 'em out in Telegram.\n" +
			"Try some of the following commands:\n" +
			"/cat (Will send a random cat.)\n" +
			"/tag type_cat_here (Will search for a cat by the tag you specify.)",
	}); err != nil {
		log.Print(throwError(4, err))
	}
}

func handlerSendPhoto(ctx context.Context, tgBot *bot.Bot, update *models.Update) {
	// if it's a reply, send a cat pic
	if isReply(update) {
		log.Print("-- Detected Reply: fetching random cat from https://cataas.com/cat")
		apiResponse, err := http.Get("https://cataas.com/cat")
		if err != nil {
			log.Print(throwError(5, err))
			return
		}
		defer apiResponse.Body.Close()

		if _, err := tgBot.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID: update.Message.Chat.ID,
			Photo: &models.InputFileUpload{
				Data: apiResponse.Body,
			},
		}); err != nil {
			log.Print(throwError(6, err))
			if _, nestedErr := tgBot.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text: fmt.Sprintf("Couldn't send cat pic; error was:\n%s\nLength of response in bytes: %d",
					err.Error(), apiResponse.ContentLength),
			}); nestedErr != nil {
				log.Print(throwError(6, nestedErr))
			}
		} else {
			log.Print("-- Sent cat successfully!")
		}
		return
	}

	apiResponse, err := http.Get("https://cataas.com/cat")
	if err != nil {
		log.Print(throwError(5, err))
		return
	}
	defer apiResponse.Body.Close()

	if _, err := tgBot.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID: update.Message.Chat.ID,
		Photo: &models.InputFileUpload{
			Data: apiResponse.Body,
		},
	}); err != nil {
		log.Print(throwError(6, err))
		if _, nestedErr := tgBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text: fmt.Sprintf("Couldn't send cat pic; error was:\n%s\nLength of response in bytes: %d",
				err.Error(), apiResponse.ContentLength),
		}); nestedErr != nil {
			log.Print(throwError(6, nestedErr))
		}
	} else {
		log.Print("-- Sent cat successfully!")
	}
}

func handlerSendPhotoByTag(ctx context.Context, tgBot *bot.Bot, update *models.Update) {
	tagToFetch := strings.TrimSpace(regexp.MustCompile(`^\/tag `).ReplaceAllString(update.Message.Text, `${1}`))
	if tagToFetch == "/tag" {
		if _, nestedErr := tgBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "What are you, nuts?\nYou didn't type in a tag.",
		}); nestedErr != nil {
			log.Print(throwError(7, nestedErr))
		}
		return
	}

	log.Print(fmt.Sprintf("-- Fetching cat by tag: %s", tagToFetch))
	apiResponse, err := http.Get(fmt.Sprintf("https://cataas.com/cat/%s", tagToFetch))
	if apiResponse.StatusCode != http.StatusOK {
		statusCatResponse, _ := http.Get(fmt.Sprintf("https://http.cat/%s", strconv.Itoa(apiResponse.StatusCode)))
		tgBot.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID: update.Message.Chat.ID,
			Photo: &models.InputFileUpload{
				Data: statusCatResponse.Body,
			},
		})
		apiResponse.Body.Close()
		statusCatResponse.Body.Close()
		log.Print(throwError(8, fmt.Errorf("status code %d", apiResponse.StatusCode)))
		return
	} else if err != nil {
		apiResponse.Body.Close()
		log.Print(throwError(5, err))
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
			Text: fmt.Sprintf("Couldn't send cat pic; error was:\n%s\nLength of response in bytes: %d",
				err.Error(), apiResponse.ContentLength),
		}); nestedErr != nil {
			log.Print(throwError(6, nestedErr))
		}
	} else {
		log.Print(fmt.Sprintf("-- Sent cat by tag '%s' successfully!", tagToFetch))
	}
}

func handlerGroupMessage(ctx context.Context, tgBot *bot.Bot, update *models.Update) {
	// Regex for Twitter URLs, this regex does that it only matches URLs from x.com
	twitterRegex := regexp.MustCompile(`https?://(?:www\.)?x\.com/[^/]+/status/[0-9]+`)
	matches := twitterRegex.FindAllString(update.Message.Text, -1)

	// if the message is from a group or supergroup, and the chat ID matches the specified chat ID
	if update.Message.Chat.Type == "group" || update.Message.Chat.Type == "supergroup" {
		specifiedChatID := os.Getenv("CENSORED_CHAT_ID")
		if specifiedChatID == "" {
			log.Print(throwError(9, fmt.Errorf("missing CENSORED_CHAT_ID")))
			return
		}

		// if the chat ID matches the specified chat ID, send the message
		if strconv.FormatInt(update.Message.Chat.ID, 10) == specifiedChatID {
			messagePayload := fmt.Sprintf(`{"message": "%s"}`, update.Message.Text)
			endpoint := os.Getenv("CUMCEN_ENDPOINT")
			if !strings.HasPrefix(endpoint, "http://") {
				endpoint = "http://" + endpoint
			}
			resp, err := http.Post(endpoint, "application/json", strings.NewReader(messagePayload))
			if err != nil {
				log.Print(throwError(10, err))
				return
			}
			defer resp.Body.Close()

			// Decode the response
			var response struct {
				ThreatProbability float64 `json:"threat_probability"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				log.Print(throwError(11, err))
				return
			}

			// Threat prolly detected 
			if response.ThreatProbability >= 0 {
				replyMessage, err := tgBot.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Warning: This message will be deleted in 15 seconds due to high threat probability.",
					ReplyParameters: &models.ReplyParameters{
						MessageID: update.Message.ID,
						ChatID:    update.Message.Chat.ID,
					},
				})
				if err != nil {
					log.Print(throwError(12, err))
				}

				go func() {
					time.Sleep(15 * time.Second)
					tgBot.DeleteMessage(ctx, &bot.DeleteMessageParams{
						ChatID:    update.Message.Chat.ID,
						MessageID: update.Message.ID,
					})
					tgBot.DeleteMessage(ctx, &bot.DeleteMessageParams{
						ChatID:    update.Message.Chat.ID,
						MessageID: replyMessage.ID,
					})
				}()
			}
			return
		}
	}

	// if the message contains a Twitter URL, replace the URL with a better one
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

func isReply(update *models.Update) bool {
	return update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage != nil
}

// throwError is a function that returns an error with a specific error message
// based on the errorID provided
func throwError(errorID int, err error) error {
	var errorMessage string
	switch errorID {
	case 1:
		errorMessage = "Couldn't load .env file"
	case 2:
		errorMessage = "Couldn't load API_TOKEN variable from .env"
	case 3:
		errorMessage = "Couldn't construct Telegram bot object"
	case 4:
		errorMessage = "Couldn't reply to /help or /start command"
	case 5:
		errorMessage = "Failed to fetch cat"
	case 6:
		errorMessage = "Couldn't send cat pic"
	case 7:
		errorMessage = "You didn't type in a tag"
	case 8:
		errorMessage = "Failed to fetch cat by tag; status wasn't 200 OK"
	case 9:
		errorMessage = "SPECIFIED_CHAT_ID environment variable is not set"
	case 10:
		errorMessage = "Failed to send message to localhost"
	case 11:
		errorMessage = "Failed to decode response"
	case 12:
		errorMessage = "Failed to send warning message"
	default:
		errorMessage = "Unknown error occurred"
	}
	return fmt.Errorf("Error [%d]: %s; %w", errorID, errorMessage, err)
}
