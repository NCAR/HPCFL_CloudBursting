/*
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
*/

package main

import (
	"log"
	"os"
	"sync"

	"github.com/NCAR/HPCFL_TerraformScripts/scripts/slurm"
	"github.com/NCAR/HPCFL_TerraformScripts/scripts/terraform"
	"github.com/NCAR/HPCFL_TerraformScripts/scripts/utils"
)

var config = "/opt/slurm/latest/etc/cloud_config.json"

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
			log.Printf("%+v\n", inst)
			if inst == nil {
				log.Printf("CRITICAL: Could not find instance %s\n", n)
				return
			}
			log.Printf("INFO: Provisioning new instance %s\n", inst)
			err := inst.Setup()
			if err != nil {
				log.Printf("ERROR:add:%v\n", err)
				return
			}
			log.Printf("INFO: Done provisioning %s\n", inst)
		}(n, &wg)
	}
	wg.Wait()
}

func main() {
	err := utils.ParseConfig(config)
	if err != nil {
		log.Fatalf("CRITICAL: Can't parse config: %s\n", err)
	}
	//set up logging file
	cleanup, err := utils.SetupLogging(utils.Config("log.add").Self())
	if err != nil {
		log.Fatalf("CRITICAL: Can't open log file: %s\n", err)
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
