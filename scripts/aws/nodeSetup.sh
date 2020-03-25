#!/bin/bash
sleep 10
tries=0
scp -oUserKnownHostsFile=/dev/null -oStrictHostKeyChecking=no -i/home/slurm/.ssh/hpcfl2 "$2saltInstall.sh" ec2-user@$1:~/
ret=$?
while [ $ret -ne 0 -a $tries -lt 20 ]; do
  sleep 20
  scp -oUserKnownHostsFile=/dev/null -oStrictHostKeyChecking=no -i/home/slurm/.ssh/hpcfl2 "$2saltInstall.sh" ec2-user@$1:~/
  ret=$?
  tries=$((tries+=1))
  #echo $tries
done

if [ $ret -ne 0 ]; then
  exit 7
fi

scp -oUserKnownHostsFile=/dev/null -oStrictHostKeyChecking=no -i/home/slurm/.ssh/hpcfl2 "$2salt-minion.service" ec2-user@$1:~/
scp -oUserKnownHostsFile=/dev/null -oStrictHostKeyChecking=no -i/home/slurm/.ssh/hpcfl2 "$2ifcfg-eth0" ec2-user@$1:~/
scp -oUserKnownHostsFile=/dev/null -oStrictHostKeyChecking=no -i/home/slurm/.ssh/hpcfl2 "$2minion" ec2-user@$1:~/
scp -oUserKnownHostsFile=/dev/null -oStrictHostKeyChecking=no -i/home/slurm/.ssh/hpcfl2 "$2keys/$1.pem" ec2-user@$1:~/minion.pem
scp -oUserKnownHostsFile=/dev/null -oStrictHostKeyChecking=no -i/home/slurm/.ssh/hpcfl2 "$2keys/$1.pub" ec2-user@$1:~/minion.pub
ssh -oUserKnownHostsFile=/dev/null -oStrictHostKeyChecking=no -i/home/slurm/.ssh/hpcfl2 ec2-user@$1 "./saltInstall.sh $1"
