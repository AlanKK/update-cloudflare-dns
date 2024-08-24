package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
)

type PBConfig struct {
	API_KEY string
}

type PushbulletNote struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

func sendPushbullet(title string, body string) {
	var configFile string

	switch runtime.GOOS {
	case "windows":
		configFile = os.Getenv("USERPROFILE") + "\\AppData\\Roaming\\pushbullet\\config.txt"
	case "darwin": // Mac
		configFile = os.Getenv("HOME") + "/.config/pushbullet"
	default: // Linux and others
		configFile = os.Getenv("HOME") + "/.config/pushbullet"
	}
	config, err := readConfig(configFile)

	if err != nil {
		log.Fatal(err)
	}

	note := PushbulletNote{
		Type:  "note",
		Title: title,
		Body:  body,
	}

	err = sendPushbulletNote(config.API_KEY, note)
	if err != nil {
		log.Fatal(err)
	}
}

func readConfig(file string) (*PBConfig, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "API_KEY=") {
			return &PBConfig{API_KEY: strings.TrimPrefix(line, "API_KEY=")}, nil
		}
	}

	return nil, fmt.Errorf("API_KEY not found in %s", file)
}

func sendPushbulletNote(apiKey string, note PushbulletNote) error {
	jsonData, err := json.Marshal(note)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.pushbullet.com/v2/pushes", strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}

	req.Header.Set("Access-Token", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	return nil
}
