# GoNordVPN
GUI NordVPN client created in Golang for Linux.
(Mainly for Linux)

Sponsored by Initial-D soundtracks.

Created by Mester. Lazy computer engineering student

Installation:
-------------
1.- Install Golang.

If you're using Arch Linux execute:

`sudo pacman -S go`

Else install from `golang.org/dl`

2.- Get your google maps API key.

You should go to: `https://console.developers.google.com`

And get your Javascript API Key

3.- Install Golang libraries:

`go get -u github.com/asticode/go-astilectron`

`go get -u github.com/mattn/go-sqlite3`

4.- Configure:

In code, you should change the fields with your own data.

`Username = "YOUR_NORDVPN_USERNAME"`

`Password = "YOUR_NORDVPN_PASSWORD"`

`APIKey = "YOUR_GOOGLE_API_KEY"`

5.- Compiling:
- `go build main.go`

6.- Execute:

`sudo ./main`

The first executing will be slower. Be patient!
