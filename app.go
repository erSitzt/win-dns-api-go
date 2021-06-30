package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kardianos/service"
)

type DnsZone struct {
	Name       string
	Type       string
	Storage    string
	Properties []string
}

type DnsRecord struct {
	Name  string
	Type  string
	TTL   int
	Value string
	Aging int
}

func dnscmd(args ...string) *exec.Cmd {
	return exec.Command("dnscmd.exe", args...)
}

func ListDNSZones(w http.ResponseWriter, r *http.Request) {
	out, err := dnscmd("/EnumZones").Output()

	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	var all_zones []DnsZone
	in_list_of_zones := false
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if !in_list_of_zones {
			if strings.HasPrefix(line, " Zone name") {
				in_list_of_zones = true
			}
		} else if line == "Command completed successfully." {
			in_list_of_zones = false
		} else {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				// line contains Zone_name, Type, Storage, [Properties...]
				all_zones = append(all_zones, DnsZone{
					Name:       fields[0],
					Type:       fields[1],
					Storage:    fields[2],
					Properties: fields[3:],
				})
			}
		}
	}
	respondWithJSON(w, http.StatusOK, all_zones)
}

func read_aging(input string) int {
	if !strings.HasPrefix(input, "[Aging:") || !strings.HasSuffix(input, "]") {
		return -1
	}
	aging, err := strconv.Atoi(input[len("[Aging:") : len(input)-1])
	if err != nil {
		return -1
	}
	// Aging is "hours since 1601-01-01", convert to Unix timestamp
	// https://social.technet.microsoft.com/forums/windowsserver/en-US/52f2c472-f8d5-42da-bcfc-d774bf93569b/dns-aging-dnscmd-time-format
	return -11644473600 + (aging * 3600)
}

func ListDNSRecords(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	out, err := dnscmd("/EnumRecords", vars["zoneName"], "@").Output()

	if err != nil {
		if err.Error() == "exit status 9601" {
			// DNS_ERROR_ZONE_DOES_NOT_EXIST
			respondWithJSON(w, http.StatusNotFound, nil)
		} else {
			respondWithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		}
		return
	}

	var all_records []DnsRecord
	in_list_of_records := false
	prev_record_name := ""
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if !in_list_of_records {
			if line == "Returned records:" {
				in_list_of_records = true
			}
		} else if line == "" || line == "Command completed successfully." {
			in_list_of_records = false
		} else {
			fields := strings.Fields(line)
			// line can contain any of:
			// somename 3600 A 1.2.3.4
			//          3600 A 5.6.7.8
			// othername [Aging:12345678] 3600 A 1.2.3.4
			//           [Aging:12345678] 3600 A 5.6.7.8
			if line[0] != '\t' {
				// line is a full line, including the name
				prev_record_name = fields[0]
				if len(fields) == 5 {
					// Aging is set - fields are name, aging, ttl, type, value
					ttl, _ := strconv.Atoi(fields[2])
					all_records = append(all_records, DnsRecord{
						Name:  fields[0],
						Aging: read_aging(fields[1]),
						TTL:   ttl,
						Type:  fields[3],
						Value: fields[4],
					})
				} else {
					// Aging is missing - fields are name, ttl, type, value
					ttl, _ := strconv.Atoi(fields[1])
					all_records = append(all_records, DnsRecord{
						Name:  fields[0],
						Aging: 0,
						TTL:   ttl,
						Type:  fields[2],
						Value: fields[3],
					})
				}
			} else {
				// the name field is missing from line, use prev_record_name
				if len(fields) == 4 {
					// Aging is set - fields are aging, ttl, type, value
					ttl, _ := strconv.Atoi(fields[1])
					all_records = append(all_records, DnsRecord{
						Name:  prev_record_name,
						Aging: read_aging(fields[1]),
						TTL:   ttl,
						Type:  fields[2],
						Value: fields[3],
					})
				} else {
					// Aging is missing - fields are ttl, type, value
					ttl, _ := strconv.Atoi(fields[0])
					all_records = append(all_records, DnsRecord{
						Name:  prev_record_name,
						Aging: 0,
						TTL:   ttl,
						Type:  fields[1],
						Value: fields[2],
					})
				}
			}
		}
	}
	respondWithJSON(w, http.StatusOK, all_records)
}

// DoDNSSet Set
func DoDNSSet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneName, dnsType, nodeName, ipAddress := vars["zoneName"], vars["dnsType"], vars["nodeName"], vars["ipAddress"]

	// Validate DNS Type
	if dnsType != "A" {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"message": "You specified an invalid record type ('" + dnsType + "'). Currently, only the 'A' (alias) record type is supported.  e.g. /dns/my.zone/A/.."})
		return
	}

	// Validate DNS Type
	var validZoneName = regexp.MustCompile(`[^A-Za-z0-9\.-]+`)

	if validZoneName.MatchString(zoneName) {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid zone name ('" + zoneName + "'). Zone names can only contain letters, numbers, dashes (-), and dots (.)."})
		return
	}

	// Validate Node Name
	var validNodeName = regexp.MustCompile(`[^A-Za-z0-9\.-]+`)

	if validNodeName.MatchString(nodeName) {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid node name ('" + nodeName + "'). Node names can only contain letters, numbers, dashes (-), and dots (.)."})
		return
	}

	// Validate Ip Address
	var validIPAddress = regexp.MustCompile(`^(([1-9]?\d|1\d\d|25[0-5]|2[0-4]\d)\.){3}([1-9]?\d|1\d\d|25[0-5]|2[0-4]\d)$`)

	if !validIPAddress.MatchString(ipAddress) {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid IP address ('" + ipAddress + "'). Currently, only IPv4 addresses are accepted."})
		return
	}

	dnsCmdDeleteRecord := dnscmd("/recorddelete", zoneName, nodeName, dnsType, "/f")

	if err := dnsCmdDeleteRecord.Run(); err != nil {
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"message": err.Error()})
		return
	}

	dnsAddDeleteRecord := dnscmd("/recordadd", zoneName, nodeName, dnsType, ipAddress)

	if err := dnsAddDeleteRecord.Run(); err != nil {
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"message": err.Error()})
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "The alias ('A') record '" + nodeName + "." + zoneName + "' was successfully updated to '" + ipAddress + "'."})
}

// DoDNSRemove Remove
func DoDNSRemove(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneName, dnsType, nodeName := vars["zoneName"], vars["dnsType"], vars["nodeName"]

	// Validate DNS Type
	if dnsType != "A" {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"message": "You specified an invalid record type ('" + dnsType + "'). Currently, only the 'A' (alias) record type is supported.  e.g. /dns/my.zone/A/.."})
		return
	}

	// Validate DNS Type
	var validZoneName = regexp.MustCompile(`[^A-Za-z0-9\.-]+`)

	if validZoneName.MatchString(zoneName) {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid zone name ('" + zoneName + "'). Zone names can only contain letters, numbers, dashes (-), and dots (.)."})
		return
	}

	// Validate Node Name
	var validNodeName = regexp.MustCompile(`[^A-Za-z0-9\.-]+`)

	if validNodeName.MatchString(nodeName) {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid node name ('" + nodeName + "'). Node names can only contain letters, numbers, dashes (-), and dots (.)."})
		return
	}

	dnsCmdDeleteRecord := dnscmd("/recorddelete", zoneName, nodeName, dnsType, "/f")

	if err := dnsCmdDeleteRecord.Run(); err != nil {
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"message": err.Error()})
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "The alias ('A') record '" + nodeName + "." + zoneName + "' was successfully removed."})
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotFound, map[string]string{"message": "Could not get the requested route."})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

const (
	serverPort = 3111
)

var logger service.Logger

type program struct {
	servaddr string
	writable bool
}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

func (p *program) run() {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	r.Methods("GET").Path("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondWithJSON(w, http.StatusOK, map[string]string{"message": "Welcome to Win DNS API Go"})
	})

	r.Methods("GET").Path("/dns/").HandlerFunc(ListDNSZones)
	r.Methods("GET").Path("/dns/{zoneName}").HandlerFunc(ListDNSRecords)

	if p.writable {
		r.Methods("POST").Path("/dns/{zoneName}/{dnsType}/{nodeName}/set/{ipAddress}").HandlerFunc(DoDNSSet)
		r.Methods("DELETE").Path("/dns/{zoneName}/{dnsType}/{nodeName}/remove").HandlerFunc(DoDNSRemove)
	}

	fmt.Printf("Listening on %s.\n", p.servaddr)

	// Start HTTP Server
	if err := http.ListenAndServe(p.servaddr, r); err != nil {
		log.Fatal(err)
	}
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	return nil
}

func main() {
	var writable = flag.Bool("rw", false, "Enable read-write mode (i.e., allow set and remove)")
	var servaddr = flag.String("addr", ":3111", "http service address")
	flag.Parse()

	svcConfig := &service.Config{
		Name:        "WinDnsApi-Go",
		DisplayName: "Windows DNS API written in Go",
		Description: "Provides an HTTP API to manage Windows DNS on " + *servaddr,
	}

	prg := &program{
		servaddr: *servaddr,
		writable: *writable,
	}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}
