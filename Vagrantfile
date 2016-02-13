# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure(2) do |config|
  config.vm.box = "ubuntu/trusty64"

  config.vm.provision "shell", inline: <<-SHELL
    curl -sf -o /tmp/go1.5.3.linux-amd64.tar.gz -L https://storage.googleapis.com/golang/go1.5.3.linux-amd64.tar.gz
    sudo mkdir -p /opt && cd /opt && sudo tar xfz /tmp/go1.5.3.linux-amd64.tar.gz && rm -f /tmp/go1.5.3.linux-amd64.tar.gz
    curl -s https://packagecloud.io/install/repositories/darron/consul/script.deb.sh | sudo bash
    sudo apt-get install -y consul git make graphviz dnsmasq
    sudo mkdir -p /var/log/dnsmasq /etc/consul.d /var/lib/consul /var/log/consul
    sudo ln -s /lib/init/upstart-job /etc/init.d/consul
    curl -s https://raw.githubusercontent.com/DataDog/kvexpress-cookbook/master/files/default/consul.conf > /tmp/consul.conf && chmod 644 /tmp/consul.conf && sudo chown root.root /tmp/consul.conf && sudo mv -f /tmp/consul.conf /etc/init/consul.conf
    sudo cat > /etc/consul.d/default.json << EOF
{
  "client_addr": "127.0.0.1",
  "data_dir": "/var/lib/consul",
  "server": true,
  "bootstrap": true,
  "recursor": "8.8.8.8",
  "bind_addr": "0.0.0.0",
  "log_level": "debug",
  "node_name": "goshe-consul"
}
EOF
    sudo cat > /etc/consul.d/kafka.json << EOF
{
  "service": {
    "name": "kafka",
    "check": {
      "interval": "60s",
      "script": "/bin/true"
    }
  }
}
EOF
    sudo cat > /etc/consul.d/casandra.json << EOF
{
  "service": {
    "name": "cassandra",
    "check": {
      "interval": "60s",
      "script": "/bin/true"
    }
  }
}
EOF
    sudo cat > /etc/hosts.consul << EOF
127.0.0.1 goshe.service.consul
127.0.0.1 vagrant.service.consul
127.0.0.1 datadog.service.consul
EOF
    sudo cat > /etc/dnsmasq.d/10-consul << EOF
server=/consul/127.0.0.1#8600
EOF
    sudo cat > /etc/default/dnsmasq << EOF
DNSMASQ_OPTS="--addn-hosts=/etc/hosts.consul --log-facility=/var/log/dnsmasq/dnsmasq --local-ttl=10"
ENABLED=1
CONFIG_DIR=/etc/dnsmasq.d,.dpkg-dist,.dpkg-old,.dpkg-new
EOF
    sudo service dnsmasq restart
    sudo service consul start
    sudo cat > /etc/profile.d/go.sh << EOF
export GOROOT="/opt/go"
export GOPATH="/home/vagrant/gocode"
export PATH="/opt/go/bin://home/vagrant/gocode/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
export GOSHE_DEBUG=1
EOF
    cd /vagrant && source /etc/profile.d/go.sh
  SHELL
end
