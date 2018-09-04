package main

import (
	"log"
	"net"
	"sort"
	"time"

	"github.com/tatsushid/go-fastping"
)

func doPing(host string) (t time.Duration) {
	pinger := fastping.NewPinger()

	ra, err := net.ResolveIPAddr("ip4:icmp", host)
	if err != nil {
		log.Println(err)
		return
	}

	pinger.AddIPAddr(ra)
	pinger.OnRecv = func(addr *net.IPAddr, tt time.Duration) {
		t = tt
	}

	err = pinger.Run()
	if err != nil {
		log.Println(err)
	}
	return
}

func sortByPing(list []VPN) {
	debug("sorting by ping")
	sort.Slice(list, func(i, j int) bool {
		t1 := doPing(list[i].IP)
		t2 := doPing(list[j].IP)
		return t1 < t2 && t1 > 0
	})
}
