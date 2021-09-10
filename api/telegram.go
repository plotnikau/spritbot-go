package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	tkapi "github.com/alexruf/tankerkoenig-go"
	tb "gopkg.in/tucnak/telebot.v2"
	"net/http"
	"os"
)

var bot *tb.Bot
var rm *tb.ReplyMarkup
var fts *tb.ReplyMarkup

const (
	CMD_SETHOME  = "/sethome"
	CMD_FUELTYPE = "/fueltype"
)

const (
	DIESEL = "diesel"
	SUPER  = "super"
	E10    = "e10"
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
		return err
	}

	setupStandardReplyKeys()
	setupFueltypeSelectionKeys()

	bot.Handle(tb.OnLocation, handleLocation)
	bot.Handle(tb.OnText, handleText)

	bot.Handle(CMD_SETHOME, handleSetHome)
	bot.Handle(CMD_FUELTYPE, handleFueltype)

	return nil
}

func setupStandardReplyKeys() {
	rm = &tb.ReplyMarkup{ResizeReplyKeyboard: true}

	btnLocation := rm.Location("Near me")
	btnHome := rm.Text("Home")
	rm.Reply(rm.Row(btnLocation), rm.Row(btnHome))

	bot.Handle(&btnLocation, handleLocation)
	bot.Handle(&btnHome, handleHome)
}

func setupFueltypeSelectionKeys() {
	fts = &tb.ReplyMarkup{}
	btnDiesel := fts.Data(DIESEL, DIESEL)
	btnSuper := fts.Data(SUPER, SUPER)
	btnE10 := fts.Data(E10, E10)
	fts.Inline(
		fts.Row(btnDiesel, btnSuper, btnE10),
	)

	bot.Handle(&btnDiesel, func(c *tb.Callback) {
		handleFueltypeChange(c, DIESEL)
	})
	bot.Handle(&btnSuper, func(c *tb.Callback) {
		handleFueltypeChange(c, SUPER)
	})
	bot.Handle(&btnE10, func(c *tb.Callback) {
		handleFueltypeChange(c, E10)
	})
}

func handleFueltypeChange(c *tb.Callback, ft string) {
	id := c.Message.Chat.ID
	loadSettings(id)
	switch ft {
	case DIESEL:
		settings.TrackDiesel = !settings.TrackDiesel
	case SUPER:
		settings.TrackSuper = !settings.TrackSuper
	case E10:
		settings.TrackE10 = !settings.TrackE10
	}
	saveSettings(id)
	text := fmt.Sprintf("diesel: %s  |  super: %s  |  e10: %s",
		getIndicator(settings.TrackDiesel), getIndicator(settings.TrackSuper), getIndicator(settings.TrackE10))

	_, _ = bot.Edit(c.Message, text, &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: fts})
	_ = bot.Respond(c, &tb.CallbackResponse{Text: "Saved: " + text})
}

func getIndicator(value bool) string {
	if value == true {
		return "✅"
	} else {
		return "❌"
	}
}

func handleText(m *tb.Message) {
	responseText := "I'm serverless now but still WIP"
	_, _ = bot.Send(m.Sender, responseText, &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: rm})
}

func handleLocation(m *tb.Message) {
	loc := m.Location
	lat := float64(loc.Lat)
	lng := float64(loc.Lng)
	if persistLoc(m.Chat.ID, loc) == true {
		stations, err := getStations(lat, lng)
		if err != nil || len(stations) == 0 {
			responseText := "No stations found for location"
			_, _ = bot.Send(m.Sender, responseText, &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: rm})
		} else {
			responseText := "Location stored as *Home*"
			_, _ = bot.Send(m.Sender, responseText, &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: rm})
		}
	} else {
		processCoordinates(m, lat, lng)
	}
}

func handleHome(m *tb.Message) {
	loadSettings(m.Chat.ID)
	if settings.Lat != 0.0 {
		processCoordinates(m, settings.Lat, settings.Lng)
	} else {
		responseText := "Use /setHome command first"
		_, _ = bot.Send(m.Sender, responseText, &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: rm})
	}
}

func handleSetHome(m *tb.Message) {
	id := m.Chat.ID
	loadSettings(id)
	settings.SetHome = true
	saveSettings(id)
	responseText := "Send me a location and i will set it as your *Home* location"
	_, _ = bot.Send(m.Sender, responseText, &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: rm})
}

func handleFueltype(m *tb.Message) {
	id := m.Chat.ID
	loadSettings(id)
	text := fmt.Sprintf("diesel: %s  |  super: %s  |  e10: %s",
		getIndicator(settings.TrackDiesel), getIndicator(settings.TrackSuper), getIndicator(settings.TrackE10))
	_, _ = bot.Send(m.Sender, text, &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: fts})
}

func processCoordinates(m *tb.Message, lat float64, lng float64) {

	stations, err := getStations(lat, lng)

	loadSettings(m.Chat.ID)
	responseText := ""
	if err == nil {
		for _, station := range stations {
			responseText += fmt.Sprintf("*%s* - %s %s, %s\n", station.Brand, station.Street, station.HouseNumber, station.Place)
			if settings.TrackDiesel {
				responseText += fmt.Sprintf("diesel - %.3f\n", station.Diesel)
			}
			if settings.TrackSuper {
				responseText += fmt.Sprintf("super - %.3f\n", station.E5)
			}
			if settings.TrackE10 {
				responseText += fmt.Sprintf("e10 - %.3f\n\n", station.E10)
			}
		}
	} else {
		responseText = fmt.Sprintf("Something bad happened: %s\n\n", err)
	}

	_, _ = bot.Send(m.Sender, responseText, &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: rm})
}

func getStations(lat float64, lng float64) ([]tkapi.Station, error) {
	tkApiKey := os.Getenv("TK_API_KEY")
	tk := tkapi.NewClient(tkApiKey, nil)
	stations, _, err := tk.Station.List(lat, lng, 5)
	return stations, err
}

func persistLoc(id int64, loc *tb.Location) bool {
	loadSettings(id)

	if settings.SetHome == true {
		settings.Lat = float64(loc.Lat)
		settings.Lng = float64(loc.Lng)
		settings.SetHome = false
		saveSettings(id)
		return true
	}
	return false
}

func parseRequest(r *http.Request) (*tb.Update, error) {
	defer r.Body.Close()

	var update tb.Update

	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		return nil, err
	}
	if update.ID == 0 {
		return nil, errors.New("invalid update id, got update id = 0")
	}
	return &update, nil
}

func TelegramHandler(w http.ResponseWriter, r *http.Request) {
	setupPersistency()
	u, err := parseRequest(r)
	if err == nil {
		err = setupBot()
		if err == nil {
			bot.ProcessUpdate(*u)
		}
	}
	w.WriteHeader(http.StatusOK)
}
