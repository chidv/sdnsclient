package sdnsclient

import (
	"encoding/json"
	"fmt"
	"github.com/miekg/dns"
	"io/ioutil"
	"log"
	"os"
	//"os/exec"
	"strings"
	"time"
)

//lookupCfg contains the list of domain names to be lookedup. It also contains Query By Second value configured by the user
type lookupCfg struct {
	Domains string
	QPS     int
	Server  string
}

//parseLookupCfg function parses the config file inputted by the user and returns the structure lookupcfg filled
func parseLookupCfg(filename string) ([]string, int, string) {

	var lcfg lookupCfg

	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		log.Fatalf("Config file open failed")
	}
	defer file.Close()

	bytes, _ := ioutil.ReadAll(file)

	//Unmarshal the bytes to the cfg struct
	json.Unmarshal(bytes, &lcfg)

	names := strings.Split(lcfg.Domains, ", ")
	return names, lcfg.QPS, lcfg.Server
}

func resolveWorker(domains <-chan string, dnsserver string) {
	for domain := range domains {

		client := dns.Client{}
		req := dns.Msg{}
		req.SetQuestion(domain+".", dns.TypeA)
		reply, _, err := client.Exchange(&req, dnsserver+":1053")
		if err != nil {
			log.Println(err)
		}

		if err == nil {
			if len(reply.Answer) == 0 {
				log.Println("No results")
			}
			for _, ans := range reply.Answer {
				Arecord := ans.(*dns.A)
				_ = Arecord
				//fmt.Printf("Domain %s resolved to %s\n", domain, Arecord.A)
			}
		}

		//Another way of doing it - OS Way of invoking the dig command
		/*cmd := exec.Command("dig", "@"+dnsserver, domain, "-p1053")
		_ = cmd.Run()*/
	}
}

func createWorkers(count int, dnsserver string) chan string {
	domainJobs := make(chan string, count)

	for index := 0; index < count; index++ {
		go resolveWorker(domainJobs, dnsserver)
	}
	return domainJobs
}

//Sdnsclientinit gets the domain names and qps rate to trigger DNS lookups
//Sample Domain JSON File
/*{
  "domains":"example1.org, example2.org, w3schools.org, golang.org",
  "qps":200,
  "server":"127.0.0.1"
}*/
func Sdnsclientinit() {

	domains, qps, dnsserver := parseLookupCfg("domain.json")
	var interval time.Duration

	sentRate := 0
	domainIndex := 0

	//Create the worker Pools
	domainJobs := createWorkers(qps, dnsserver)

	//Divide the domains equally across the qps rate
	interval = time.Second / time.Duration(qps)

	//Start the timer with the interval
	timerChan := time.NewTicker(interval).C
	oneSecChan := time.NewTimer(time.Second).C

	for {
		select {
		case <-timerChan:
			domainJobs <- domains[domainIndex]
			domainIndex = (domainIndex + 1) % len(domains)
			sentRate = sentRate + 1
		case <-oneSecChan:
			domainIndex = 0
			fmt.Println("CurrentRate is", sentRate)
			sentRate = 0
			oneSecChan = time.NewTimer(time.Second).C
		}
	}
}
