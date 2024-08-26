package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type IPChangeConfig struct {
	PushApiKey   string `json:"pushbullet-api-key"`
	CfAPI_Key    string `json:"cf-api-key"`
	UpdateTarget []struct {
		Name   string `json:"name"`
		ID     string `json:"id"`
		ZoneID string `json:"zone-id"`
	} `json:"update-target"`
}

type DNSRecord struct {
	Result struct {
		Content string `json:"content"`
		Proxied bool   `json:"proxied"`
		TTL     int    `json:"ttl"`
	} `json:"result"`
	Success bool `json:"success"`
}

type UpdateResult struct {
	Result struct {
		Content string `json:"content"`
	} `json:"result"`
	Success bool `json:"success"`
}

func getCurrentIP() (string, error) {
	resp, err := http.Get("http://checkip.amazonaws.com")
	if err != nil {
		return "", nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(body)), nil
}

func getCurrentDNSEntry(apiKey string, zoneID string, recordID string) (*DNSRecord, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneID, recordID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+apiKey)
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var dnsRecord DNSRecord
	err = json.NewDecoder(resp.Body).Decode(&dnsRecord)
	if err != nil {
		return nil, err
	}
	return &dnsRecord, nil
}

func updateDNSRecord(apiKey string, zoneID string, recordID string, name string, content string, ttl int, proxied bool) (*UpdateResult, error) {
	data := fmt.Sprintf(`{"type":"A","name":"%s","content":"%s","ttl":%d,"proxied":%t}`, name, content, ttl, proxied)
	req, err := http.NewRequest("PUT", fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneID, recordID), bytes.NewBufferString(data))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+apiKey)
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var updateResult UpdateResult
	err = json.NewDecoder(resp.Body).Decode(&updateResult)
	if err != nil {
		return nil, err
	}
	return &updateResult, nil
}

func getArgs() string {
	if len(os.Args) != 3 {
		fmt.Println("Usage:", os.Args[0], "<title> <body>")
		os.Exit(1)
	}

	var configPath string
	flag.StringVar(&configPath, "c", "", "Path to config file")
	flag.Parse()

	if configPath != "" {
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			log.Fatalf("Config file '%s' does not exist", configPath)
		}
	}
	return configPath
}

func main() {
	configPath := getArgs()

	configFile, err := os.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}
	defer configFile.Close()

	var config IPChangeConfig
	err = json.NewDecoder(configFile).Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

	maxRetries := 120 // 10 mins
	retries := 0
	var currentIP string

	// Keep trying until we can get a connection and result from the internet
	for {
		currentIP, err = getCurrentIP()
		if currentIP == "" || err != nil {
			log.Printf("No internet connection. Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
			retries++
			if retries >= maxRetries {
				log.Printf("Failed to get current IP after max retries")
				return
			}
			continue
		}
		break
	}

	// Change the cloudflare dns record
	for _, target := range config.UpdateTarget {
		dnsRecord, err := getCurrentDNSEntry(config.CfAPI_Key, target.ZoneID, target.ID)
		if err != nil {
			log.Printf("Error Occurred While Accessing Current DNS Status. May Caused by outdated config file. Please re-generate config.json file (run configure.bash)")
			return
		}
		if dnsRecord.Result.Content != currentIP {
			_, err := updateDNSRecord(config.CfAPI_Key, target.ZoneID, target.ID, target.Name, currentIP, dnsRecord.Result.TTL, dnsRecord.Result.Proxied)
			if err != nil {
				log.Printf("Error While updating " + target.Name)
			} else {
				log.Printf(target.Name + ": successfully updated from " + dnsRecord.Result.Content + " to " + currentIP)
				sendPushbullet(config.PushApiKey, "EC2 Instance Ready", currentIP)
			}
		}
	}
}
