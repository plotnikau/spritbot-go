package handler

import (
	"fmt"
	tkapi "github.com/alexruf/tankerkoenig-go"
	"net/http"
	"os"
	"time"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	tkApiKey := os.Getenv("TK_API_KEY")
	tk := tkapi.NewClient(tkApiKey, nil)

	stations, _, err := tk.Station.List(48.4123317, 10.1335412, 5)
	if err != nil {
		fmt.Fprintf(w, "Something bad happened: %s\n\n", err)
		return
	}

	fmt.Fprintf(w, "Prices %s\n\n", time.Now().Format(time.RFC822))
	for _, station := range stations {
		fmt.Fprintf(w, "Brand: %s\n", station.Brand)
		fmt.Fprintf(w, "Name: %s\n", station.Name)
		fmt.Fprintf(w, "Adress: %s %s, %d %s\n", station.Street, station.HouseNumber, station.PostCode, station.Place)
		fmt.Fprintf(w, "Diesel:\t%f EUR/l\n", station.Diesel)
		fmt.Fprintf(w, "E5:\t%f EUR/l\n", station.E5)
		fmt.Fprintf(w, "E10:\t%f EUR/l\n\n", station.E10)
	}
}
