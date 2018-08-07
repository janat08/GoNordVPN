#!/bin/bash

go get ./...
go build -o GoNordVPN
currentPath=`pwd`
dirPath=`basename $currentPath`
cd
mv $currentPath /usr/local/share/$dirPath
ln -s /usr/local/share/$dirPath/GoNordVPN /usr/local/bin/gonordvpn
