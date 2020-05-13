#!/bin/bash

#Copyright (c) 2020, University Corporation for Atmospheric Research
#All rights reserved.
#
#Redistribution and use in source and binary forms, with or without 
#modification, are permitted provided that the following conditions are met:
#
#1. Redistributions of source code must retain the above copyright notice, 
#this list of conditions and the following disclaimer.
#
#2. Redistributions in binary form must reproduce the above copyright notice,
#this list of conditions and the following disclaimer in the documentation
#and/or other materials provided with the distribution.
#
#3. Neither the name of the copyright holder nor the names of its contributors
#may be used to endorse or promote products derived from this software without
#specific prior written permission.
#
#THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" 
#AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
#IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE 
#ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
#LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR 
#CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF 
#SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS 
#INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, 
#WHETHER IN CONTRACT, STRICT LIABILITY,
#OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
#OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

sudo hostnamectl set-hostname $1

ROUTER=$2
SALT=$3

sudo ip r del default &
sleep 2
sudo ip r add default via $ROUTER

sed -i "s/ROUTER/$ROUTER/g" /home/centos/ifcfg-eth0
printf "\n$SALT salt\n" | sudo sh -c "cat >> /etc/hosts"
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
