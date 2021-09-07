package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"net/http"
	"os"
)

var rdb *redis.Client
var ctx = context.Background()

type Settings struct {
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	TrackDiesel bool    `json:"diesel"`
	TrackSuper  bool    `json:"super"`
	TrackE10    bool    `json:"e10"`
}

func setupPersistency() {
	redisUrl := os.Getenv("REDIS_URL")
	opt, _ := redis.ParseURL(redisUrl)
	rdb = redis.NewClient(opt)
}

func saveSettings(id int64, settings *Settings) {
	js, _ := json.Marshal(settings)
	rdb.Set(ctx, string(id), js, 0)
}

func loadSettings(id int64) *Settings {
	var settings Settings
	val, err := rdb.Get(ctx, string(id)).Result()
	if err == nil {
		json.Unmarshal([]byte(val), &settings)
	}
	return &settings
}

func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ok")
}
