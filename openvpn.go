package main

import (
	"fmt"
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
	return exec.Command("pkill", "-9", "openvpn").Run()
}

func startOpenVPN(country, proto string) error {
	err := doModprobe()
	if err != nil {
		return err
	}

	file, err := os.Create("/etc/resolv.conf")
	if err != nil {
		return err
	}
	for _, dns := range strings.Split(*dnsServers, ",") {
		fmt.Fprintf(file, "nameserver %s\n", dns)
	}
	file.Close()

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
	file, err = os.OpenFile(authFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0700)
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
		for {
			cmd := exec.Command("openvpn", "--config", conFile, "--auth-user-pass", authFile, "--auth-nocache")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Start()
			time.Sleep(time.Second * 5)
			os.Remove(authFile)
			if err := cmd.Wait(); err != nil {
				break
			}
		}
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
