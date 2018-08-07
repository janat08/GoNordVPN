package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/erikdubbelboer/fasthttp"
)

type VPN struct {
	ID             int           `json:"id"`
	IP             string        `json:"ip_address"`
	SearchKeywords []interface{} `json:"search_keywords"`
	Categories     []struct {
		Name string `json:"name"`
	} `json:"categories"`
	Name     string `json:"name"`
	Domain   string `json:"domain"`
	Price    int    `json:"price"`
	Flag     string `json:"flag"`
	Country  string `json:"country"`
	Location struct {
		Lat  float64 `json:"lat"`
		Long float64 `json:"long"`
	} `json:"location"`
	Load     int `json:"load"`
	Features struct {
		Ikev2            bool `json:"ikev2"`
		OpenvpnUDP       bool `json:"openvpn_udp"`
		OpenvpnTCP       bool `json:"openvpn_tcp"`
		Socks            bool `json:"socks"`
		Proxy            bool `json:"proxy"`
		Pptp             bool `json:"pptp"`
		L2Tp             bool `json:"l2tp"`
		OpenvpnXorUDP    bool `json:"openvpn_xor_udp"`
		OpenvpnXorTCP    bool `json:"openvpn_xor_tcp"`
		ProxyCybersec    bool `json:"proxy_cybersec"`
		ProxySsl         bool `json:"proxy_ssl"`
		ProxySslCybersec bool `json:"proxy_ssl_cybersec"`
	} `json:"features"`
}

type VPNS struct {
	Username string
	Password string
	VPNList  []VPN
}

func fetchVPNList() error {
	debug("fetching vpn list")
	os.Remove(*dataFile)

	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)
	req.SetRequestURI("https://api.nordvpn.com/server")

	err := client.Do(req, res)
	if err != nil {
		return err
	}
	vpns := make([]VPN, 0)
	err = json.Unmarshal(res.Body(), &vpns)
	if err != nil {
		return err
	}
vpnLabel:
	for _, vpn := range vpns {
		for _, vpn2 := range config.VPNList {
			if vpn.Domain == vpn2.Domain || vpn.Country == vpn2.Country {
				continue vpnLabel
			}
		}
		config.VPNList = append(config.VPNList, vpn)
	}
	vpns = nil
	file, err := os.Create(*dataFile)
	if err == nil {
		res.BodyWriteTo(file)
		file.Close()
	}
	return nil
}

func fetchLocal() error {
	data, err := ioutil.ReadFile(*dataFile)
	if err != nil {
		return err
	}

	vpns := make([]VPN, 0)
	err = json.Unmarshal(data, &vpns)
	if err != nil {
		return err
	}
vpnLabel:
	for _, vpn := range vpns {
		for _, vpn2 := range config.VPNList {
			if vpn.Domain == vpn2.Domain || vpn.Country == vpn2.Country {
				continue vpnLabel
			}
		}
		config.VPNList = append(config.VPNList, vpn)
	}
	return nil
}

var client = fasthttp.Client{
	Name:           "GoNordVPN 0.1",
	ReadBufferSize: 1 << 22,
}

func downloadZip(dst string) error {
	var err error
	debug("Downloading VPN information")
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)
	url := "https://nordvpn.com/api/files/zip"

	for {
		req.Reset()
		res.Reset()
		req.SetRequestURI(url)

		err := client.Do(req, res)
		if err != nil {
			break
		}
		if res.StatusCode() > 200 && res.StatusCode() < 400 {
			url = string(res.Header.Peek("Location"))
			continue
		}
		if res.StatusCode() != 200 {
			err = fmt.Errorf("error getting zip. Code: %d", res.StatusCode())
			break
		}
		file, err := os.Create(dst)
		if err == nil {
			err = res.BodyWriteTo(file)
			file.Close()
		}
		break
	}
	return err
}

func unZip(src string) (err error) {
	debug("unzipping files")

	deleteOVPNFiles()

	var reader *zip.ReadCloser
	reader, err = zip.OpenReader(src)
	if err != nil {
		return
	}

	var br io.ReadCloser
	var bw io.WriteCloser
	buf := make([]byte, 512, 512)
	for _, file := range reader.File {
		br, err = file.Open()
		if err != nil {
			return
		}
		bw, err = os.Create(
			filepath.Join(*ovpnDir, file.Name),
		)
		if err != nil {
			br.Close()
			return
		}
		_, err = io.CopyBuffer(bw, br, buf)
		br.Close()
		bw.Close()
		if err != nil {
			break
		}
	}
	if err == io.EOF {
		err = nil
	}
	buf = nil
	return
}

func deleteOVPNFiles() {
	files, err := ioutil.ReadDir(*ovpnDir)
	if err != nil {
		return
	}
	for _, file := range files {
		os.Remove(filepath.Join(*ovpnDir, file.Name()))
	}
}

func fetchData(fetchNew bool) error {
	debug("fetching data")
	if !fetchNew {
		return fetchLocal()
	}
	path := filepath.Join(os.TempDir(), "nordvpn.zip")
	err := downloadZip(path)
	if err == nil {
		err = unZip(path)
	}
	if err == nil {
		os.Remove(path)
		err = fetchVPNList()
	}
	return err
}

func checkFiles() bool {
	_, err := os.Stat(*dataFile)
	if err != nil {
		return true
	}
	files, err := ioutil.ReadDir(*ovpnDir)
	if err != nil {
		return true
	}
	return len(files) == 0
}
