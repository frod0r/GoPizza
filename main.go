package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"pizza_decision_bot/data"
	"pizza_decision_bot/util"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	stringHowMany = "Du musst mir schon sagen, wie viele Pizzen ihr wollt. (Antworte auf diese Nachricht mit der anzahl an Pizzen die ihr bestellen wollt)"
	stringWho     = "Du musst mir schon sagen, wer mitessen will. Erwähne (mention) user in deiner Nachricht als antwort auf diese."
	stringWho2    = "Und für wen (außer dir selbst?)"
)

func main() {
	data.UserIds = make(map[string]int)
	data.FirstNames = make(map[int]string)
	data.ToppingsTheyLike = make(map[int]map[string]bool)

	err := data.Restore("./data.Toppings.gob", &data.Toppings)
	if err != nil {
		log.Printf("Error occured restoring maps, %v\n", err)
	} else {
		log.Printf("data.Restored map: %v", data.Toppings)
	}
	err = data.Restore("./data.UserIds.gob", &data.UserIds)
	if err != nil {
		log.Printf("Error occured restoring maps, %v\n", err)
	} else {
		log.Printf("data.Restored map: %v", data.UserIds)
	}
	err = data.Restore("./data.FirstNames.gob", &data.FirstNames)
	if err != nil {
		log.Printf("Error occured restoring maps, %v\n", err)
	} else {
		log.Printf("data.Restored map: %v", data.FirstNames)
	}
	err = data.Restore("./data.ToppingsTheyLike.gob", &data.ToppingsTheyLike)
	if err != nil {
		log.Printf("Error occured restoring maps, %v\n", err)
	} else {
		log.Printf("data.Restored map: %v", data.ToppingsTheyLike)
	}

	data.CloseHandler()

	//toppingButtons = make(map[string]tgbotapi.InlineKeyboardButton)
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Println("No api token given, quitting")
		os.Exit(1)
		//panic(err)
	}
	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	// Create a new UpdateConfig struct with an offset of 0. Offsets are used
	// to make sure Telegram knows we've handled previous values, and we don't
	// need them repeated.
	updateConfig := tgbotapi.NewUpdate(0)

	// Tell Telegram we should wait up to 30 seconds on each request for an
	// update. This way we can get information just as quickly as making many
	// frequent requests without having to send nearly as many.
	updateConfig.Timeout = 60

	// Start polling Telegram for updates.
	updates := bot.GetUpdatesChan(updateConfig)

	type partialCommand struct {
		numberOfPizzas int
		IDs            []int
		//lastMessage *tgbotapi.Message
		lastRelevant time.Time
	}

	go func() {
		for now := range time.Tick(2 * time.Hour) {
			if now.Hour() > 7 || now.Hour() < 23 {
				log.Println("Time to save stuff")
				err := data.Save("./data.UserIds.gob", data.UserIds)
				if err != nil {
					log.Printf("Error occured saving maps, %v\n", err)
				}
				err = data.Save("./data.FirstNames.gob", data.FirstNames)
				if err != nil {
					log.Printf("Error occured saving maps, %v\n", err)
				}
				err = data.Save("./data.ToppingsTheyLike.gob", data.ToppingsTheyLike)
				if err != nil {
					log.Printf("Error occured saving maps, %v\n", err)
				}
				err = data.Save("./data.Toppings.gob", data.Toppings)
				if err != nil {
					log.Printf("Error occured saving maps, %v\n", err)
				}
				log.Println("Done saving stuff")
			}
		}
	}()

	//messageID -> partialCommand
	partialCommands := make(map[int]partialCommand)
	clearInterval := 4 * time.Minute

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
			data.ToppingsTheyLike[update.CallbackQuery.From.ID] = data.UpdatePrefs(data.ToppingsTheyLike[update.CallbackQuery.From.ID], update.CallbackQuery.Data)
			log.Println(data.ToppingsTheyLike)
			log.Printf("Lol was ist debuggen?Toggled:%v, List: %v", update.CallbackQuery.Data, data.ToppingsTheyLike[update.CallbackQuery.From.ID])
			//toppingButtons[i] = tgbotapi.NewInlineKeyboardButtonData("s" + data.Toppings[i], update.CallbackQuery.Data)
			/*if update.CallbackQuery.Data == "switch_to_private" {

				_, err = bot.AnswerInlineQuery(tgbotapi.InlineConfig{
					InlineQueryID: update.CallbackQuery.ID,
					Results:       nil,
					SwitchPMText:  "hhhhh",
					SwitchPMParameter: "AAAAA_BBBBB",
				})
				if err != nil {
					log.Println(err)
				}
				continue
			}*/
			if update.CallbackQuery.Data == "done" {
				//todo add const for this
				_, _ = bot.Send(tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)) //todo not delete markup keyboard?
				_, _ = bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Okay, ist gespeichert!"))
			}
			toppingKeyboard := util.PersonalMarkupKeyboard(update.CallbackQuery.From.ID)
			_, _ = bot.Send(tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, toppingKeyboard))
			//_, _ = bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data))
		}
		if update.Message != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Was soll man da sagen...")
			util.GetIdsFromMentions(update)
			//todo evtl. nur für setup nachrichten
			if username := update.Message.From.UserName; username != "" {
				log.Printf("data.Saved user with username %v and ID %v", username, update.Message.From.ID)
				data.UserIds[username] = update.Message.From.ID
			}
			data.FirstNames[update.Message.From.ID] = update.Message.From.FirstName
			// UserFriendly(TM) hell begins here, it is trying to answer to malformed command strings:
			if replyTo := update.Message.ReplyToMessage; replyTo != nil {
				if replyTo.Text == stringHowMany {
					msg.ReplyToMessageID = update.Message.MessageID
					re, err := regexp.Compile("\\d+")
					if err != nil {
						log.Printf("Error parsing expression: %v", err)
						msg.Text = "Hoppla"
						_, _ = bot.Send(msg)
						continue
					}
					numberOfPizzasStr := re.FindAllString(update.Message.Text, 1) //we take the first number, if they write more, their problem
					if len(numberOfPizzasStr) == 0 {
						msg.ReplyToMessageID = update.Message.MessageID
						msg.Text = stringHowMany
						sent, _ := bot.Send(msg)
						partialCommands[sent.MessageID] = partialCommand{
							lastRelevant: sent.Time(),
						}
						go time.AfterFunc(clearInterval, func() { delete(partialCommands, sent.MessageID) })
						delete(partialCommands, replyTo.MessageID)
						/*if replyTo.ReplyToMessage != nil {
							//original user message we answered, they answered
							delete(partialCommands, replyTo.ReplyToMessage.MessageID)
						}*/
						continue
					}
					numberOfPizzas, err := strconv.Atoi(numberOfPizzasStr[0])
					if err != nil {
						log.Printf("Error parsing expression: %v", err)
						msg.Text = "Hoppla"
						_, _ = bot.Send(msg)
						continue
					}
					if numberOfPizzas == 0 {
						msg.Text = "Sehr witzig."
						_, _ = bot.Send(msg)
						continue
					}

					var ids []int
					if partialCommand, ok := partialCommands[replyTo.MessageID]; ok && partialCommand.IDs != nil {
						ids = partialCommand.IDs
					}
					ids = append(ids, util.GetIdsFromMentions(update)...)
					if !util.In(update.Message.From.ID, ids) {
						ids = append(ids, update.Message.From.ID)
					}
					if len(ids) <= 1 {
						msg.Text = stringWho2
						sent, _ := bot.Send(msg)
						partialCommands[sent.MessageID] = partialCommand{
							numberOfPizzas: numberOfPizzas,
							lastRelevant:   sent.Time(),
						}
						go time.AfterFunc(clearInterval, func() { delete(partialCommands, sent.MessageID) })
						delete(partialCommands, replyTo.MessageID)
						continue
					}
					msg.Text = util.AnnounceDecision(util.Decide(numberOfPizzas, ids))
					msg.ParseMode = "Markdown"
					msg.Text = stringWho2
					_, _ = bot.Send(msg)
					delete(partialCommands, replyTo.MessageID)
					continue
				} else if replyTo.Text == stringWho || replyTo.Text == stringWho2 {
					msg.ReplyToMessageID = update.Message.MessageID
					log.Printf("\nReplyTo: %v\npartialcommands: %+v\n", replyTo.MessageID, partialCommands)
					var numberOfPizzas int
					if partialCommand, ok := partialCommands[replyTo.MessageID]; !ok {
						msg.Text = "Hoppla, die Nachricht auf die du geantwortet hast war wohl zu alt. Probier's nochmal von vorne."
						//msg.Text = fmt.Sprintf("Hoppla\n%+v", partialCommands)
						_, _ = bot.Send(msg)
						continue
					} else {
						numberOfPizzas = partialCommand.numberOfPizzas
					}
					ids := util.GetIdsFromMentions(update)
					if !util.In(update.Message.From.ID, ids) {
						ids = append(ids, update.Message.From.ID)
					}
					if len(ids) <= 1 {
						msg.Text = stringWho2
						sent, _ := bot.Send(msg)
						partialCommands[sent.MessageID] = partialCommand{
							numberOfPizzas: numberOfPizzas,
							lastRelevant:   sent.Time(),
						}
						go time.AfterFunc(clearInterval, func() { delete(partialCommands, sent.MessageID) })
						delete(partialCommands, replyTo.MessageID)
						continue
					}
					msg.Text = util.AnnounceDecision(util.Decide(numberOfPizzas, ids))
					msg.ParseMode = "Markdown"
					_, _ = bot.Send(msg)
					continue
				}
				msg.Text = "Etwas ging schief im UserFriendly(TM) Mode, bitte schick einfach ein gescheites Kommando, so schwer ist das nicht..."
			}
			//  UserFriendly(TM) hell ends here
			switch text := strings.ToLower(update.Message.Text); {
			//case text == "open":
			//msg.ReplyMarkup = toppingKeyboard
			case strings.Contains(text, "nein"):
				msg.Text = "Doch!"
			case strings.Contains(text, "doch"):
				msg.Text = "Oooh!"
			}

			if update.Message.IsCommand() {
			commandSwitch:
				switch update.Message.Command() {
				case "v2":
					msg.Text = "New and Improved! Jetzt auch mit tatsächlicher Funktionalität! (wow)"
					_, _ = bot.Send(msg)
					fallthrough
				case "help":
					if update.Message.Chat.IsGroup() {
						msg.Text = "Ich vermeide lange diskussionen darüber, welche Pizzen bestellt werden sollten.\n" +
							"Sende mir `/setup` [in einer Privaten Nachricht](https://t.me/pizza_entscheide_bot?start=setup), um festzulegen, welche Zutaten du magst.\n\n" +
							"Wenn dann alle soweit sind, sende \n/kompromiss `anzahl` `@user1` ... `@userN`,\t mit der `anzahl` an Pizzen die ihr bestellen wollt " +
							"und allen leuten erwähnt (@), die außer dir mitessen wollen."
					} else {
						msg.Text = "Ich vermeide lange diskussionen darüber, welche Pizzen bestellt werden sollten.\n" +
							"Sende mir /setup, um festzulegen, welche Zutaten du magst.\n\n" +
							"Wenn dann alle soweit sind, sende \n/kompromiss `anzahl` `@user1` ... `@userN`,\t mit der `anzahl` an Pizzen die ihr bestellen wollt " +
							"und allen leuten erwähnt (@), die außer dir mitessen wollen."
					}
					msg.ParseMode = "Markdown"
				case "resetPartial":
					log.Printf("Deleting partial command %v\n", partialCommands[update.Message.From.ID])
					delete(partialCommands, update.Message.From.ID) //todo Funktioniert nicht... partielle kommandos werden mit message id gespeichert
					msg.Text = "Diese Unenstchlossenheit kotzt mich an! (Partielle Kommandos zurückgesetzt"
				case "setup", "start":
					if update.Message.Chat.IsGroup() {
						//msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("setup im Privaten chat", "switch_to_private")))
						msg.Text = "Sende mir eine [Private Nachricht](https://t.me/pizza_entscheide_bot?start=setup), um spam zu vermeiden"
						msg.ParseMode = "Markdown"
						_, err = bot.Send(msg)
						if err != nil {
							log.Println("Error sending reply markup")
						}
						continue
					}
					msg.Text = "Was magst du denn?"
					msg.ReplyMarkup = util.PersonalMarkupKeyboard(update.Message.From.ID)
				case "kompromiss":
					re, err := regexp.Compile("\\d+")
					if err != nil {
						log.Printf("Error parsing expression: %v", err)
						msg.Text = "Hoppla"
						break commandSwitch
					}
					numberOfPizzasStr := re.FindAllString(update.Message.Text, 1) //we take the first number, if they write more, their problem
					if len(numberOfPizzasStr) == 0 {
						msg.ReplyToMessageID = update.Message.MessageID
						msg.Text = stringHowMany
						sent, _ := bot.Send(msg)
						partialCommands[sent.MessageID] = partialCommand{
							lastRelevant: sent.Time(),
						}
						go time.AfterFunc(clearInterval, func() { delete(partialCommands, sent.MessageID) })
						continue
					}
					numberOfPizzas, err := strconv.Atoi(numberOfPizzasStr[0])
					if err != nil {
						log.Printf("Error parsing expression: %v", err)
						msg.Text = "Hoppla"
						break commandSwitch
					}
					ids := util.GetIdsFromMentions(update)
					if !util.In(update.Message.From.ID, ids) {
						ids = append(ids, update.Message.From.ID)
					}
					if len(ids) == 0 {
						msg.ReplyToMessageID = update.Message.MessageID
						msg.Text = stringWho
						sent, _ := bot.Send(msg)
						partialCommands[sent.MessageID] = partialCommand{
							numberOfPizzas: numberOfPizzas,
							lastRelevant:   sent.Time(),
						}
						go time.AfterFunc(clearInterval, func() { delete(partialCommands, sent.MessageID) })
						continue
					}
					msg.Text = util.AnnounceDecision(util.Decide(numberOfPizzas, ids))
					msg.ParseMode = "Markdown"
				case "superSecretAddTopping":
					for i, entity := range update.Message.Entities {
						log.Printf("\n\nEntity %v: %v\n", i, entity)
						switch entity.Type {
						case "bot_command":
							//+1 to offset space character
							newTopping := update.Message.Text[entity.Offset+entity.Offset+entity.Length+1:]
							data.Toppings = append(data.Toppings, newTopping)
							for id, _ := range data.ToppingsTheyLike {
								data.ToppingsTheyLike[id][newTopping] = true
							}
							msg.Text = "🤫"
							break commandSwitch
						default:
							continue
						}
					}
				case "superSecretRemoveTopping":
					for i, entity := range update.Message.Entities {
						log.Printf("\n\nEntity %v: %v\n", i, entity)
						switch entity.Type {
						case "bot_command":
							//+1 to offset space character
							rmTopping := update.Message.Text[entity.Offset+entity.Offset+entity.Length+1:]
							newToppings := make([]string, len(data.Toppings)-1)
							i := 0
							for _, topping := range data.Toppings {
								if topping != rmTopping {
									newToppings[i] = topping
									i++
								}
							}
							data.Toppings = newToppings
							for id, _ := range data.ToppingsTheyLike {
								delete(data.ToppingsTheyLike[id], rmTopping) //could also not delete this but am not sure if this opens up for error cases, safer to delete.
							}
							msg.Text = "🤫"
							break commandSwitch
						default:
							continue
						}
					}
				case "entscheide":
					msg.Text = "Einmal zwei Party Pizzen, Margherita und Schinken-Salami." //
				case "entscheide_veg":
					msg.Text = "Margherita."
				case "entscheide_carni":
					msg.Text = "Schinken und Salami."
				default:
					msg.Text = "Was willst du? (Probier's mal mit /help)"
				}
			}
			_, _ = bot.Send(msg)
		}
		/*
			if !update.Message.IsCommand() { // any non-command Messages
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Keine Widerrede!")
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
