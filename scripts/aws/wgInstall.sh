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

function remote_run {
    #TODO don't ignore hostkey every time, only the first
    ssh -i$SSHKEY "ec2-user@"$IP "$1"
}

function initial_setup {
	ssh -i$SSHKEY -oStrictHostKeyChecking=no "ec2-user@"$IP "sudo yum update -y"
	remote_run "sudo amazon-linux-extras install epel -y"
	remote_run "sudo curl -Lo /etc/yum.repos.d/wireguard.repo https://copr.fedorainfracloud.org/coprs/jdoss/wireguard/repo/epel-7/jdoss-wireguard-epel-7.repo"
	remote_run "sudo yum install wireguard-dkms wireguard-tools -y"
	remote_run "sudo hostnamectl set-hostname $HOSTNAME"
	remote_run "sudo reboot"

	echo "rebooting"
	#wait a bit for reboot
    sleep 10
}

function wg_setup {
tries=0
ret=255
while [ $ret -ne 0 -a $tries -lt 20 ]; do
  ssh -i$SSHKEY "ec2-user@"$IP "echo connected"
  ret=$?
  sleep 20
  tries=$((tries+=1))
  #echo $tries
done
	remote_run "sudo dkms autoinstall"
	remote_run "sudo sysctl -w net.ipv4.ip_forward=1"
	remote_run "ls /lib/modules/4.14.146-120.181.amzn2.x86_64/kernel/net/"
	remote_run "sudo modprobe ipv6"
	remote_run "sudo modprobe udp_tunnel"
	remote_run "sudo modprobe ip6_udp_tunnel"
	echo "attempt 1 at modprobe wireguard"
	remote_run "sudo lsmod"
	echo "depmod -a :"
	remote_run "sudo depmod -a" 
	echo "lsmod:"
	remote_run "sudo lsmod"
	echo "modprobe"
	remote_run "sudo modprobe wireguard"
	echo "lsmod:" 
	remote_run "sudo lsmod"
	set -e
	remote_run "sudo modprobe wireguard"
	echo "it worked"
	# create remote keys
	remote_run "umask 077; wg genkey > private"
	remote_run "wg pubkey < private > public"
	echo "Made remote keys"
	# setup interfaces
	sudo ip link add wg0 type wireguard
	remote_run "sudo ip link add wg0 type wireguard"
	sudo ip addr add 10.0.0.1/24 dev wg0
	remote_run "sudo ip addr add 10.0.0.2/24 dev wg0"
	sudo wg set wg0 private-key /home/slurm/terraform/wg/private
	remote_run "sudo wg set wg0 private-key ./private"
	sudo ip link set wg0 up
	remote_run "sudo ip link set wg0 up"
	echo "wg interfaces created"
	# keepalive for NAT traversal
#	sudo wg set wg0 persistent-keepalive 10
	sudo wg set wg0 listen-port 51820
	remote_run "sudo wg set wg0 listen-port 33434"

	# figure out what remote and local pub keys are
	REMOTE_KEY=$(remote_run "sudo wg | grep -A1 wg0 | grep 'public key' | cut -d ':' -f 2 | cut -c 2-")
	LOCAL_KEY=$(sudo wg | grep -A1 wg0 | grep 'public key' | cut -d ':' -f 2 | cut -c 2-)

	# make wireguard connection
	remote_run "sudo wg set wg0 peer $LOCAL_KEY allowed-ips 10.0.0.1/32,192.168.0.0/24"
	sudo wg set wg0 peer $REMOTE_KEY allowed-ips 10.0.0.2/32,192.168.2.0/24 endpoint $IP:33434 #TODO

	sudo ip r add 192.168.2.0/24 dev wg0 via 10.0.0.2
	remote_run "sudo ip r add 192.168.0.0/24 dev wg0 via 10.0.0.1"
	echo "wg tunnel created"
}


#TODO check user passes ip and sshkey filepath
HOSTNAME=$1
IP=$2 #public IP of node
SSHKEY="/home/slurm/.ssh/hpcfl2"
#set -e

sudo ip link del wg0
initial_setup 
wg_setup
remote_run "sudo iptables -t nat -A POSTROUTING -o eth0 ! -d 192.168.2.0/24 -j MASQUERADE"
echo "$1 set up"
