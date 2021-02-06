#! /bin/bash

installSoftware() {
    apt -qq -y install nginx
}

installSRCE() {
    mkdir -p /var/www/srce
    curl -Lo- https://github.com/sunshineplan/srce/releases/download/v1.0/release.tar.gz | tar zxC /var/www/srce
    cd /var/www/srce
    chmod +x srce
}

configSRCE() {
    read -p 'Please enter metadata server: ' server
    read -p 'Please enter VerifyHeader header: ' header
    read -p 'Please enter VerifyHeader value: ' value
    read -p 'Please enter unix socket(default: /run/srce.sock): ' unix
    [ -z $unix ] && unix=/run/srce.sock
    read -p 'Please enter host(default: 127.0.0.1): ' host
    [ -z $host ] && host=127.0.0.1
    read -p 'Please enter port(default: 12345): ' port
    [ -z $port ] && port=12345
    read -p 'Please enter log path(default: /var/log/app/srce.log): ' log
    [ -z $log ] && log=/var/log/app/srce.log
    read -p 'Please enter update URL: ' update
    read -p 'Please enter exclude files: ' exclude
    mkdir -p $(dirname $log)
    sed "s,\$server,$server," /var/www/srce/config.ini.default > /var/www/srce/config.ini
    sed -i "s/\$header/$header/" /var/www/srce/config.ini
    sed -i "s/\$value/$value/" /var/www/srce/config.ini
    sed -i "s,\$unix,$unix," /var/www/srce/config.ini
    sed -i "s,\$log,$log," /var/www/srce/config.ini
    sed -i "s/\$host/$host/" /var/www/srce/config.ini
    sed -i "s/\$port/$port/" /var/www/srce/config.ini
    sed -i "s,\$update,$update," /var/www/srce/config.ini
    sed -i "s|\$exclude|$exclude|" /var/www/srce/config.ini
    ./srce install || exit 1
    service srce start
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
    cp -s /var/www/srce/scripts/srce.conf /etc/nginx/conf.d
    sed -i "s/\$domain/$domain/" /var/www/srce/scripts/srce.conf
    sed -i "s,\$unix,$unix," /var/www/srce/scripts/srce.conf
    service nginx reload
}

main() {
    read -p 'Please enter domain:' domain
    installSoftware
    installSRCE
    configSRCE
    writeLogrotateScrip
    setupNGINX
}

main