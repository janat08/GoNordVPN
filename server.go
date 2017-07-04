package main

import (
	"database/sql"
	"errors"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

var (
	PIDFile     string
	OutConfig   string
	OutDatabase string
	fileUser    string
)

func selectMap(ip string, tcp bool) (string, error) {
	if len(OutDatabase) == 0 {
		return "", errors.New("No database has provided")
	}
	db, err := sql.Open("sqlite3", OutDatabase)
	if err != nil {
		return "", err
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		return "", err
	}

	var file string
	var query string

	if tcp {
		query = "SELECT file FROM vpnlist WHERE ip = '" + ip + "' AND udp = 0 AND tcp = 1"
	} else {
		query = "SELECT file FROM vpnlist WHERE ip = '" + ip + "' AND udp = 1 AND tcp = 0"
	}

	rows, err := db.Query(query)
	if err != nil {
		return "", err
	}

	for rows.Next() {
		rows.Scan(&file)
	}

	return file, nil
}

func execOpenVPN(file string) error {
	if len(fileUser) == 0 {
		return errors.New("You should enter file with user and password")
	}

	stopOpenVPN()

	// OpenDNS servers
	ioutil.WriteFile("/etc/resolv.conf", []byte("nameserver 208.67.222.222\nnameserver 208.67.220.220\n"), 0644)

	cmd := exec.Command("openvpn", "--config", file, "--auth-user-pass", fileUser)

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func connectDBUS() error {

	return nil
}

func stopOpenVPN() {
	exec.Command("killall", "openvpn").Run()
}

func startWebServer() {
	if len(OutConfig) == 0 {
		return
	}

	if len(PIDFile) == 0 {
		PIDFile = os.TempDir() + string(os.PathSeparator) + "nordvpn.pid"
	}

	ioutil.WriteFile(PIDFile, []byte(strconv.Itoa(os.Getpid())), 0666)

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL, syscall.SIGABRT, syscall.SIGTERM)

	http.HandleFunc("/nord", func(w http.ResponseWriter, r *http.Request) {
		var tcp bool

		ip := r.FormValue("ip")
		tcpForm := r.FormValue("tcp")

		if strings.Contains(tcpForm, "true") {
			tcp = true
		} else {
			tcp = false
		}

		if ip == "0" {
			stopOpenVPN()
		} else {
			file, err := selectMap(ip, tcp)
			if err != nil {
				panic(err)
			}

			err = execOpenVPN(OutConfig + string(os.PathSeparator) + file)
			if err != nil {
				panic(err)
			}
		}
	})

	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		stopOpenVPN()
	})

	http.HandleFunc("/exit", func(w http.ResponseWriter, r *http.Request) {
		stopOpenVPN()

		os.Remove(PIDFile)
		os.Remove(fileUser)
		os.Exit(0)
	})

	go http.ListenAndServe("127.0.0.1:8084", nil)

	<-signals
	os.Remove(PIDFile)
	os.Remove(fileUser)

	stopOpenVPN()
}

func main() {
	if os.Getuid() != 0 {
		println("You should execute this program has root")
		os.Exit(1)
	}

	flag.StringVar(&OutDatabase, "database", "", "Database location")
	flag.StringVar(&fileUser, "file", "", "File with user and password")
	flag.StringVar(&OutConfig, "config", "", "OpenVPN directory with configuration files")
	flag.StringVar(&PIDFile, "pid", "", "PID file for server")

	flag.Parse()

	if _, err := os.Stat(PIDFile); err == nil {
		println("Server is already running")
		os.Exit(1)
	}

	startWebServer()
}
