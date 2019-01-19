# GoNordVPN
GUI NordVPN client in Golang for Linux, like [nordpy](https://github.com/morpheusthewhite/NordPy) (worse looking, easier to install).

Created by lazy computer science student

![alt text](https://raw.githubusercontent.com/dgrr/GoNordVPN/master/gui.png)

Installation:
-------------

If you don't have Go command:

1. Download GoNordVPN from [latest releases](https://github.com/dgrr/GoNordVPN/releases/)

2. Execute:

```bash
unzip GoNordVPN.zip
cd GoNordVPN/
go get ./...
make && sudo make install
sudo vim /usr/share/GoNordVPN/nordvpn.conf # Edit your config with your credentials.
rc-service gonordvpn restart
```

If you have Go command:

```bash
cd /tmp/
git clone https://github.com/dgrr/GoNordVPN
cd GoNordVPN
go get ./...
make && sudo make install
sudo vim /usr/share/GoNordVPN/nordvpn.conf # Edit your config with your credentials.
rc-service gonordvpn restart
```

Usage:
------

Cmd usage:
```bash
gonordvpn -u your@email.com -fetch
```
OpenRC usage:
```bash
rc-service gonordvpn start
```

```bash
xdg-open http://localhost:9114
```
