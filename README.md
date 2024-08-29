# cloudflare-dyn-dns
Update Cloudflare DNS entry with IP address when external IP address has changed (different from the DNS entry)

Optionally send a notification with Pushbullet

## Usage

```
ip_dns_update -c <path to config file>
```

* Create a config file from config.template.json and populate it with Cloudflare data
* Add a pushbullet API key if you want those notifications

## Releases
Prebuilt binaries can be downloaded from the bin directory in repo for Windows, Linux, and MacOS 