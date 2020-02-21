package main

import (
	"github.com/NCAR/HPCFL_TerraformScripts/scripts/slurm"
	"github.com/NCAR/HPCFL_TerraformScripts/scripts/terraform"
	"log"
	"os"
	"sync"
)

func main() {
	//set up logging file
	//TODO remove hard coded filepath
	f, err := os.OpenFile("/var/lib/slurm/log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("ERROR: Could not open log file /var/lib/slurm/log, %s\n", err)
	}
	defer f.Close()
	log.SetOutput(f)

	//check called with an arg
	if len(os.Args) != 2 {
		log.Fatalf("ERROR: Usage: $s <instance name glob>\n", os.Args[0])
	}
	log.Printf("INFO: %s called with args %s\n", os.Args[0], os.Args[1])

	//create config files for new instances
	newInsts := slurm.NodeNames(os.Args[1])
	for _, inst := range newInsts {
		log.Printf("INFO: Creating ec2 Instance %s\n", inst)
		terraform.Add(inst)
	}

	//update infrastructure to reflect config updates
	log.Printf("INFO: Pushing infrastructure update\n")
	terraform.Update()

	//do basic setup of new instances in parallel
	var wg sync.WaitGroup
	for _, i := range newInsts {
		wg.Add(1)
		go func(name string, wg *sync.WaitGroup) {
			defer wg.Done()
			inst := terraform.Info(name)
			log.Printf("INFO: Provisioning new instance %s\n", inst)
			inst.Setup()
			log.Printf("INFO: Done provisioning %s\n", inst)
		}(i, &wg)
	}
	wg.Wait()

	log.Printf("INFO: Done adding instances\n")
}
