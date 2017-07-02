package main

import (
	"archive/zip"
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/asticode/go-astilectron"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

const (
	Username          = "YOUR_NORDVPN_USERNAME"
	Password          = "YOUR_NORDVPN_PASSWORD"
	APIKey            = "YOUR_GOOGLE_API_KEY"
	OutHTML           = "map.html"
	OutConfig         = "/configurations"
	OutDatabase       = "./NordVPN.db"
	DatabaseStructure = `
		CREATE TABLE vpnlist (
			file VARCHAR(128) PRIMARY KEY,
			ip VARCHAR(15) NOT NULL,
			port int NULL,
			tcp BOOLEAN NOT NULL,
			udp BOOLEAN NOT NULL
		);
		`
	MapBody = `<!DOCTYPE html>
<html>
  <head>
    <meta name="viewport" content="initial-scale=1.0, user-scalable=no">
    <meta charset="utf-8">
    <title>NordVPN Servers</title>
    <style>
			.top-titlebar {
  			position: absolute;
			  left: 0px;
			  top: 0px;
			  width: 100%;
			  height: 32px;
			  background-color: #7a7c7c;
			  -webkit-user-select: none;
			  -webkit-app-region: drag;
			}

      #map {
        height: 100%;
      }

      html, body {
        height: 100%;
        margin: 0;
        padding: 0;
      }
    </style>
  </head>
  <body>
    <div id="map"></div>
    <script>`
	MapFooter = `
      }
    </script>
    <script async defer
    	src="https://maps.googleapis.com/maps/api/js?key=` + APIKey + `&callback=initMap">
    </script>
  </body>
</html>`
)

var (
	InitMark = `function httpGet(ip) {
      var http = new XMLHttpRequest();

      http.open("GET", "http://localhost:8084/nord?ip="+ip, true);
      http.send();                                     
    }

		function addMark(map, location, title, ip) {
      var marker = new google.maps.Marker({
        position: location,
        map: map,
        title: title,
        ip: ip
      });

      marker.addListener('click', function() {
        httpGet(marker.ip);
      });
    }

		function initMap() {
      var myLatLng = {lat: {{.Latitude}}, lng: {{.Longitude}} };

      var map = new google.maps.Map(document.getElementById('map'), {
       	zoom: 4,
	      center: myLatLng
  	  });

    	var marker = new google.maps.Marker({
        position: myLatLng,
       	map: map,
	      title: '{{.Name}}',
				ip: '{{.IP}}'
  	  });
			`

	AddMark = `
	addMark(map, {lat: {{.Latitude}}, lng: {{.Longitude}} }, '{{.Name}}', '{{.IP}}');`
)

type Categories struct {
	Name string `json:"name"`
}

type Location struct {
	Lat  float32 `json:"lat"`
	Long float32 `json:"long"`
}

type Features struct {
	Ikev2         bool `json:"ikev2"`
	OpenvpnUDP    bool `json:"openvpn_udp"`
	OpenvpnTCP    bool `json:"openvpn_tcp"`
	Socks         bool `json:"socks"`
	Proxy         bool `json:"proxy"`
	Pptp          bool `json:"pptp"`
	L2tp          bool `json:"l2tp"`
	OpevpnXORudp  bool `json:"openvpn_xor_udp"`
	OpenvpnXORtcp bool `json:"openvpn_xor_tcp"`
}

type Data struct {
	Id        int          `json:"id"`
	IPaddress string       `json:"ip_address"`
	Keywords  []string     `json:"search_keywords"`
	Cat       []Categories `json:"categories"`
	Name      string       `json:"name"`
	Domain    string       `json:"domain"`
	Price     int          `json:"price"`
	Flag      string       `json:"flag"`
	Country   string       `json:"country"`
	Loc       Location     `json:"location"`
	Load      int          `json:"load"`
	Feature   Features     `json:"features"`
}

type DataScript struct {
	Latitude  float32
	Longitude float32
	Name      string
	IP        string
}

type VPNList struct {
	file string
	ip   string
	port int
	tcp  bool
	udp  bool
}

func downloadFiles(basedir string) error {
	res, err := http.Get("https://nordvpn.com/api/files/zip")
	if err != nil {
		return err
	}

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(basedir+"/files.zip", content, 0666); err != nil {
		return err
	}

	return nil
}

func unzipFile(outdir, file string) error {
	reader, err := zip.OpenReader(file)
	if err != nil {
		return err
	}
	defer reader.Close()

	for i := range reader.File {
		r, err := reader.File[i].Open()
		if err != nil {
			return err
		}

		content, err := ioutil.ReadAll(r)
		if err != nil {
			r.Close()
			return err
		}

		var name string = outdir + string(os.PathSeparator) + reader.File[i].Name
		var mode os.FileMode = reader.File[i].Mode()

		if err = ioutil.WriteFile(name, content, mode); err != nil {
			r.Close()
			return err
		}

		r.Close()
	}

	return nil
}

func getFileConfiguration(filename string) (list VPNList, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return list, err
	}
	defer file.Close()

	list.file = filepath.Base(filename)
	list.tcp = false
	list.udp = false

	reader := bufio.NewReader(file)

	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}

		s := strings.SplitN(string(line), " ", -1)

		if strings.Compare(s[0], "proto") == 0 {
			if strings.Compare(s[1], "tcp") == 0 {
				list.tcp = true
			} else if strings.Compare(s[1], "udp") == 0 {
				list.udp = true
			}
		} else if strings.Compare(s[0], "remote") == 0 {
			list.ip = s[1]
			list.port, _ = strconv.Atoi(s[2])
		}
	}

	return list, nil
}

func configureDatabase(basedir string) (err error) {
	db, err := sql.Open("sqlite3", OutDatabase)
	if err != nil {
		return
	}

	_, err = db.Exec(DatabaseStructure)
	if err != nil {
		return
	}

	dir, err := ioutil.ReadDir(basedir)
	if err != nil {
		return
	}

	for i := range dir {
		vpnlist, err := getFileConfiguration(basedir + string(os.PathSeparator) + dir[i].Name())
		if err != nil {
			return err
		}

		_, err = db.Exec("INSERT INTO vpnlist (file, ip, port, tcp, udp) VALUES (?, ?, ?, ?, ?)",
			vpnlist.file, vpnlist.ip, vpnlist.port, vpnlist.tcp, vpnlist.udp)
		if err != nil {
			return err
		}
	}

	return
}

func selectMap(ip string) (string, error) {
	db, err := sql.Open("sqlite3", OutDatabase)
	if err != nil {
		return "", err
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		return "", err
	}

	var file string

	rows, err := db.Query("SELECT file FROM vpnlist WHERE ip = '" + ip + "'")
	if err != nil {
		return "", err
	}

	for rows.Next() {
		rows.Scan(&file)
	}

	return file, nil
}

func createMap() error {
	res, err := http.Get("http://api.nordvpn.com/server")
	if err != nil {
		return err
	}

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	data := make([]Data, 0)

	err = json.Unmarshal(content, &data)
	if err != nil {
		return err
	}

	if _, err := os.Stat(OutHTML); err == nil {
		os.Remove(OutHTML)
	}

	file, err := os.OpenFile(OutHTML, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	file.Write([]byte(MapBody))

	for i := range data {
		script := DataScript{
			data[i].Loc.Lat,
			data[i].Loc.Long,
			data[i].Name,
			data[i].IPaddress,
		}

		if i == 0 {
			tf := template.Must(template.New("").Parse(string(InitMark)))
			tf.Execute(file, script)
			continue
		}

		t := template.Must(template.New("html").Parse(string(AddMark)))
		t.Execute(file, script)
	}

	file.Write([]byte(MapFooter))

	return nil
}

func execOpenVPN(file string) error {
	err := ioutil.WriteFile(os.TempDir()+"/auth.txt", []byte(Username+"\n"+Password+"\n"), 0666)
	if err != nil {
		return err
	}

	exec.Command("killall", "openvpn").Run()

	cmd := exec.Command("openvpn", "--config", file, "--auth-user-pass", os.TempDir()+"/auth.txt")

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}

	os.Remove(os.TempDir() + "/auth.txt")

	return nil
}

func startWebServer(basedir string) {
	http.HandleFunc("/nord", func(w http.ResponseWriter, r *http.Request) {
		ip := r.FormValue("ip")

		file, err := selectMap(ip)
		if err != nil {
			panic(err)
		}

		fmt.Println("Using '" + file + "' configuartion file")

		err = execOpenVPN(basedir + string(os.PathSeparator) +
			OutConfig + string(os.PathSeparator) + file)
		if err != nil {
			panic(err)
		}
	})
	http.ListenAndServe("127.0.0.1:8084", nil)
}

func main() {
	if os.Getuid() != 0 {
		fmt.Println("Execute this program as root")
		os.Exit(1)
	}

	basedir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Configure options for launcher
	handler, err := astilectron.New(astilectron.Options{
		AppName:            "GoNordVPN",
		AppIconDefaultPath: basedir + "/logo.png",
		AppIconDarwinPath:  "<your .icns icon>",
		BaseDirectoryPath:  basedir,
	})
	if err != nil {
		panic(err)
	}
	defer handler.Close()

	// Start library
	if err = handler.Start(); err != nil {
		panic(err)
	}

	if err = createMap(); err != nil {
		panic(err)
	}

	var fullscreen = true
	var transparent = true
	var barStyle = "hidden-inset"
	var frame = false

	// Configuring window
	var w *astilectron.Window
	if w, err = handler.NewWindow("./map.html", &astilectron.WindowOptions{
		Center:        astilectron.PtrBool(true),
		Height:        astilectron.PtrInt(600),
		Width:         astilectron.PtrInt(600),
		Fullscreen:    &fullscreen,
		Transparent:   &transparent,
		TitleBarStyle: &barStyle,
		Frame:         &frame,
	}); err != nil {
		panic(err)
	}
	defer func() {
		if err := w.Close(); err != nil {
			panic(err)
		}
	}()

	fmt.Println("Creating window")
	if err = w.Create(); err != nil {
		panic(err)
	}

	// check the output directory with configuration files
	if _, err = os.Stat(basedir + OutConfig); err != nil {
		if err = os.Mkdir(basedir+OutConfig, 0777); err != nil {
			panic(err)
		}

		fmt.Println("Downloading zip file with configurations...")
		err = downloadFiles(basedir + OutConfig)
		if err != nil {
			panic(err)
		}

		fmt.Println("Unzipping files...")
		err = unzipFile(basedir+OutConfig,
			basedir+OutConfig+string(os.PathSeparator)+"files.zip")
		if err != nil {
			panic(err)
		}
	}

	if _, err = os.Stat(OutDatabase); err != nil {
		fmt.Println("Creating database configuration...")
		err = configureDatabase(basedir + OutConfig)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Starting web server...")
	go startWebServer(basedir)

	handler.Wait()
}
