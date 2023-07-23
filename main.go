package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

// API Key from https://www.weatherapi.com/
var weatherAPIKey string = "api_key_here"

type Weather struct {
	Location struct {
		Name    string `json:"name"`
		Country string `json:"country"`
	} `json:"location"`
	Current struct {
		TempC     float64 `json:"temp_c"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
	Forecast struct {
		Forecastday []struct {
			Hour []struct {
				TimeEpoch int64   `json:"time_epoch"`
				TempC     float64 `json:"temp_c"`
				Condition struct {
					Text string `json:"text"`
				} `json:"condition"`
				ChanceOfRain float64 `json:"chance_of_rain"`
			} `json:"hour"`
		} `json:"forecastday"`
	} `json:"forecast"`
}

func getWeather(location string) (*Weather, error) {
	res, err := http.Get("https://api.weatherapi.com/v1/forecast.json?key=" + weatherAPIKey + "&q=" + location + "&days=1&aqi=no&alerts=no")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Weather API not available")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var weather Weather
	err = json.Unmarshal(body, &weather)
	if err != nil {
		return nil, err
	}

	return &weather, nil
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	location := "LA"

	for {
		if len(os.Args) >= 2 {
			location = os.Args[1]
		}

		weather, err := getWeather(location)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		locationInfo := weather.Location
		currentInfo := weather.Current

		fmt.Printf("%s, %s: %.0fC, %s\n",
			locationInfo.Name,
			locationInfo.Country,
			currentInfo.TempC,
			currentInfo.Condition.Text,
		)

		hours := weather.Forecast.Forecastday[0].Hour
		for _, hour := range hours {
			date := time.Unix(hour.TimeEpoch, 0)

			if date.Before(time.Now()) {
				continue
			}

			msg := fmt.Sprintf("%s - %.0fC, %.0f%%, %s\n",
				date.Format("15:04"),
				hour.TempC,
				hour.ChanceOfRain,
				hour.Condition.Text,
			)

			if hour.ChanceOfRain < 40 {
				fmt.Print(msg)
			} else {
				color.Red(msg)
			}
		}

		fmt.Println("Enter a new location (or type 'exit' to quit):")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		location = strings.TrimSpace(input)

		if location == "exit" {
			break
		}
	}
}
