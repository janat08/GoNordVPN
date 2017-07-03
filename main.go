package main

import (
	"archive/zip"
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/asticode/go-astilectron"
	"github.com/howeyc/gopass"
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
      #map {
        height: 100%;
      }

      html, body {
        height: 100%;
        margin: 0;
        padding: 0;
      }

			#over_map {
				position: absolute;
				top: 10px;
				left: 10px;
				z-index: 99;
			}
    </style>
  </head>
  <body>
    <div id="map">
		</div>
		<div id="over_map">
			<img src="logo.png" style="position: fixed; right: 1em; margin: 0 0 10px 10px;width:auto;height:auto;">
		</div>
    <script>`
	InitMark = `function httpGet(ip) {
      var http = new XMLHttpRequest();

      http.open("GET", "http://localhost:8084/nord?ip="+ip, true);
      http.send();                                     
    }

		var contains = function(needle) {
			var findNaN = needle !== needle;
			var indexOf;

			if(!findNaN && typeof Array.prototype.indexOf === 'function') {
				indexOf = Array.prototype.indexOf;
			} else {
				indexOf = function(needle) {
        	var i = -1, index = -1;

            for(i = 0; i < this.length; i++) {
                var item = this[i];

                if((findNaN && item !== item) || item === needle) {
                    index = i;
                    break;
                }
            }

            return index;
        	};
    	}

    	return indexOf.call(this, needle) > -1;
		};

		var marks = [""];
		var lastMark;

		function addMark(map, location, title, ip) {
			if ( !marks.includes(title) ) {
	      var marker = new google.maps.Marker({
  	      position: location,
    	    map: map,
      	  title: title,
        	ip: ip
	      });

  	    marker.addListener('click', function() {
					if ( lastMark != null )
						lastMark.setIcon('http://maps.google.com/mapfiles/ms/icons/red-dot.png')
					lastMark = marker;
					marker.setIcon('http://maps.google.com/mapfiles/ms/icons/blue-dot.png')
  	      httpGet(marker.ip);
    	  });

				marks.push(title);
    	}
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

var (
	Username    string
	Password    string
	APIKey      string
	OutHTML     string
	OutConfig   string
	OutDatabase string
	Basedir     string
	PIDFile     = os.TempDir() + string(os.PathSeparator) + "nordvpn.pid"
	AuthFile    = os.TempDir() + string(os.PathSeparator) + "authNord.txt"
	MapFooter   = `
      }
    </script>
    <script async defer
    	src="https://maps.googleapis.com/maps/api/js?key=` + APIKey + `&callback=initMap">
    </script>
  </body>
</html>`
)

type Config struct {
	Username    string
	Password    string
	APIKey      string
	OutHTML     string
	OutConfig   string
	OutDatabase string
	Basedir     string
}

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

	file, err := os.OpenFile(OutHTML, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	file.Write([]byte(MapBody))

	for i := range data {
		s := strings.Split(data[i].Name, " ")

		data[i].Name = s[0]
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

func createConfig(stdin bool, configuration Config) {
	var err error

	if stdin {
		in := bufio.NewReader(os.Stdin)

		fmt.Print("Username: ")
		username, _, err := in.ReadLine()
		if err != nil {
			panic(err)
		}
		configuration.Username = string(username)

		fmt.Print("Password: ")
		pass, err := gopass.GetPasswdMasked()
		if err != nil {
			panic(err)
		}
		configuration.Password = string(pass)

		fmt.Print("API Key: ")
		api, _, err := in.ReadLine()
		if err != nil {
			panic(err)
		}
		configuration.APIKey = string(api)
	}

	conf, err := json.Marshal(configuration)
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile("GoNordVPN.conf", conf, 0666)
}

func getConfig(file string) {
	var configuration Config

	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &configuration)
	if err != nil {
		panic(err)
	}

	Username = configuration.Username
	Password = configuration.Password
	APIKey = configuration.APIKey
	OutHTML = configuration.OutHTML
	OutConfig = configuration.OutConfig
	OutDatabase = configuration.OutDatabase
	Basedir = configuration.Basedir
}

func startWebServer(basedir string) {
	ioutil.WriteFile(AuthFile, []byte(Username+"\n"+Password+"\n"), 0600)

	exec.Command("nordvpn-server", "-database", OutDatabase, "-file", AuthFile, "-config", OutConfig, "-pid", PIDFile).Start()
}

func stopServer() {
	http.Get("http://localhost:8084/stop")
}

func killServer() {
	http.Get("http://localhost:8084/exit")
}

func main() {
	var config string
	var confStruct Config

	kill := flag.Bool("kill", false, "Kill server process")
	stop := flag.Bool("stop", false, "Stop OpenVPN process")
	start := flag.Bool("start", false, "Start server process (requires root)")
	create := flag.Bool("make-config", false, "Creates configuration file json")
	useStdin := flag.Bool("stdin", false, "Use stdin to configure file")
	flag.StringVar(&confStruct.OutHTML, "out-html", "", "File output map")
	flag.StringVar(&confStruct.Basedir, "basedir", "", "Working directory")
	flag.StringVar(&config, "config", "/etc/GoNordVPN.conf", "Configuration file in json format")
	flag.StringVar(&confStruct.OutConfig, "out-config", "", "Folder with OpenVPN configuration files")
	flag.StringVar(&confStruct.OutDatabase, "out-db", "", "Database with specific data of configuration files")

	flag.Parse()

	if *kill {
		killServer()
		os.Exit(0)
	}

	if *stop {
		stopServer()
		os.Exit(0)
	}

	// Create configuration file and exit
	if *create {
		createConfig(*useStdin, confStruct)
		os.Exit(0)
	}

	// Check if configuration file exists
	if _, err := os.Stat(config); err != nil {
		fmt.Println("You should provide a configuration file")
		os.Exit(1)
	}

	// Getting configuration from file
	getConfig(config)

	if *start {
		if os.Getuid() != 0 {
			fmt.Println("Execute this program as root to start server")
			os.Exit(1)
		}

		startWebServer(Basedir)
		os.Exit(0)
	}

	var initServer bool = false

	// Check if server is running
	if _, err := os.Stat(PIDFile); err != nil {
		initServer = true
		if os.Getuid() != 0 {
			fmt.Println("Execute this program or use option '-start' as root to start server")
			os.Exit(1)
		}
	}

	if len(Basedir) == 0 {
		Basedir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		fmt.Println("Using:", Basedir)
	}

	// Configure options for launcher
	handler, err := astilectron.New(astilectron.Options{
		AppName:            "GoNordVPN",
		AppIconDefaultPath: Basedir + "/logo.png",
		AppIconDarwinPath:  "<your .icns icon>",
		BaseDirectoryPath:  Basedir,
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

	var fullscreen = false
	var transparent = true
	var barStyle = "hidden-inset"
	var frame = true

	// Configuring window
	var w *astilectron.Window
	if w, err = handler.NewWindow(OutHTML, &astilectron.WindowOptions{
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
	if _, err = os.Stat(OutConfig); err != nil {
		if err = os.Mkdir(OutConfig, 0777); err != nil {
			panic(err)
		}

		fmt.Println("[*] Downloading zip file with configurations...")
		err = downloadFiles(OutConfig)
		if err != nil {
			panic(err)
		}
		fmt.Println("[*] Downloaded")

		fmt.Println("[*] Unzipping files...")
		err = unzipFile(OutConfig,
			OutConfig+string(os.PathSeparator)+"files.zip")
		if err != nil {
			panic(err)
		}
		fmt.Println("[*] Unzipped")
	}

	if _, err = os.Stat(OutDatabase); err != nil {
		fmt.Println("Creating database configuration...")
		err = configureDatabase(OutConfig)
		if err != nil {
			panic(err)
		}
		fmt.Pritln("[*] Database created")
	}

	if initServer {
		fmt.Println("Starting web server...")
		go startWebServer(Basedir)
	}

	handler.HandleSignals()

	handler.Wait()
}
