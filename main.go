package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strings"
)

var (
	toppings = []string{"Käse", "Toastbrot", "Zuccini"}
	//toppingKeyboard tgbotapi.InlineKeyboardMarkup
	//Maps UserID to Map of Topping and Preference
	toppingsTheyLike map[int]map[string]bool
	// Maps username to userID
	userIds map[string]int
)

func updatePrefs(prefs map[string]bool, toggleToppings ...string) map[string]bool {
	if prefs == nil {
		log.Printf("NEUE LISTE ICH WIEDERHOLE NEUE LISTE\n")
		prefs = make(map[string]bool)
		for _, topping := range toppings {
			prefs[topping] = true
		}
	}
	log.Printf("To toggle %v, Prefs %v", toggleToppings, prefs)
	for _, toggled := range toggleToppings {
		prefs[toggled] = !prefs[toggled]
		log.Printf("Toggled %v", toggled)
	}
	log.Printf("Did toggle %v, Prefs %v", toggleToppings, prefs)
	return prefs
}

func personalMarkupKeyboard(user int) tgbotapi.InlineKeyboardMarkup {
	if toppingsTheyLike[user] == nil {
		toppingsTheyLike[user] = updatePrefs(toppingsTheyLike[user])
	}
	var toppingButtons []tgbotapi.InlineKeyboardButton
	for topping, pref := range toppingsTheyLike[user] {
		if pref {
			toppingButtons = append(toppingButtons, tgbotapi.NewInlineKeyboardButtonData("✅ "+topping, topping))
		} else {
			toppingButtons = append(toppingButtons, tgbotapi.NewInlineKeyboardButtonData("❎ "+topping, topping))
		}
	}
	return tgbotapi.NewInlineKeyboardMarkup(toppingButtons)
}

func main() {
	userIds = make(map[string]int)
	toppingsTheyLike = make(map[int]map[string]bool)
	//toppingButtons = make(map[string]tgbotapi.InlineKeyboardButton)
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		panic(err)
	}
	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	// Create a new UpdateConfig struct with an offset of 0. Offsets are used
	// to make sure Telegram knows we've handled previous values and we don't
	// need them repeated.
	updateConfig := tgbotapi.NewUpdate(0)

	// Tell Telegram we should wait up to 30 seconds on each request for an
	// update. This way we can get information just as quickly as making many
	// frequent requests without having to send nearly as many.
	updateConfig.Timeout = 60

	// Start polling Telegram for updates.
	updates := bot.GetUpdatesChan(updateConfig)

	// Let's go through each update that we're getting from Telegram.
	for update := range updates {
		// Telegram can send many types of updates depending on what your Bot
		// is up to. We only want to look at messages for now, so we can
		// discard any other updates.

		log.Println(update)
		if update.CallbackQuery != nil {
			//_, _ = bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
			//update.CallbackQuery.
			//update.CallbackQuery.Message
			//i, err := strconv.Atoi(update.CallbackQuery.Data)
			toppingsTheyLike[update.CallbackQuery.From.ID] = updatePrefs(toppingsTheyLike[update.CallbackQuery.From.ID], update.CallbackQuery.Data)
			log.Println(toppingsTheyLike)
			log.Printf("Lol was ist debuggen?Toggled:%v, List: %v", update.CallbackQuery.Data, toppingsTheyLike[update.CallbackQuery.From.ID])
			//toppingButtons[i] = tgbotapi.NewInlineKeyboardButtonData("s" + toppings[i], update.CallbackQuery.Data)
			toppingKeyboard := personalMarkupKeyboard(update.CallbackQuery.From.ID)
			_, _ = bot.Send(tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, toppingKeyboard))
			//_, _ = bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data))
		}
		if update.Message != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			if update.Message.Entities != nil {
				for i, entity := range update.Message.Entities {
					log.Printf("\n\nEntity %v: %v\n", i, entity)
					var userId int
					switch entity.Type {
					case "mention":
						//+1 to offset @ character
						username := update.Message.Text[entity.Offset+1 : entity.Offset+entity.Length]
						log.Println(username)
						userId = userIds[username]
					case "text_mention":
						log.Println(entity.User)
						userId = entity.User.ID
					default:
						continue
					}
					log.Printf("Id I got was %v\n", userId)
				}
			}
			//todo evtl nur für setup nachrichten
			if username := update.Message.From.UserName; username != "" {
				log.Printf("Saved user with username %v and ID %v", username, update.Message.From.ID)
				userIds[username] = update.Message.From.ID
			}
			switch text := strings.ToLower(update.Message.Text); {
			//case text == "open":
			//msg.ReplyMarkup = toppingKeyboard
			case strings.Contains(text, "nein"):
				msg.Text = "Doch!"
			case strings.Contains(text, "doch"):
				msg.Text = "Oooh!"
			}

			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "help":
					msg.Text = "Ich vermeide lange diskussionen darüber, welche Pizzen bestellt werden sollten."
				case "setup":

					msg.Text = "Was magst du denn?"
					msg.ReplyMarkup = personalMarkupKeyboard(update.Message.From.ID)
				case "kompromiss":

				case "entscheide":
					msg.Text = "Einmal zwei Party Pizzen, Margherita und Schinken." //
				case "entscheide_veg":
					msg.Text = "Margherita."
				case "entscheide_carni":
					msg.Text = "Schinken."
				default:
					msg.Text = "Was willst du?"
				}
			}
			_, _ = bot.Send(msg)
		}
		/*
			if !update.Message.IsCommand() { // any non-command Messages
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Keine Wiederrede!")
				if strings.Contains(strings.ToLower(update.Message.Text), "aber"){//Contains incoming message text
					msg.Text = "Nein!"

				}
				if strings.Contains(update.Message.Text, "🍕"){
					msg.Text = "🍕🍕!"
				}
				msg.ReplyToMessageID = update.Message.MessageID
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
				continue
			}*/
		/*

			// Create a new MessageConfig.
			// We'll take the Chat ID from the incoming message
			// and use it to create a new message.
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, "")
			// We'll also say that this message is a reply to the previous message.
			// For any other specifications than Chat ID or Text, you'll need to
			// set fields on the `MessageConfig`.
			//msg.ReplyToMessageID = update.Message.MessageID

			// Extract the command from the Message.


			// Okay, we're sending our message off! We don't care about the message
			// we just sent, so we'll discard it.
			if _, err := bot.Send(msg); err != nil {
				// Note that panics are a bad way to handle errors. Telegram can
				// have service outages or network errors, you should retry sending
				// messages or more gracefully handle failures.
				panic(err)
			}*/
	}

}
