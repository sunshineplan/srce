#! /bin/bash

installSoftware() {
    apt -qq -y install nginx
    apt -qq -y -t $(lsb_release -sc)-backports install golang-go
}

installSRCE() {
    curl -Lo- https://github.com/sunshineplan/srce-go/archive/v1.0.tar.gz | tar zxC /var/www
    mv /var/www/srce-go* /var/www/srce-go
    cd /var/www/srce-go
    go build
}

configSRCE() {
    read -p 'Please enter metadata server: ' server
    read -p 'Please enter VerifyHeader header: ' header
    read -p 'Please enter VerifyHeader value: ' value
    read -p 'Please enter unix socket(default: /run/srce-go.sock): ' unix
    [ -z $unix ] && unix=/run/srce-go.sock
    read -p 'Please enter host(default: 127.0.0.1): ' host
    [ -z $host ] && host=127.0.0.1
    read -p 'Please enter port(default: 12345): ' port
    [ -z $port ] && port=12345
    read -p 'Please enter log path(default: /var/log/app/srce-go.log): ' log
    [ -z $log ] && log=/var/log/app/srce-go.log
    mkdir -p $(dirname $log)
    sed "s,\$server,$server," /var/www/srce-go/config.ini.default > /var/www/srce-go/config.ini
    sed -i "s/\$header/$header/" /var/www/srce-go/config.ini
    sed -i "s/\$value/$value/" /var/www/srce-go/config.ini
    sed -i "s,\$unix,$unix," /var/www/srce-go/config.ini
    sed -i "s,\$log,$log," /var/www/srce-go/config.ini
    sed -i "s/\$host/$host/" /var/www/srce-go/config.ini
    sed -i "s/\$port/$port/" /var/www/srce-go/config.ini
}

setupsystemd() {
    cp -s /var/www/srce-go/scripts/srce-go.service /etc/systemd/system
    systemctl enable srce-go
    service srce-go start
}

writeLogrotateScrip() {
    if [ ! -f '/etc/logrotate.d/app' ]; then
	cat >/etc/logrotate.d/app <<-EOF
		/var/log/app/*.log {
		    copytruncate
		    rotate 12
		    compress
		    delaycompress
		    missingok
		    notifempty
		}
		EOF
    fi
}

setupNGINX() {
    cp -s /var/www/srce-go/scripts/srce-go.conf /etc/nginx/conf.d
    sed -i "s/\$domain/$domain/" /var/www/srce-go/scripts/srce-go.conf
    sed -i "s,\$unix,$unix," /var/www/srce-go/scripts/srce-go.conf
    service nginx reload
}

main() {
    read -p 'Please enter domain:' domain
    installSoftware
    installSRCE
    configSRCE
    setupsystemd
    writeLogrotateScrip
    setupNGINX
}

main