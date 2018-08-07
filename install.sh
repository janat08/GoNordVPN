#!/bin/bash

go get ./...
go build -o GoNordVPN
currentPath=`pwd`
dirPath=`basename $currentPath`
cd
cp -r $currentPath /usr/local/share/$dirPath
cat <<EOF > /usr/local/bin/gonordvpn
#!/bin/bash
cd "/usr/local/share/$dirPath/" && ./GoNordVPN "\$@"
EOF
chmod +x /usr/local/bin/gonordvpn
