package data

import (
	"encoding/gob"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	Toppings = []string{
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
	} // todo: adding function to add Toppings on the fly is not hard, but making it so that it cannot easily be abused, is.
	//toppingKeyboard tgbotapi.InlineKeyboardMarkup
	//Maps UserID to Map of Topping and Preference
	ToppingsTheyLike map[int]map[string]bool
	// Maps username to userID
	UserIds map[string]int
	//id to name
	FirstNames map[int]string

	lock sync.Mutex
)

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

func UpdatePrefs(prefs map[string]bool, toggleToppings ...string) map[string]bool {
	if prefs == nil {
		log.Printf("NEUE LISTE ICH WIEDERHOLE NEUE LISTE\n")
		prefs = make(map[string]bool)
		for _, topping := range Toppings {
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

func CloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	go func() {
		<-c
		log.Println("Signal to terminate received")
		err := Save("./data.UserIds.gob", UserIds)
		if err != nil {
			log.Printf("Error occured saving maps, %v\n", err)
		}
		err = Save("./data.FirstNames.gob", FirstNames)
		if err != nil {
			log.Printf("Error occured saving maps, %v\n", err)
		}
		err = Save("./data.ToppingsTheyLike.gob", ToppingsTheyLike)
		if err != nil {
			log.Printf("Error occured saving maps, %v\n", err)
		}
		err = Save("./data.Toppings.gob", Toppings)
		if err != nil {
			log.Printf("Error occured saving maps, %v\n", err)
		}
		log.Println("Done saving. Exiting")
		os.Exit(0)
	}()
}
