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

3.- Configure and compile: 

`sudo -E ./install.sh`

4.- Execute:

`sudo nordvpn-client -start && nordvpn-client`

(Only requires root to start nordvpn-server)

The first execution will be slower. Be patient!
