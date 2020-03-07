#!/bin/bash

sudo hostnamectl set-hostname $1

sudo ip r del default &
sleep 2
sudo ip r add default via 192.168.2.10

printf "\n192.168.0.120 salt\n" | sudo sh -c "cat >> /etc/hosts"
sudo yum install https://repo.saltstack.com/py3/amazon/salt-py3-amzn2-repo-latest.amzn2.noarch.rpm -y
sudo amazon-linux-extras install epel -y
sudo yum clean expire-cache
sudo yum update -y
sudo yum install salt-minion --disablerepo=epel -y
sudo mv -f /home/ec2-user/salt-minion.service /usr/lib/systemd/system/salt-minion.service
sudo mv -f /home/ec2-user/ifcfg-eth0 /etc/sysconfig/network-scripts/ifcfg-eth0
sudo mv -f /home/ec2-user/minion /etc/salt/minion
sudo systemctl daemon-reload
sudo systemctl enable salt-minion
sudo reboot
