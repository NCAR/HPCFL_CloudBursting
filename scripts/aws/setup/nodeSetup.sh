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

SSHKEY='/home/slurm/.ssh/hpcfl2'

# make sure old key isn't in know hosts
sed -i '/'$1'/d' ~/.ssh/known_hosts

sleep 10
tries=0
ret=255
while [ $ret -ne 0 -a $tries -lt 20 ]; do
  sleep 20
  scp -oStrictHostKeyChecking=no -i$SSHKEY "$2saltInstall.sh" centos@$1:~/
  ret=$?
  tries=$((tries+=1))
  #echo $tries
done

if [ $ret -ne 0 ]; then
  exit 7
fi

scp -i$SSHKEY "$2salt-minion.service" centos@$1:~/
scp -i$SSHKEY "$2ifcfg-eth0" centos@$1:~/
scp -i$SSHKEY "$2minion" centos@$1:~/
scp -i$SSHKEY "$2keys/$1.pem" centos@$1:~/minion.pem
scp -i$SSHKEY "$2keys/$1.pub" centos@$1:~/minion.pub
ssh -i$SSHKEY centos@$1 "chmod +x saltInstall.sh"
ssh -i$SSHKEY centos@$1 "./saltInstall.sh $1"
