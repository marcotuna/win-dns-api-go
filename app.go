package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"regexp"

	"github.com/gorilla/mux"
)

// DoDNSSet Set
func DoDNSSet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zoneName, dnsType, nodeName, ipAddress := vars["zoneName"], vars["dnsType"], vars["nodeName"], vars["ipAddress"]

	// Validate DNS Type
	if dnsType != "A" {
		respondWithError(w, http.StatusBadRequest, "You specified an invalid record type ('"+dnsType+"'). Currently, only the 'A' (alias) record type is supported.  e.g. /dns/my.zone/A/..")
	}

	// Validate DNS Type
	var validZoneName = regexp.MustCompile(`[^A-Za-z0-9\.-]+`)

	if validZoneName.MatchString(zoneName) {
		respondWithError(w, http.StatusBadRequest, "Invalid zone name ('"+zoneName+"'). Zone names can only contain letters, numbers, dashes (-), and dots (.).")
		return
	}

	// Validate Node Name
	var validNodeName = regexp.MustCompile(`[^A-Za-z0-9\.-]+`)

	if validNodeName.MatchString(nodeName) {
		respondWithError(w, http.StatusBadRequest, "Invalid node name ('"+nodeName+"'). Node names can only contain letters, numbers, dashes (-), and dots (.).")
		return
	}

	// Validate Ip Address
	var validIPAddress = regexp.MustCompile(`^(([1-9]?\d|1\d\d|25[0-5]|2[0-4]\d)\.){3}([1-9]?\d|1\d\d|25[0-5]|2[0-4]\d)$`)

	if !validIPAddress.MatchString(ipAddress) {
		respondWithError(w, http.StatusBadRequest, "Invalid IP address ('"+ipAddress+"'). Currently, only IPv4 addresses are accepted.")
		return
	}

	dnsCmdDeleteRecord := exec.Command("cmd", "/C", "dnscmd /recorddelete "+zoneName+" "+nodeName+" "+dnsType+" /f")

	if err := dnsCmdDeleteRecord.Run(); err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not run dnscmd /recorddelete")
		return
	}

	dnsAddDeleteRecord := exec.Command("cmd", "/C", "dnscmd /recordadd "+zoneName+" "+nodeName+" "+dnsType+" "+ipAddress)

	if err := dnsAddDeleteRecord.Run(); err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not run dnscmd /recordadd")
		return
	}

	respondWithJSON(w, http.StatusOK, "Record Updated")
}

// DoDNSRemove Remove
func DoDNSRemove(w http.ResponseWriter, r *http.Request) {

}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w, code, map[string]string{"error": msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func main() {
	r := mux.NewRouter()
	r.Methods("GET").Path("/dns/{zoneName}/{dnsType}/{nodeName}/set/{ipAddress}").HandlerFunc(DoDNSSet)
	r.Methods("POST").Path("/dns/{zoneName}/{dnsType}/{nodeName}/set/{ipAddress}").HandlerFunc(DoDNSSet)

	r.Methods("GET").Path("/dns/{zoneName}/{dnsType}/{nodeName}/remove").HandlerFunc(DoDNSRemove)
	r.Methods("POST").Path("/dns/{zoneName}/{dnsType}/{nodeName}/remove").HandlerFunc(DoDNSRemove)

	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
