package main

import (
	"log"
	"os"
	"sync"

	"github.com/NCAR/HPCFL_TerraformScripts/scripts/slurm"
	"github.com/NCAR/HPCFL_TerraformScripts/scripts/terraform"
	"github.com/NCAR/HPCFL_TerraformScripts/scripts/utils"
)

func tfAdd(names []string) {
	for _, inst := range names {
		log.Printf("INFO: Creating Instance %s\n", inst)
		terraform.Add(inst)
	}
}

func setup(names []string) {
	var wg sync.WaitGroup
	//setup instances concurrently
	for _, n := range names {
		wg.Add(1)
		go func(name string, wg *sync.WaitGroup) {
			defer wg.Done()
			inst := terraform.Info(name)
			log.Printf("INFO: Provisioning new instance %s\n", inst)
			err := inst.Setup()
			if err != nil {
				log.Printf("ERROR:addEC2:%v\n", err)
			}
			log.Printf("INFO: Done provisioning %s\n", inst)
		}(n, &wg)
	}
	wg.Wait()
}

func main() {
	//set up logging file
	cleanup, err := utils.SetupLogging("addEC2")
	if err != nil {
		log.Fatalf("CRITICAL: %s\n", err)
	}
	defer cleanup()

	//check called with an arg
	err = utils.CheckNumArgs(2, os.Args, "<instance name glob>")
	if err != nil {
		log.Fatalf("CRITICAL: %s\n", err)
	}

	log.Printf("INFO: %s called with args %s\n", os.Args[0], os.Args[1])

	//get new instance names
	newInsts := slurm.NodeNames(os.Args[1])

	//create new instances
	tfAdd(newInsts)

	//update infrastructure to reflect config updates
	log.Printf("INFO: Pushing infrastructure update\n")
	terraform.Update()

	//do basic setup of new instances in parallel
	setup(newInsts)

	log.Printf("INFO: Done adding instances\n")
}
