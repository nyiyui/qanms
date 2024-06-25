src = .
flags = -race -tags sdnotify

coord-server:
	go build ${flags} ${src}/cmd/coord-server

device-client:
	go build ${flags} ${src}/cmd/device-client

device-dns:
	go build ${flags} ${src}/cmd/device-dns

gen-keys:
	go build ${flags} ${src}/cmd/gen-keys

install-device: device-client device-dns
	install -m 755 -o root -g root device-client ${pkgdir}/usr/bin/qrystal-device-client
	install -m 755 -o root -g root device-dns ${pkgdir}/usr/bin/qrystal-device-dns
	mkdir -p ${pkgdir}/usr/lib/sysusers.d
	install -m 644 ${src}/config/sysusers-device.conf ${pkgdir}/usr/lib/sysusers.d/qrystal-device.conf
	systemctl restart systemd-sysusers
	#
	mkdir -p ${pkgdir}/etc/qrystal-device/
	chown root:qrystal-node ${pkgdir}/etc/qrystal-device/
	chmod 755 ${pkgdir}/etc/qrystal-device/
	install -m 640 -o root -g qrystal-device ${src}/config/device-client-config.json ${pkgdir}/etc/qrystal-device/client-config.json
	install -m 640 -o root -g qrystal-device ${src}/config/device-dns-config.json ${pkgdir}/etc/qrystal-device/dns-config.json
	#
	mkdir -p ${pkgdir}/usr/lib/systemd/system
	install ${src}/config/device-client.service ${pkgdir}/usr/lib/systemd/system/qrystal-device-client.service
	install ${src}/config/device-dns.service ${pkgdir}/usr/lib/systemd/system/qrystal-device-dns.service
	install ${src}/config/device-dns.socket ${pkgdir}/usr/lib/systemd/system/qrystal-device-dns.socket
	systemd daemon-reload

uninstall-device:
	rm ${pkgdir}/usr/bin/qrystal-device-client
	rm ${pkgdir}/usr/bin/qrystal-device-dns
	rm ${pkgdir}/usr/lib/sysusers.d/qrystal-device.conf
	rm ${pkgdir}/usr/lib/systemd/system/qrystal-device-client.service
	rm ${pkgdir}/usr/lib/systemd/system/qrystal-device-dns.service
	rm ${pkgdir}/usr/lib/systemd/system/qrystal-device-dns.socket
