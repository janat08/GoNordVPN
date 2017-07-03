#!/bin/bash

if [ `whoami` != 'root' ]; then
  echo "Execute as root";
  exit;
fi

go get -u github.com/howeyc/gopass
go get -u github.com/asticode/go-astilectron
go get -u github.com/mattn/go-sqlite3

go build main.go && go build server.go

mkdir /usr/share/GoNordVPN
mv main /usr/share/GoNordVPN/nordvpn-client
mv server /usr/share/GoNordVPN/nordvpn-server
cp logo.png /usr/share/GoNordVPN

/usr/share/GoNordVPN/nordvpn-client -make-config -out-html /usr/share/GoNordVPN/map.html -out-config /usr/share/GoNordVPN/configurations -out-db /usr/share/GoNordVPN/NordVPN.db -basedir /usr/share/GoNordVPN -stdin

mv GoNordVPN.conf /etc/GoNordVPN.conf

chmod 777 -R /usr/share/GoNordVPN

ln -s /usr/share/GoNordVPN/nordvpn-client /usr/bin/nordvpn-client
ln -s /usr/share/GoNordVPN/nordvpn-server /usr/bin/nordvpn-server
