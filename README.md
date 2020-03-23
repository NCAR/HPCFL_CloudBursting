# Cloud Bursting
The goal of this repo is to contain all required configuration, code, and instructions, to dynamically spin up/down cloud instances when needed via the hooks provided by a batch scheduler (Slurm, PBS, etc) on multiple different clouds (AWS, Azure, GCP).

## Other resources
Helpful documents related to this repo and its goal can be found in the docs dir

## Building scripts
```bash
cd scripts
# can't build in place because
go build ../addEC2 addEC2/addEC2.go
go build ../rmEC2 rmEC2/rmEC2.go
```

## Setup

### slurm
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

### salt
- Used for provisioning instances
- Need to update ip of salt master in setup scripts (TODO make this easy)
- Can find salt setup that goes with this repo here (TODO add salt repo link)

### aws
- ```sudo -u slurm terraform apply -auto-approve tfFiles/```
  - sets up router instance
  - can be found in tfFiles/infra.tf
- ```sudo wgInstall router0 <public ip from last command>```
  - sets up router0 instance as a wireguard router
- need to put aws credentials in ~slurm/.aws/credentials
- need to put aws config in ~slurm/.aws/config
  - might actually be optional (TODO check)
  - ex: ```
[default]
region = us-east-2
output = json
```

### other clouds
- other clouds are easily added one simply needs to:
  1. Create a package for that cloud that implements a struct that fills the utils.Instance interface
  2. Update terraform.{Info,Add} function switch statments to support new cloud
  3. Update salt to do the rest of the provisioning neccesarry for the new instance types
  4. Update slurm.conf so slurm knows about the new instances
