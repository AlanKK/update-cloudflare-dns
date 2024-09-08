package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

type IPChangeConfig struct {
	PushApiKey   string `json:"pushbullet-api-key"`
	CfAPI_Key    string `json:"cf-api-key"`
	UpdateTarget []struct {
		Name   string `json:"name"`
		ID     string `json:"id"`
		ZoneID string `json:"zone-id"`
	} `json:"cf-update-target"`
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

func getArgs() (string, string) {
	if len(os.Args) != 5 {
		fmt.Println("Usage:", os.Args[0], "-c <config file> -i <ip address>")
		os.Exit(1)
	}

	var configPath string
	var ipAddress string
	flag.StringVar(&configPath, "c", "", "Path to config file")
	flag.StringVar(&ipAddress, "i", "", "IP address")
	flag.Parse()

	if configPath != "" {
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			log.Fatalf("Config file '%s' does not exist", configPath)
		}
	}
	return configPath, ipAddress
}

func main() {
	configPath, ipAddress := getArgs()

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

	// Change the cloudflare dns record
	for _, target := range config.UpdateTarget {
		dnsRecord, err := getCurrentDNSEntry(config.CfAPI_Key, target.ZoneID, target.ID)
		if err != nil {
			log.Printf("Error Occurred While Accessing Current DNS Status. May Caused by outdated config file. Please re-generate config.json file (run configure.bash)")
			return
		}
		if dnsRecord.Result.Content != ipAddress {
			_, err := updateDNSRecord(config.CfAPI_Key, target.ZoneID, target.ID, target.Name, ipAddress, dnsRecord.Result.TTL, dnsRecord.Result.Proxied)
			if err != nil {
				log.Printf("Error While updating " + target.Name)
			} else {
				log.Printf(target.Name + ": updated from " + dnsRecord.Result.Content + " to " + ipAddress)
			}
		}
	}
}
