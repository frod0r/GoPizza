package util

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"pizza_decision_bot/data"
)

func PersonalMarkupKeyboard(user int) tgbotapi.InlineKeyboardMarkup {
	if data.ToppingsTheyLike[user] == nil {
		data.ToppingsTheyLike[user] = data.UpdatePrefs(data.ToppingsTheyLike[user])
	}
	var toppingButtons [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton
	const itemsPerRow = 3
	i := 0

	for _, topping := range data.Toppings {
		//for topping, pref := range data.ToppingsTheyLike[user] {// range over maps is randomized https://stackoverflow.com/questions/23330781/sort-go-map-values-by-keys
		if i >= itemsPerRow {
			i = 0
			toppingButtons = append(toppingButtons, currentRow)
			currentRow = nil
		}
		if data.ToppingsTheyLike[user][topping] { //pref
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
