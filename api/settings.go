package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"net/http"
	"os"
	"strconv"
)

var rdb *redis.Client
var ctx = context.Background()
var settings *Settings = nil

type Settings struct {
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	TrackDiesel bool    `json:"diesel"`
	TrackSuper  bool    `json:"super"`
	TrackE10    bool    `json:"e10"`
	SetHome     bool    `json:"setHome"`
}

func setupPersistency() {
	redisUrl := os.Getenv("REDIS_URL")
	opt, _ := redis.ParseURL(redisUrl)
	rdb = redis.NewClient(opt)
}

func saveSettings(id int64) {
	key := os.Getenv("VERCEL_ENV") + "/" + strconv.FormatInt(id, 10)
	js, _ := json.Marshal(settings)
	rdb.Set(ctx, key, js, 0)
}

func loadSettings(id int64) {
	var s Settings
	key := os.Getenv("VERCEL_ENV") + "/" + strconv.FormatInt(id, 10)
	val, err := rdb.Get(ctx, key).Result()
	if err == nil {
		_ = json.Unmarshal([]byte(val), &s)
		settings = &s
	} else {
		settings = &Settings{
			Lat:         0,
			Lng:         0,
			TrackDiesel: true,
			TrackSuper:  true,
			TrackE10:    true,
			SetHome:     false,
		}
	}
}

func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "ok")
}
