#!/bin/bash
sleep 10
tries=0
scp -oUserKnownHostsFile=/dev/null -oStrictHostKeyChecking=no "$2saltInstall.sh" $1:~/
ret=$?
while [ $ret -ne 0 -a $tries -lt 20 ]; do
  sleep 20
  scp -oUserKnownHostsFile=/dev/null -oStrictHostKeyChecking=no "$2saltInstall.sh" $1:~/
  ret=$?
  tries=$((tries+=1))
  #echo $tries
done

if [ $ret -ne 0 ]; then
  exit 7
fi

scp -oUserKnownHostsFile=/dev/null -oStrictHostKeyChecking=no "$2salt-minion.service" $1:~/
scp -oUserKnownHostsFile=/dev/null -oStrictHostKeyChecking=no "$2ifcfg-eth0" $1:~/
scp -oUserKnownHostsFile=/dev/null -oStrictHostKeyChecking=no "$2minion" $1:~/
scp -oUserKnownHostsFile=/dev/null -oStrictHostKeyChecking=no "$2keys/$1.pem" $1:~/minion.pem
scp -oUserKnownHostsFile=/dev/null -oStrictHostKeyChecking=no "$2keys/$1.pub" $1:~/minion.pub
ssh -oUserKnownHostsFile=/dev/null -oStrictHostKeyChecking=no $1 "./saltInstall.sh $1"

