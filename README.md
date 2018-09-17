# GoNordVPN
GUI NordVPN client in Golang for Linux.

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
sudo -E ./install.sh
```

If you have Go command:

```bash
cd /tmp/
git clone https://github.com/dgrr/GoNordVPN
cd GoNordVPN
sudo -E ./install.sh
```

Usage:
------

```bash
gonordvpn -u your@email.com -fetch
```
```bash
xdg-open http://localhost:9114
```
