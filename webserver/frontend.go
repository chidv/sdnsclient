package main

import (
	"html/template"
	"log"
	"net/http"
	"sdnsclient"
)

const (
	version     = "1.0.0"
	statusfile  = "status.html"
	cfgfilename = "domain.json"
)

//Stats holds the configured values in domain.json file
type Stats struct {
	Domains   []string
	Qps       int
	Dnsserver string
	Version   string
}

func fetchStatus(w http.ResponseWriter, r *http.Request) {
	domains, qps, dnsserver := sdnsclient.ParseLookupCfg(cfgfilename)
	S := Stats{Domains: domains, Qps: qps, Dnsserver: dnsserver, Version: version}
	t, _ := template.ParseFiles(statusfile)
	t.Execute(w, S)
}

func main() {
	http.HandleFunc("/", fetchStatus)
	log.Fatal(http.ListenAndServe(":12345", nil))
}
