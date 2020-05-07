# Cloud Bursting
The goal of this repo is to contain all required configuration, code, and instructions, to dynamically spin up/down cloud instances when needed via the hooks provided by a batch scheduler (Slurm, PBS, etc) on multiple different clouds (AWS, Azure, GCP).

## Other resources
Helpful documents related to this repo and its goal can be found in the docs dir

## Building Scripts
```bash
cd HPCFL_TerraformScripts/scripts
# can't build in place because script names are the same as existing dirs

go build -ldflags="-X main.config=</path/to/config/file.json>" -o ../add add/add.go
go build -ldflags="-X main.config=</path/to/config/file.json>" -o ../rm rm/rm.go
```
- Instead of using the ```ldflags``` flag one could also change the ```config``` variable in add/add.go and rm/rm.go 
- The config file location defaults to /opt/slurm/latest/etc/cloud_config.json
- ```config.json``` is an example config file

## Setup

### Wireguard
- Just need it installed locally

### Slurm
- slurm.conf need following options
```
SuspendProgram=/path/to/rmInstance/Program
ResumeProgram=/path/to/addInstance/Program
SuspendTime=<seconds to leave an instance idle before shutting it down>
ResumeTimeout=<seconds an instance can take to start up>
TreeWidth=<int greater than number of cloud nodes, max value is 65533>
NodeName=<name glob> Weight=<uint> Feature=cloud State=Cloud
PrivateData=cloud # technically optional, but required to make sinfo etc. output usefull
```

### Terraform
- Need it, can't remember if there was any weird setup
- (TODO add link to website)

### Salt
- Example salt setup can be found [here](https://github.com/NCAR/HPCFL_CloudBurstingSalt)
- Used for provisioning instances
- Need to update ip of salt master in setup scripts (TODO make this easy)
- Can find salt setup that goes with this repo here (TODO add salt repo link)
- Salt keys should be pregenerated for minions and put in scripts/aws/keys and preaccepted by the salt master
  - Can generate keys via ```salt-key --gen-keys=<name>``` command
  - make sure keys are readable by slurm user
  - to accept keys copy $NAME.pub to /etc/salt/pki/master/minions/$NAME (remove .pub)

### AWS
- ```sudo -u slurm terraform apply -auto-approve tfFiles/```
  - Sets up router instance
  - Can be found in tfFiles/infra.tf
- ```sudo wgInstall router0 <public ip from last command>```
  - Sets up router0 instance as a wireguard router
- Need to put aws credentials in ~slurm/.aws/credentials
- Need to put aws config in ~slurm/.aws/config
  - Might actually be optional (TODO check)
  - ex:
```
[default]
region = us-east-2
output = json
```

## Starting
### AWS
- after completing all of the setup steps to start all required infrastructure to use aws nodes run:
  - `./terraform apply -auto-approve tfFiles`
  - after a while (~1min) this will finish with the last bit of output looking similar to:
```
router0 = {
  "nodeName" = "router0"
  "privateIP" = "192.168.2.10"
  "publicIP" = "3.16.135.102"
}
```
  - terraform has configured and started an ec2 instance called `router0` with a publically addressable ip `3.16.135.102` (this will change each time the instance is started, and an ip of `192.168.2.10` on the a private subnet (this is static)
  - to connect this instance to your on premise network over a vpn tunnel run `./scripts/aws/wgInstall.sh router0 3.16.135.102` as a user with sudo priveleges
    - this can take a few minutes to finish
    - this provisions `router0`, then sets up a wireguared vpn connection between the local machine and `router0`   
  - now you are ready to run jobs!

### Other Clouds
- Other clouds are easily added one simply needs to:
  1. Create a package for that cloud that implements a struct that fills the utils.Instance interface
  2. Update terraform.{Info,Add} function switch statments to support new cloud
  3. Update salt to do the rest of the provisioning neccesarry for the new instance types
  4. Update slurm.conf so slurm knows about the new instances

## Notes
### Slurm
The output of sinfo and scontrol can be quite criptic for cloud nodes. Here are some notes on what some of the characters mean when contained in a nodes `state`
- `~` in power saving mode
- `#` powering up / being configured
- `$` in a maintenance reservation
- `@` scheduled to reboot
- `\*` unreachable
- `%` powering off

## License
```
Copyright (c) 2020, University Corporation for Atmospheric Research
All rights reserved.

Redistribution and use in source and binary forms, with or without 
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, 
this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation
and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors
may be used to endorse or promote products derived from this software without
specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" 
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE 
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR 
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF 
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS 
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, 
WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
```
