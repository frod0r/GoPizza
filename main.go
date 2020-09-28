package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

var (
	toppings = []string{
		"Schinken",
		"Salami",
		"Pilze",
		"Zuccini",
		"Paprika",
		"Zwiebeln",
		"Mais",
		"Pesto",
		"Spinat",
		"Hähnchen",
		"BBQ",
		"Ananas 😡",
		"Thunfisch 😡",
	}
	//toppingKeyboard tgbotapi.InlineKeyboardMarkup
	//Maps UserID to Map of Topping and Preference
	toppingsTheyLike map[int]map[string]bool
	// Maps username to userID
	userIds map[string]int
	//id to name
	firstNames map[int]string
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
	var toppingButtons [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton
	const itemsPerRow = 3
	i := 0

	for topping, pref := range toppingsTheyLike[user] {
		if i >= itemsPerRow {
			i = 0
			toppingButtons = append(toppingButtons, currentRow)
			currentRow = nil
		}
		if pref {
			currentRow = append(currentRow, tgbotapi.NewInlineKeyboardButtonData(topping+" ✅", topping))
		} else {
			currentRow = append(currentRow, tgbotapi.NewInlineKeyboardButtonData(topping+" ❌", topping))
		}

		i++
	}
	toppingButtons = append(toppingButtons, currentRow)
	toppingButtons = append(toppingButtons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Fertig", "done")))
	//todo NewInlineKeyboardButtonSwitch
	return tgbotapi.NewInlineKeyboardMarkup(toppingButtons...)
}

func main() {
	userIds = make(map[string]int)
	firstNames = make(map[int]string)
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
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			getIdsFromMentions(update)
			//todo evtl nur für setup nachrichten
			if username := update.Message.From.UserName; username != "" {
				log.Printf("Saved user with username %v and ID %v", username, update.Message.From.ID)
				userIds[username] = update.Message.From.ID
				firstNames[update.Message.From.ID] = update.Message.From.FirstName
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
				case "setup", "start":

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
	//todo moment mal, ich will ja mögliche kombinationen mit allen elementen drin also quasi alle permutationen mit numberofpizas -1 aufteilungen dazwischen und das dann ohne duplikate...
	// siehe https://chat.stackexchange.com/transcript/message/3837894#3837894 and https://mathematica.stackexchange.com/questions/3044/partition-a-set-into-subsets-of-size-k/3050#3050
}*/

func announceDecision(decision compromise) string {
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
	//toppingss := make([][]string, len(IDs))
	consensusOnDislikedToppings := make(map[string]bool)
	for _, id := range IDs {
		for topping, pref := range toppingsTheyLike[id] {
			if !pref {
				consensusOnDislikedToppings[topping] = true
			}
		}
		//toppingss[i] = filterLikedToppings(id)
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
