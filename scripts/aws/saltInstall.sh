#!/bin/bash

sudo hostnamectl set-hostname $1

sudo ip r del default &
sleep 2
sudo ip r add default via 192.168.2.10

printf "\n192.168.0.120 salt\n" | sudo sh -c "cat >> /etc/hosts"
sudo yum install https://repo.saltstack.com/yum/redhat/salt-repo-latest.el7.noarch.rpm -y
#sudo yum install https://repo.saltstack.com/py3/amazon/salt-py3-amzn2-repo-latest.amzn2.noarch.rpm -y
sudo yum install epel-release -y
#sudo amazon-linux-extras install epel -y
sudo yum clean expire-cache
sudo yum update -y
sudo yum install salt-minion --disablerepo=epel -y
sudo mv -f /usr/lib/systemd/system/salt-minion.service /home/centos/oldsalt-minion.service
sudo mv -f /home/centos/salt-minion.service /usr/lib/systemd/system/salt-minion.service
sudo mv -f /home/centos/ifcfg-eth0 /etc/sysconfig/network-scripts/ifcfg-eth0
sudo mv -f /home/centos/minion.pem /etc/salt/pki/minion/minion.pem
sudo chown root:root /etc/salt/pki/minion/minion.pem
sudo chmod 400 /etc/salt/pki/minion/minion.pem
sudo mv -f /home/centos/minion.pub /etc/salt/pki/minion/minion.pub
sudo chown root:root /etc/salt/pki/minion/minion.pem
sudo chmod 644 /etc/salt/pki/minion/minion.pub
sudo mv -f /home/centos/minion /etc/salt/minion
sudo systemctl daemon-reload
sudo systemctl enable salt-minion
sudo reboot
