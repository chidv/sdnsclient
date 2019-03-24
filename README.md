sDNS Client
===========

Scalable DNS Client can be used to trigger DNS queries to the domains and the rate configured. 
This will be really useful in scale testing of DNS Servers. 
The domain names specified in the JSON file will be requested in DNS queries in a round robin fashion.
Throttling mechanism is also added to ensure the requests are not sent in burst fashion and they are distributed evenly.

Configuration Requirements
==========================
This package needs the list of domain names, query per second (qps) rate and the DNS server. 

E.g

{
  "domains":"example1.org, example2.org, w3schools.org, golang.org",
  "qps":200,
  "server":"127.0.0.1:1053"
}

Sample Testing application using this package
==============================================
package main

import "sdnsclient"

func main() {
	sdnsclient.Sdnsclientinit()
}

./sdns

