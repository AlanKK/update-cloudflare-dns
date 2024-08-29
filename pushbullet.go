package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type PushbulletNote struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

func sendPushbullet(api_key string, title string, body string) {

	note := PushbulletNote{
		Type:  "note",
		Title: title,
		Body:  body,
	}

	err := SendToPushbulletAPI(api_key, note)
	if err != nil {
		log.Fatal(err)
	}
}

func SendToPushbulletAPI(apiKey string, note PushbulletNote) error {
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
