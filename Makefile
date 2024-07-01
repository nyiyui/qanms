src = .
flags = -race -tags sdnotify

clean:
	rm coord-server
	rm device-client
	rm gen-keys

coord-server:
	go build ${flags} ${src}/cmd/coord-server

device-client:
	go build ${flags} ${src}/cmd/device-client

gen-keys:
	go build ${flags} ${src}/cmd/gen-keys

install-coord: coord-server
	install -m 755 -o root -g root coord-server ${pkgdir}/usr/bin/qrystal-coord-server
	#
	mkdir -p ${pkgdir}/etc/qrystal-coord/
	install ${src}/config/coord-server-config.json ${pkgdir}/etc/qrystal-coord/config.json
	chown root:root -R ${pkgdir}/etc/qrystal-coord/
	chmod ugo+rX -R ${pkgdir}/etc/qrystal-coord/
	#
	mkdir -p ${pkgdir}/usr/lib/systemd/system
	install ${src}/config/coord-server.service ${pkgdir}/usr/lib/systemd/system/qrystal-coord-server.service
	systemctl daemon-reload

uninstall-coord:
	rm ${pkgdir}/usr/bin/qrystal-coord-server
	rm ${pkgdir}/etc/qrystal-coord
	rm ${pkgdir}/usr/lib/systemd/system/qrystal-coord-server.service

install-device: device-client
	install -m 755 -o root -g root device-client ${pkgdir}/usr/bin/qrystal-device-client
	mkdir -p ${pkgdir}/usr/lib/sysusers.d
	install -m 644 ${src}/config/sysusers-device.conf ${pkgdir}/usr/lib/sysusers.d/qrystal-device.conf
	systemctl restart systemd-sysusers
	#
	mkdir -p ${pkgdir}/etc/qrystal-device/
	chown root:qrystal-device ${pkgdir}/etc/qrystal-device/
	chmod 755 ${pkgdir}/etc/qrystal-device/
	install -m 640 -o root -g qrystal-device ${src}/config/device-client-config.json ${pkgdir}/etc/qrystal-device/client-config.json
	install -m 640 -o root -g qrystal-device ${src}/config/device-dns-config.json ${pkgdir}/etc/qrystal-device/dns-config.json
	#
	mkdir -p ${pkgdir}/usr/lib/systemd/system
	install ${src}/config/device-client.service ${pkgdir}/usr/lib/systemd/system/qrystal-device-client.service
	systemctl daemon-reload

uninstall-device:
	rm ${pkgdir}/usr/bin/qrystal-device-client
	rm ${pkgdir}/usr/bin/qrystal-device-dns
	rm ${pkgdir}/usr/lib/sysusers.d/qrystal-device.conf
	rm ${pkgdir}/usr/lib/systemd/system/qrystal-device-client.service
	rm ${pkgdir}/usr/lib/systemd/system/qrystal-device-dns.service
	rm ${pkgdir}/usr/lib/systemd/system/qrystal-device-dns.socket

.PHONY: clean install-coord uninstall-coord install-device uninstall-device
