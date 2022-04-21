package main

import (
	"encoding/gob"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	toppings = []string{
		"Schinken",
		"Salami",
		"Pilze",
		"Zucchini",
		"Paprika",
		"Zwiebeln",
		"Mais",
		"Pesto",
		"Spinat",
		"Hähnchen",
		"BBQ",
		"Ananas 😡",
		"Thunfisch 😡",
	} // todo: adding function to add toppings on the fly is not hard, but making it so that it cannot easily be abused, is.
	//toppingKeyboard tgbotapi.InlineKeyboardMarkup
	//Maps UserID to Map of Topping and Preference
	toppingsTheyLike map[int]map[string]bool
	// Maps username to userID
	userIds map[string]int
	//id to name
	firstNames map[int]string

	lock sync.Mutex
)

const (
	stringHowMany = "Du musst mir schon sagen, wie viele Pizzen ihr wollt. (Antworte auf diese Nachricht mit der anzahl an Pizzen die ihr bestellen wollt)"
	stringWho     = "Du musst mir schon sagen, wer mitessen will. Erwähne (mention) user in deiner Nachricht als antwort auf diese."
	stringWho2    = "Und für wen (außer dir selbst?"
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
		if toggled == "done" {
			continue
		}
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
	var toppingButtons [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton
	const itemsPerRow = 3
	i := 0

	for _, topping := range toppings {
		//for topping, pref := range toppingsTheyLike[user] {// range over maps is randomized https://stackoverflow.com/questions/23330781/sort-go-map-values-by-keys
		if i >= itemsPerRow {
			i = 0
			toppingButtons = append(toppingButtons, currentRow)
			currentRow = nil
		}
		if toppingsTheyLike[user][topping] { //pref
			currentRow = append(currentRow, tgbotapi.NewInlineKeyboardButtonData("✅ "+topping, topping))
		} else {
			currentRow = append(currentRow, tgbotapi.NewInlineKeyboardButtonData("❌ "+topping, topping))
		}

		i++
	}
	toppingButtons = append(toppingButtons, currentRow)
	toppingButtons = append(toppingButtons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Fertig", "done")))
	//todo NewInlineKeyboardButtonSwitch
	return tgbotapi.NewInlineKeyboardMarkup(toppingButtons...)
}

func Restore(path string, mapToRestore interface{}) error {
	lock.Lock()
	defer lock.Unlock()
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	d := gob.NewDecoder(file)
	err = d.Decode(mapToRestore)
	if err != nil {
		return err
	}
	return nil
}

func Save(path string, mapToSave interface{}) error {
	lock.Lock()
	defer lock.Unlock()
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	e := gob.NewEncoder(file)
	err = e.Encode(mapToSave)
	if err != nil {
		return err
	}
	return nil
}

func CloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	go func() {
		<-c
		log.Println("Signal to terminate received")
		err := Save("./userIds.gob", userIds)
		if err != nil {
			log.Printf("Error occured saving maps, %v\n", err)
		}
		err = Save("./firstNames.gob", firstNames)
		if err != nil {
			log.Printf("Error occured saving maps, %v\n", err)
		}
		err = Save("./toppingsTheyLike.gob", toppingsTheyLike)
		if err != nil {
			log.Printf("Error occured saving maps, %v\n", err)
		}
		err = Save("./toppings.gob", toppings)
		if err != nil {
			log.Printf("Error occured saving maps, %v\n", err)
		}
		log.Println("Done saving. Exiting")
		os.Exit(0)
	}()
}

func main() {
	userIds = make(map[string]int)
	firstNames = make(map[int]string)
	toppingsTheyLike = make(map[int]map[string]bool)

	err := Restore("./toppings.gob", &toppings)
	if err != nil {
		log.Printf("Error occured restoring maps, %v\n", err)
	} else {
		log.Printf("Restored map: %v", toppings)
	}
	err = Restore("./userIds.gob", &userIds)
	if err != nil {
		log.Printf("Error occured restoring maps, %v\n", err)
	} else {
		log.Printf("Restored map: %v", userIds)
	}
	err = Restore("./firstNames.gob", &firstNames)
	if err != nil {
		log.Printf("Error occured restoring maps, %v\n", err)
	} else {
		log.Printf("Restored map: %v", firstNames)
	}
	err = Restore("./toppingsTheyLike.gob", &toppingsTheyLike)
	if err != nil {
		log.Printf("Error occured restoring maps, %v\n", err)
	} else {
		log.Printf("Restored map: %v", toppingsTheyLike)
	}

	CloseHandler()

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
				err := Save("./userIds.gob", userIds)
				if err != nil {
					log.Printf("Error occured saving maps, %v\n", err)
				}
				err = Save("./firstNames.gob", firstNames)
				if err != nil {
					log.Printf("Error occured saving maps, %v\n", err)
				}
				err = Save("./toppingsTheyLike.gob", toppingsTheyLike)
				if err != nil {
					log.Printf("Error occured saving maps, %v\n", err)
				}
				err = Save("./toppings.gob", toppings)
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
			toppingsTheyLike[update.CallbackQuery.From.ID] = updatePrefs(toppingsTheyLike[update.CallbackQuery.From.ID], update.CallbackQuery.Data)
			log.Println(toppingsTheyLike)
			log.Printf("Lol was ist debuggen?Toggled:%v, List: %v", update.CallbackQuery.Data, toppingsTheyLike[update.CallbackQuery.From.ID])
			//toppingButtons[i] = tgbotapi.NewInlineKeyboardButtonData("s" + toppings[i], update.CallbackQuery.Data)
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
			toppingKeyboard := personalMarkupKeyboard(update.CallbackQuery.From.ID)
			_, _ = bot.Send(tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, toppingKeyboard))
			//_, _ = bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data))
		}
		if update.Message != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Was soll man da sagen...")
			getIdsFromMentions(update)
			//todo evtl. nur für setup nachrichten
			if username := update.Message.From.UserName; username != "" {
				log.Printf("Saved user with username %v and ID %v", username, update.Message.From.ID)
				userIds[username] = update.Message.From.ID
			}
			firstNames[update.Message.From.ID] = update.Message.From.FirstName
			// UserFriendly(TM) hell begins here, it is trying to answer to malformed command strings:
			if replyTo := update.Message.ReplyToMessage; replyTo != nil {
				if replyTo.Text == stringHowMany {
					msg.ReplyToMessageID = update.Message.MessageID
					re, err := regexp.Compile("[0-9]+")
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
					ids = append(ids, getIdsFromMentions(update)...)
					if !in(update.Message.From.ID, ids) {
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
					msg.Text = announceDecision(decide(numberOfPizzas, ids))
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
					ids := getIdsFromMentions(update)
					if !in(update.Message.From.ID, ids) {
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
					msg.Text = announceDecision(decide(numberOfPizzas, ids))
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
					msg.ReplyMarkup = personalMarkupKeyboard(update.Message.From.ID)
				case "kompromiss":
					re, err := regexp.Compile("[0-9]+")
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
					ids := getIdsFromMentions(update)
					if !in(update.Message.From.ID, ids) {
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
					msg.Text = announceDecision(decide(numberOfPizzas, ids))
					msg.ParseMode = "Markdown"
				case "superSecretAddTopping":
					for i, entity := range update.Message.Entities {
						log.Printf("\n\nEntity %v: %v\n", i, entity)
						switch entity.Type {
						case "bot_command":
							//+1 to offset space character
							newTopping := update.Message.Text[entity.Offset+entity.Offset+entity.Length+1:]
							toppings = append(toppings, newTopping)
							for id, _ := range toppingsTheyLike {
								toppingsTheyLike[id][newTopping] = true
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
							newToppings := make([]string, len(toppings)-1)
							i := 0
							for _, topping := range toppings {
								if topping != rmTopping {
									newToppings[i] = topping
									i++
								}
							}
							toppings = newToppings
							for id, _ := range toppingsTheyLike {
								delete(toppingsTheyLike[id], rmTopping) //could also not delete this but am not sure if this opens up for error cases, safer to delete.
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

func getIdsFromMentions(update tgbotapi.Update) (ids []int) {
	if update.Message.Entities != nil {
		for i, entity := range update.Message.Entities {
			log.Printf("\n\nEntity %v: %v\n", i, entity)
			switch entity.Type {
			case "mention":
				//+1 to offset @ character
				username := update.Message.Text[entity.Offset+1 : entity.Offset+entity.Length]
				//todo: If user is still unknown to bot then id will be 0.
				if userIds[username] != 0 {
					ids = append(ids, userIds[username])
				}
			case "text_mention":
				log.Println(entity.User)
				firstNames[entity.User.ID] = entity.User.FirstName
				ids = append(ids, entity.User.ID)
			default:
				continue
			}
		}
		log.Printf("I got the following Ids %v\n", ids)
	}
	return
}

// Own try at implementing a partition algorithm
/*func partition(IDs []int, k int) (result [][][]int) {
	//nehme an IDs ist sortiert
	n := len(IDs)
	b := combin.Binomial(n, k) / k
	l := len(IDs) / k
	result = make([][][]int, b)
	if l == 1 {
		for i := 0; i < k; i++ {
			result[i][0][0] = IDs[i]
		}
		return result
	}

	*
		for sb, id := range IDs {
				result[0][0] = id
				restResult := partition(IDs[sb:], k-1)
	*

	for sb := 0; sb < b; sb++ {
		result[sb] = make([][]int, k)
		result[sb][0][0] = IDs[0]
		result[sb][0][1] = IDs[1]
		for id, j := range IDs[0:] {
			if j < l {
				result[sb][0] = []int{id} // added afterwards without much thought, see below
			}
		}
		restResult := partition(IDs[sb:], k)
		result = append(result, restResult...) // added afterwards without much thought to make it compilable without
		// removing or commenting out this code to not get confused with the other commented out parts of code here,
		// but to be able to still see my original thoughts on this. To be removed in future commits
	}

	return
}*/

/*func oldDecide(numberOfPizzas int, IDs []int) [][][]string {
	//todo fails if number of pizzas is larger than number of people
	combinations := combin.Combinations(len(IDs), numberOfPizzas)
	//replace indices of ids with ids. todo integrate in Combinations func
	//also for every combination save the result
	results := make([][][]string, len(combinations))
	//var allOthers = make(map[uint64]nothing) todo compare if this contains already picked set of users
	for combNo, combination := range combinations {
		picked := make(map[int]nothing, len(IDs))
		others := copySliceToMap(IDs)
		for _, idIndex := range combination {
			delete(others, idIndex) // Todo: Das funktioniert so bis jetzt nur für 2 pizzen...
			//combination[i] = IDs[idIndex]
			picked[IDs[idIndex]] = nothing{}
		}

		results[combNo] = make([][]string, 2)
		results[combNo][0] = compromiseFor(picked)
		results[combNo][1] = compromiseFor(others)

	}
	return results
	//todo moment mal, ich will ja mögliche kombinationen mit allen elementen drin also quasi alle permutationen mit numberOfPizzas -1 aufteilungen dazwischen und das dann ohne duplikate...
	// siehe https://chat.stackexchange.com/transcript/message/3837894#3837894 and https://mathematica.stackexchange.com/questions/3044/partition-a-set-into-subsets-of-size-k/3050#3050
}*/

func announceDecision(decision compromise) string {
	if decision.toppings == nil || decision.participants == nil {
		return "Kein Kompromiss erlangt."
	}
	var b strings.Builder
	for i, tops := range decision.toppings {
		_, _ = fmt.Fprintf(&b, "Pizza %v für ", i) // I *could* add an offset here to count like a human but I won't
		mentions := idsToMentions(decision.participants[i])
		for j, m := range mentions {
			b.WriteString(m)
			if j < len(mentions)-2 {
				b.WriteString(", ")
			} else if j == len(mentions)-2 {
				b.WriteString(" und ")
			} else {
				b.WriteString(":\t")
			}
		}
		if len(tops) > 0 {
			for j, t := range tops {
				b.WriteString(t)
				if j < len(tops)-2 {
					b.WriteString(", ")
				} else if j == len(tops)-2 {
					b.WriteString(" und ")
				}
			}
			b.WriteString(".")
		} else {
			b.WriteString("Margherita, wirklich?")
		}
		if i < len(decision.toppings)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func decide(numberOfPizzas int, IDs []int) compromise {
	compromises := allCompromises(numberOfPizzas, IDs)
	if len(compromises) == 0 {
		return compromise{}
	}
	sort.Slice(compromises, func(i, j int) bool {
		//return people[i].Age < people[j].Age })
		sumI := 0
		for _, toppingsI := range compromises[i].toppings {
			sumI += len(toppingsI)
		}
		sumJ := 0
		for _, toppingsJ := range compromises[j].toppings {
			sumJ += len(toppingsJ)
		}
		return sumI > sumJ
	})
	return compromises[0]
	//for i, compromise := range compromises
}

func idsToMentions(IDs []int) (mentions []string) {
	mentions = make([]string, len(IDs))
	for i, id := range IDs {
		if firstName, exists := firstNames[id]; exists {
			mentions[i] = "[" + firstName + "](tg://user?id=" + strconv.Itoa(id) + ")"
		} else {
			mentions[i] = "[Stranger " + strconv.Itoa(i) + "](tg://user?id=" + strconv.Itoa(id) + ")"
		}
	}
	return
}

// when compromise is unwrapped, a group of participants ([]int) corresponds to a set of toppings ([]string) of a pizza they share
type compromise struct {
	toppings     [][]string
	participants [][]int
}

// allCompromises calculates all possible compromises that can be made with the given people.

func allCompromises(numberOfPizzas int, IDs []int) (compromises []compromise) {
	//todo add sanity checks
	sort.Ints(IDs)
	l := (len(IDs) + numberOfPizzas - 1) / numberOfPizzas // ceil of n/noOfPizzas, alternatively ```l := 1 + (len(IDs) - 1) / numberOfPizzas``` to avoid overflows, but in that case we have all other kinds of problems.
	partitions := partSub(IDs, l)
	compromises = make([]compromise, len(partitions)) //compromises = make([][][]string, len(partitions))
	for i, part := range partitions {
		compromises[i].toppings = make([][]string, len(part))
		compromises[i].participants = part
		for j, pizzaPeople := range part {
			compromises[i].toppings[j] = compromiseFor(pizzaPeople)
		}
	}
	return
}

/*func copySliceToMap(IDs []int) map[int]nothing {
	//~key = index value = id~ key = value, index omitted
	var mappedIDs = make(map[int]nothing, len(IDs))
	for _, value := range IDs {
		mappedIDs[value] = nothing{}
	}
	return mappedIDs
}

type nothing struct{}*/

func compromiseFor(IDs []int) (resultToppings []string) {
	consensusOnDislikedToppings := make(map[string]bool)
	for _, id := range IDs {
		for topping, pref := range toppingsTheyLike[id] {
			if !pref {
				consensusOnDislikedToppings[topping] = true
			}
		}
	}
	for _, topping := range toppings {
		if !consensusOnDislikedToppings[topping] {
			resultToppings = append(resultToppings, topping)
		}
	}
	return resultToppings
}

//jakobs algo
func partSub(values []int, partSize int) (partitions [][][]int) {
	if values == nil || len(values) == 0 {
		return
	}
	partitions = make([][][]int, 0)
	partSize = min(partSize, len(values))

	tuples := fixedFirstTupleBuilder(values[0], values, partSize)
	for _, tup := range tuples {
		remaining := calculateRemaining(values, tup)
		remPartitions := partSub(remaining, partSize)
		if remPartitions != nil && len(remPartitions) > 0 {
			for _, part := range remPartitions {
				partitions = append(partitions, append([][]int{tup}, part...)) //todo inefficient, see https://stackoverflow.com/questions/53737435/how-to-prepend-int-to-slice and https://github.com/golang/go/wiki/SliceTricks
			}
		} else {
			partitions = append(partitions, [][]int{tup})
		}
	}
	return
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func calculateRemaining(values, tuple []int) []int {
	//python: remaining = [v for v in values if v not in tup]
	//result := make([]int, len(values)) //uses more memory than needed, but is maybe more efficient than append
	var result []int
	for _, v := range values {
		if !in(v, tuple) {
			result = append(result, v)
		}
	}
	return result
}

//todo make more efficient by using int->noting maps, see copySliceToMap
func in(v int, tuple []int) bool {
	for _, t := range tuple {
		if v == t {
			return true
		}
	}
	return false
}

func fixedFirstTupleBuilder(fixed int, values []int, depth int) (tuples [][]int) {
	//python: values.remove(fixed) but only usage is called with fixed == values[0] so slicing is easier
	values = values[1:]
	depth--
	if depth == 0 {
		return [][]int{{fixed}}
	}
	suffixTuples := tupleBuilder(values, depth)
	tuples = make([][]int, 0)
	for _, tup := range suffixTuples {
		tuples = append(tuples, append([]int{fixed}, tup...))
	}
	return
}

func tupleBuilder(values []int, depth int) (result [][]int) {
	result = make([][]int, 0)
	if depth <= 0 {
		return
	}
	nextLevelValues := make([]int, len(values))
	copy(nextLevelValues, values)
	for _, i := range values {
		//python: next_level_values.remove(i), since we iterate over elements of values, and remove the current, without manipulating
		// nextLevelValues further, it is more efficient to simply slice here.
		nextLevelValues = nextLevelValues[1:] // alternatively we could also use the index returned by the range expression.
		if len(nextLevelValues) < depth-1 {   //not cap(nextLevelValues), reslicing changes the length but not the capacity!
			continue
		}
		nextLevelTuples := tupleBuilder(nextLevelValues, depth-1)
		if nextLevelTuples != nil && len(nextLevelTuples) > 0 {
			for _, tup := range nextLevelTuples {
				result = append(result, append([]int{i}, tup...))
			}
		} else {
			result = append(result, []int{i})
		}
	}
	return
}
