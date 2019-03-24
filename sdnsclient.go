package sdnsclient

import (
	"encoding/json"
	"fmt"
	"github.com/miekg/dns"
	"io/ioutil"
	"log"
	"os"
	//"os/exec"
	//"net"
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

func resolveWorker(domains <-chan string, dnsserver string, udpConn *dns.Conn) {
	for domain := range domains {
		req := dns.Msg{}
		req.SetQuestion(domain+".", dns.TypeA)

		//Send the Query
		if err := udpConn.WriteMsg(&req); err != nil {
			log.Println(err)
		}

		//Wait and Receive for the response
		reply, err := udpConn.ReadMsg()
		if err == nil {
			if len(reply.Answer) == 0 {
				log.Println("No results")
			}
			//log.Printf("Received Results")
			for _, ans := range reply.Answer {
				Arecord := ans.(*dns.A)
				_ = Arecord
				//log.Printf("Domain %s resolved to %s\n", domain, Arecord.A)
			}
		}

		//Another way of doing it - OS Way of invoking the dig command
		/*cmd := exec.Command("dig", "@"+dnsserver, domain, "-p1053")
		_ = cmd.Run()*/
	}
}

func openUDPConn(dnsserver string) *dns.Conn {

	client := dns.Client{}
	udpConn, err := client.Dial(dnsserver)
	if err != nil {
		return nil
	}
	return udpConn
}

func createWorkers(count int, dnsserver string) (chan string, []*dns.Conn) {
	domainJobs := make(chan string, count)
	udpConns := make([]*dns.Conn, count)

	for index := 0; index < count; index++ {
		udpConn := openUDPConn(dnsserver)
		udpConns = append(udpConns, udpConn)
		go resolveWorker(domainJobs, dnsserver, udpConn)
	}
	return domainJobs, udpConns
}

//Sdnsclientinit gets the domain names and qps rate to trigger DNS lookups
//Sample Domain JSON File
/*{
  "domains":"example1.org, example2.org, w3schools.org, golang.org",
  "qps":200,
  "server":"127.0.0.1:1053"
}*/
func Sdnsclientinit() {

	domains, qps, dnsserver := parseLookupCfg("domain.json")
	var interval time.Duration

	sentRate := 0
	domainIndex := 0

	//Create the worker Pools
	domainJobs, udpConns := createWorkers(qps, dnsserver)

	for _, udpConn := range udpConns {
		if udpConn != nil {
			defer udpConn.Close()
		}
	}

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
