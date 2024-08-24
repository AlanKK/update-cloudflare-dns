package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	API          string `json:"api"`
	UpdateTarget []struct {
		Name   string `json:"name"`
		ID     string `json:"id"`
		ZoneID string `json:"zone_id"`
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

func logMessage(message string) {
	timestamp := time.Now().Format("%Y-%m-%d %H:%M:%S")
	fmt.Printf("%s %s\n", timestamp, message)
}

func getCurrentIP() (string, error) {
	resp, err := http.Get("http://checkip.amazonaws.com")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}

func getDNSStatus(apiKey, zoneID, recordID string) (*DNSRecord, error) {
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

func updateDNSRecord(apiKey, zoneID, recordID, name, content string, ttl int, proxied bool) (*UpdateResult, error) {
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

func main() {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer configFile.Close()
	var config Config
	err = json.NewDecoder(configFile).Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

	currentIP, err := getCurrentIP()
	if err != nil {
		logMessage("Check your internet connection")
		return
	}

	for _, target := range config.UpdateTarget {
		dnsRecord, err := getDNSStatus(config.API, target.ZoneID, target.ID)
		if err != nil {
			logMessage("Error Occurred While Accessing Current DNS Status. May Caused by outdated config file. Please re-generate config.json file (run configure.bash)")
			return
		}
		if dnsRecord.Result.Content != currentIP {
			_, err := updateDNSRecord(config.API, target.ZoneID, target.ID, target.Name, currentIP, dnsRecord.Result.TTL, dnsRecord.Result.Proxied)
			if err != nil {
				logMessage("Error While updating " + target.Name)
			} else {
				logMessage(target.Name + ": successfully updated from " + dnsRecord.Result.Content + " to " + currentIP)
			}
		}
	}
}
