package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/howeyc/gopass"
)

func debug(str string) {
	if *verbose {
		log.Println(str)
	}
}

var (
	fetch       = flag.Bool("fetch", false, "Fetch VPN server list and OVPN files")
	verbose     = flag.Bool("v", false, "Verbose mode")
	certFile    = flag.String("cert", "", "SSL Certificate")
	keyFile     = flag.String("key", "", "SSL private key for certificate")
	logfile     = flag.String("o", "", "Log file (default stdout)")
	ovpnDir     = flag.String("ovpn-dir", "./ovpn-files/", "OVPN output file directory")
	dataFile    = flag.String("data", "./servers.json", "VPN Servers data")
	templateDir = flag.String("t", "./templates/*", "Templates files dir")
	httpDir     = flag.String("html-dir", "./html", "HTML file directory")
	dnsServers  = flag.String("dns", "8.8.8.8,8.8.4.4", "DNS servers separated by commas")
	disableRoot = flag.Bool("no-root", false, "Disables root checking")
	username    = flag.String("u", "", "NordVPN username")

	config = VPNS{
		VPNList: make([]VPN, 0),
	}
)

func init() {
	flag.Parse()
	os.MkdirAll(path.Dir(*ovpnDir), 700)
	os.MkdirAll(path.Dir(*dataFile), 700)
}

func main() {
	if !*disableRoot && os.Getuid() != 0 {
		log.Fatalln("Must be executed as root because of openvpn")
	}
	if *logfile != "" {
		file, err := os.Create(*logfile)
		if err != nil {
			log.Fatalln(err)
		}
		defer file.Close()
		log.SetOutput(file)
	}
	if (*certFile != "" && *keyFile == "") ||
		(*certFile == "" && *keyFile != "") {
		log.Fatalln("Certfile and Keyfile must be provided")
	}

	if *username == "" {
		log.Fatalln("username must be specified with -u parameter")
	}
	fmt.Printf("%s password: ", *username)
	pass, err := gopass.GetPasswd()
	if err != nil {
		log.Fatalln(err)
	}

	config.Username = *username
	config.Password = string(pass)
	pass = nil

	if !*fetch {
		*fetch = checkFiles()
	}
	err = fetchData(*fetch)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(startServer())
	stopOpenVPN()
}
