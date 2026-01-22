package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
)

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

	// Accept comma-separated IPs, validate each
	var validIPAddress = regexp.MustCompile(`^(([1-9]?\d|1\d\d|25[0-5]|2[0-4]\d)\.){3}([1-9]?\d|1\d\d|25[0-5]|2[0-4]\d)$`)

	ipList := strings.Split(ipAddress, ",")
	cleanedIPs := make([]string, 0, len(ipList))
	for _, ip := range ipList {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		if !validIPAddress.MatchString(ip) {
			respondWithJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid IP address ('" + ip + "'). Currently, only IPv4 addresses are accepted."})
			return
		}
		cleanedIPs = append(cleanedIPs, ip)
	}

	if len(cleanedIPs) == 0 {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"message": "No valid IP addresses provided."})
		return
	}

	dnsCmdDeleteRecord := exec.Command("cmd", "/C", "dnscmd /recorddelete "+zoneName+" "+nodeName+" "+dnsType+" /f")

	if err := dnsCmdDeleteRecord.Run(); err != nil {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	// Add each IP as a separate A record
	for _, ip := range cleanedIPs {
		dnsAddRecord := exec.Command("cmd", "/C", "dnscmd /recordadd "+zoneName+" "+nodeName+" "+dnsType+" "+ip)
		if err := dnsAddRecord.Run(); err != nil {
			respondWithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
			return
		}
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "The alias ('A') record '" + nodeName + "." + zoneName + "' was successfully updated to '" + strings.Join(cleanedIPs, ",") + "'."})
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

	dnsCmdDeleteRecord := exec.Command("cmd", "/C", "dnscmd /recorddelete "+zoneName+" "+nodeName+" "+dnsType+" /f")

	if err := dnsCmdDeleteRecord.Run(); err != nil {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "The alias ('A') record '" + nodeName + "." + zoneName + "' was successfully removed."})
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusBadRequest, map[string]string{"message": "Could not get the requested route."})
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

func main() {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	r.Methods("GET").Path("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondWithJSON(w, http.StatusOK, map[string]string{"message": "Welcome to Win DNS API Go"})
	})

	r.Methods("GET").Path("/dns/{zoneName}/{dnsType}/{nodeName}/set/{ipAddress}").HandlerFunc(DoDNSSet)
	r.Methods("POST").Path("/dns/{zoneName}/{dnsType}/{nodeName}/set/{ipAddress}").HandlerFunc(DoDNSSet)

	r.Methods("GET").Path("/dns/{zoneName}/{dnsType}/{nodeName}/remove").HandlerFunc(DoDNSRemove)
	r.Methods("POST").Path("/dns/{zoneName}/{dnsType}/{nodeName}/remove").HandlerFunc(DoDNSRemove)

	fmt.Printf("Listening on port %d.\n", serverPort)

	// Start HTTP Server
	if err := http.ListenAndServe(
		fmt.Sprintf(":%d", serverPort),
		r,
	); err != nil {
		log.Fatal(err)
	}
}
