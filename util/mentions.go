package util

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"pizza_decision_bot/data"
	"strconv"
	"strings"
)

func GetIdsFromMentions(update tgbotapi.Update) (ids []int) {
	if update.Message.Entities != nil {
		for i, entity := range update.Message.Entities {
			log.Printf("\n\nEntity %v: %v\n", i, entity)
			switch entity.Type {
			case "mention":
				//+1 to offset @ character
				username := update.Message.Text[entity.Offset+1 : entity.Offset+entity.Length]
				//todo: If user is still unknown to bot then id will be 0.
				if data.UserIds[username] != 0 {
					ids = append(ids, data.UserIds[username])
				}
			case "text_mention":
				log.Println(entity.User)
				data.FirstNames[entity.User.ID] = entity.User.FirstName
				ids = append(ids, entity.User.ID)
			default:
				continue
			}
		}
		log.Printf("I got the following Ids %v\n", ids)
	}
	return
}

func AnnounceDecision(decision Compromise) string {
	if decision.Toppings == nil || decision.Participants == nil {
		return "Kein Kompromiss erlangt."
	}
	var b strings.Builder
	for i, tops := range decision.Toppings {
		_, _ = fmt.Fprintf(&b, "Pizza %v für ", i) // I *could* add an offset here to count like a human but I won't
		mentions := IdsToMentions(decision.Participants[i])
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
		if i < len(decision.Toppings)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func IdsToMentions(IDs []int) (mentions []string) {
	mentions = make([]string, len(IDs))
	for i, id := range IDs {
		if firstName, exists := data.FirstNames[id]; exists {
			mentions[i] = "[" + firstName + "](tg://user?id=" + strconv.Itoa(id) + ")"
		} else {
			mentions[i] = "[Stranger " + strconv.Itoa(i) + "](tg://user?id=" + strconv.Itoa(id) + ")"
		}
	}
	return
}
