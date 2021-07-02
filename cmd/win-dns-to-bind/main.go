package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Svedrin/win-dns-api-go/templates"
	"github.com/Svedrin/win-dns-api-go/types"
)

func listZones(server_url string) {
	// Server URL only; list its zones
	resp, err := http.Get(fmt.Sprintf("%s/dns/", server_url))
	if err != nil {
		log.Fatalf("Error: %s", err.Error())
	}
	defer resp.Body.Close()

	var all_zones []types.DnsZone
	if err = json.NewDecoder(resp.Body).Decode(&all_zones); err != nil {
		log.Fatalf("Error: %s", err.Error())
	}

	err = templates.ZoneListTemplate.Execute(os.Stdout,
		struct {
			AllZones []types.DnsZone
		}{
			AllZones: all_zones,
		})
	if err != nil {
		log.Fatalf("Error: %s", err.Error())
	}
}

func listRecords(server_url string, zone_name string) {
	resp, err := http.Get(fmt.Sprintf("%s/dns/%s", server_url, zone_name))
	if err != nil {
		log.Fatalf("Error: %s", err.Error())
	}
	defer resp.Body.Close()

	var all_records []types.DnsRecord
	if err = json.NewDecoder(resp.Body).Decode(&all_records); err != nil {
		log.Fatalf("Error: %s", err.Error())
	}

	err = templates.ZoneTemplate.Execute(os.Stdout,
		struct {
			AllRecords []types.DnsRecord
			ZoneName   string
		}{
			AllRecords: all_records,
			ZoneName:   zone_name,
		})
	if err != nil {
		log.Fatalf("Error: %s", err.Error())
	}
}

func main() {
	if len(os.Args) == 1 {
		log.Fatalf("Usage: %s <server_url> [<domain>]\n", os.Args[0])
	}

	if len(os.Args) == 2 {
		listZones(os.Args[1])
	} else {
		for _, zone_name := range os.Args[2:] {
			listRecords(os.Args[1], zone_name)
		}
	}
}
