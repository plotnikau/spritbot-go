package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	tkapi "github.com/alexruf/tankerkoenig-go"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"net/http"
	"os"
)

var bot *tb.Bot
var rm *tb.ReplyMarkup

func setupBot() error {
	settings := tb.Settings{
		Token:       os.Getenv("TELEGRAM_BOT_TOKEN"),
		Synchronous: true,
		Verbose:     true,
	}
	b, err := tb.NewBot(settings)
	bot = b
	if err != nil {
		fmt.Println(err)
		return err
	}

	rm = &tb.ReplyMarkup{ResizeReplyKeyboard: true}
	// Reply buttons:
	btnLocation := rm.Location("Near me")
	rm.Reply(rm.Row(btnLocation))

	bot.Handle(&btnLocation, handleLocation)
	bot.Handle(tb.OnLocation, handleLocation)
	bot.Handle(tb.OnText, handleText)

	return nil
}

func handleText(m *tb.Message) {
	responseText := "I'm serverless now but still WIP"
	bot.Send(m.Sender, responseText, &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: rm})
}

func handleLocation(m *tb.Message) {
	loc := m.Location

	tkApiKey := os.Getenv("TK_API_KEY")
	tk := tkapi.NewClient(tkApiKey, nil)
	stations, _, err := tk.Station.List(float64(loc.Lat), float64(loc.Lng), 5)

	responseText := ""
	if err == nil {
		for _, station := range stations {
			responseText += fmt.Sprintf("*%s* - %s %s, %s\n", station.Brand, station.Street, station.HouseNumber, station.Place)
			responseText += fmt.Sprintf("diesel - %.3f\n", station.Diesel)
			responseText += fmt.Sprintf("super - %.3f\n", station.E5)
			responseText += fmt.Sprintf("e10 - %.3f\n\n", station.E10)
		}
	} else {
		responseText = fmt.Sprintf("Something bad happened: %s\n\n", err)
	}

	bot.Send(m.Sender, responseText, &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: rm})
}

func parseRequest(r *http.Request) (*tb.Update, error) {
	defer r.Body.Close()

	var update tb.Update

	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Printf("could not decode incoming update %s", err.Error())
		return nil, err
	}
	if update.ID == 0 {
		log.Printf("invalid update id, got update id = 0")
		return nil, errors.New("invalid update id, got update id = 0")
	}
	return &update, nil
}

func TelegramHandler(w http.ResponseWriter, r *http.Request) {
	u, err := parseRequest(r)
	if err == nil {
		err = setupBot()
		if err == nil {
			bot.ProcessUpdate(*u)
		}
	}
	w.WriteHeader(http.StatusOK)
}
