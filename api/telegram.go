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

const (
	CMD_SETHOME      = "/sethome"
	CMD_TRACKING_ON  = "/tracking_on"
	CMD_TRACKING_OFF = "/tracking_off"
)

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
	btnHome := rm.Text("Home")
	rm.Reply(rm.Row(btnLocation), rm.Row(btnHome))

	bot.Handle(&btnLocation, handleLocation)
	bot.Handle(&btnHome, handleHome)
	bot.Handle(tb.OnLocation, handleLocation)
	bot.Handle(tb.OnText, handleText)

	// commands
	bot.Handle(CMD_SETHOME, handleSetHome)

	return nil
}

func handleText(m *tb.Message) {
	responseText := "I'm serverless now but still WIP"
	bot.Send(m.Sender, responseText, &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: rm})
}

func handleLocation(m *tb.Message) {
	loc := m.Location
	processCoordinates(m, float64(loc.Lat), float64(loc.Lng))
	if persistLoc(m.Chat.ID, loc) == true {
		responseText := "Location stored as *Home*"
		bot.Send(m.Sender, responseText, &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: rm})
	}
}

func handleHome(m *tb.Message) {
	settings := loadSettings(m.Chat.ID)
	if settings.Lat != 0.0 {
		processCoordinates(m, settings.Lat, settings.Lng)
	} else {
		responseText := "Use /setHome command first"
		bot.Send(m.Sender, responseText, &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: rm})
	}
}

func handleSetHome(m *tb.Message) {
	id := m.Chat.ID
	settings := loadSettings(id)
	settings.SetHome = true
	saveSettings(id, settings)
	responseText := "Send me a location and i will set it as your *Home* location"
	bot.Send(m.Sender, responseText, &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: rm})
}

func processCoordinates(m *tb.Message, lat float64, lng float64) {
	tkApiKey := os.Getenv("TK_API_KEY")
	tk := tkapi.NewClient(tkApiKey, nil)
	stations, _, err := tk.Station.List(lat, lng, 5)

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

func persistLoc(id int64, loc *tb.Location) bool {
	settings := loadSettings(id)

	if settings.SetHome == true {
		settings.Lat = float64(loc.Lat)
		settings.Lng = float64(loc.Lng)
		settings.SetHome = false
		saveSettings(id, settings)
		return true
	}
	return false
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
			setupPersistency()
			bot.ProcessUpdate(*u)
		}
	}
	w.WriteHeader(http.StatusOK)
}
