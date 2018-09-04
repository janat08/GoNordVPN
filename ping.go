package main

import (
	"fmt"
	"log"
	"net"
	"sort"
	"sync"
	"sync/atomic"
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
		debug(fmt.Sprintf("Result: %s\t%v", addr.IP.String(), tt))
	}

	err = pinger.Run()
	if err != nil {
		log.Println(err)
	}
	return
}

func sortByPing(list []VPN) []VPN {
	debug("sorting by ping")

	var wg sync.WaitGroup
	var n int32
	for i := range list {
		if n > 512 {
			time.Sleep(time.Millisecond * 20)
		}

		wg.Add(1)
		atomic.AddInt32(&n, 1)
		go func(i int) {
			defer atomic.AddInt32(&n, -1)
			defer wg.Done()
			list[i].ping = doPing(list[i].IP)
		}(i)
	}
	wg.Wait()

	sort.Slice(list, func(i, j int) bool {
		t1 := list[i].ping
		t2 := list[j].ping
		return t1 < t2 && t1 > 0
	})
	return list
}
