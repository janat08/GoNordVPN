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

4.- Configure and compile: 

`sudo -E ./install.sh`

5.- Execute:

`sudo nordvpn-client`

The first execution will be slower. Be patient!
