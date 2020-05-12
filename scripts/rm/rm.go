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

	"github.com/NCAR/HPCFL_TerraformScripts/scripts/slurm"
	"github.com/NCAR/HPCFL_TerraformScripts/scripts/terraform"
	"github.com/NCAR/HPCFL_TerraformScripts/scripts/utils"
)

var config = "/opt/slurm/latest/etc/cloud_config.json"

func rm(names string) {
	for _, inst := range slurm.NodeNames(names) {
		log.Printf("INFO: Removing %s instance\n", inst)
		terraform.Del(inst)
	}
}

func main() {
	utils.ParseConfig(config)
	//setup logging
	cleanup, err := utils.SetupLogging(utils.Config("log.rm").Self())
	if err != nil {
		log.Fatalf("Critical: %s\n", err)
	}
	defer cleanup()

	//check was given an arg
	err = utils.CheckNumArgs(2, os.Args, "<instance name glob>")
	if err != nil {
		log.Fatalf("CRITICAL: %s\n", err)
	}

	// clean up each instance to be removed
	rm(os.Args[1])

	//check in terraform is in sync with scheduler
	//make a set of all on instances in slurm
	found := make(map[string]struct{})
	for _, inst := range slurm.On() {
	found[inst] = struct{}{}
	}
	//check against on instances in terraform
	for _, inst := range terraform.On() {
	if _, ok := found[inst]; ok {
	    //this is good, remove from set as inst is in a good state
	    delete(found, inst)
	}else{
	    //out of sync for this instance
	    //scheduler doesn't know about it so just kill it
	    terraform.Del(inst)
	}
	}
	// see what instances are in slurm and not terraform
	// anything left in set was not in terraform
	// scheduler wants these instances, but they are in a weird state
	// so probably safest to just kill them
	for inst, _ := range found {
	terraform.Del(inst)
	}

	// update cloud infrastructure
	log.Printf("INFO: Pushing Updates\n")
	terraform.Update()
	log.Printf("INFO: Done Updating\n")
}
