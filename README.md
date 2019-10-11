# Cloud Bursting
The goal of this repo is to contain all required configuration, code, and instructions, to dynamically spin up/down cloud instances when needed via a batch scheduler {slurm,pbs} on multiple different clouds {aws,azure}

## Usage
1. cd scripts; go build -o ../addEC2 addEC2/addEC2.go
2. ./scripts/shellScripts/wgInstall.sh <router0 public ip> <ssh key path> router0
3. cd scripts/shellScripts && ./nodeSetup aws0
4. sudo salt-key --accept=aws0

## Requirments
- golang
- cloud account with {aws,azure}
- terraform cli
- {slurm,pbs}

## Notes
- scripts must be run from terraform dir
  - working on fixing this (because of terraform only likely fix is 'cd terraform\_dir; ./terraform')

## Setup
- need to get aws credentials if using aws (See aws website)
  - easiest to use aws cli
- need an ssh key pair
  - create and update the infra.tf file in tfFiles

## Current Status
- able to bring up all aws infra up needed to run ec2 instances
- Automate generation of a VPN between local network and aws VPC

## In Progress
- Automate ec2 instance provisioning

## TODO
- Connect it all to Slurm
- Make VPC connection to cloud redundant
- Find efficient way to move data between local network and aws
- Test scallability
- Use azure
- Use pbs

