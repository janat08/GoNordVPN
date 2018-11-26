
.PHONY: clean install uninstall

INSTALL_DIR=/usr/share/GoNordVPN/
GO = go
GFLAGS = -ldflags '-X main.InstallPath=${INSTALL_DIR}'

gonordvpn:
	$(GO) build $(GFLAGS) -o $@

clean:
	rm -f gonordvpn

install: gonordvpn
	mkdir /usr/share/GoNordVPN
	cp -r `pwd`/* ${INSTALL_DIR}
	ln -s ${INSTALL_DIR}/gonordvpn /usr/bin/gonordvpn
	cp init/openrc/gonordvpn /etc/init.d/
	chmod +x /etc/init.d/gonordvpn
	rc-update add gonordvpn default
	rc-service gonordvpn start

uninstall: clean
	rc-service gonordvpn stop
	rc-update del gonordvpn default
	rm -f /etc/init.d/gonordvpn
	unlink /usr/bin/gonordvpn
	rm -rf /usr/share/GoNordVPN
