package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/valyala/fastrand"
)

func doModprobe() error {
	return exec.Command("modprobe", "tun").Run()
}

func stopOpenVPN() error {
	err := exec.Command("pkill", "-9", "openvpn").Run()
	if err == nil {
		err = restoreResolvConf()
	}
	return err
}

func backupResolvConf() error {
	bkp, err := os.Create("/etc/.resolv.conf")
	if err != nil {
		return err
	}
	defer bkp.Close()

	rev, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return err
	}
	defer rev.Close()

	_, err = io.Copy(bkp, rev)
	return err
}

func restoreResolvConf() error {
	bkp, err := os.Open("/etc/.resolv.conf")
	if err != nil {

	}
	defer os.Remove("/etc/.resolv.conf")
	defer bkp.Close()

	rev, err := os.Create("/etc/resolv.conf")
	if err != nil {
		return err
	}
	defer rev.Close()

	_, err = io.Copy(rev, bkp)
	return err
}

func startOpenVPN(country, proto string) error {
	err := doModprobe()
	if err != nil {
		return err
	}

	port := 0
	if proto == "tcp" {
		port = 443
	} else {
		port = 1194
	}

	var vpn = VPN{
		ping: time.Duration(1 << 31),
	}
	for _, vpn2 := range config.VPNList {
		if vpn2.Country == country && vpn2.ping < vpn.ping {
			vpn = vpn2
		}
	}
	currentServer = &vpn
	conFile := path.Join(*ovpnDir, fmt.Sprintf("%s.%s%d.ovpn", vpn.Domain, proto, port))

	authFile := randFile()
	file, err := os.OpenFile(authFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0700)
	if err != nil {
		return err
	}
	fmt.Fprintf(file, "%s\n", config.Username)
	fmt.Fprintf(file, "%s\n", config.Password)
	file.Close()
	debug(
		fmt.Sprintf("executing openvpn --config %s --auth-user-pass %s --auth-nocache", conFile, authFile),
	)
	go func() {
		// backing up resolv.conf
		if err = backupResolvConf(); err != nil {
			log.Println(err)
			return
		}

		file, err := os.Create("/etc/resolv.conf")
		if err != nil {
			log.Println(err)
			return
		}
		for _, dns := range strings.Split(*dnsServers, ",") {
			fmt.Fprintf(file, "nameserver %s\n", dns)
		}
		file.Close()

		for {
			cmd := exec.Command("openvpn", "--config", conFile, "--auth-user-pass", authFile, "--auth-nocache")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Start()
			if err := cmd.Wait(); err != nil {
				break
			}
		}
		os.Remove(authFile)
	}()
	return err
}

func randFile() string {
	return path.Join(os.TempDir(), randStr())
}

func randStr() string {
	s := ""
	var str = "qwruioxdcfvgbmjzxcvbnmWTRSCGHBNSDUVYUIB"
	for i := 0; i < 10; i++ {
		s += string(str[fastrand.Uint32n(uint32(len(str)))])
	}
	return s
}
