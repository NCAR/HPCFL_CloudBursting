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

package slurm

import (
	"log"
	"os/exec"
	"regexp"
	"strings"

	"github.com/NCAR/HPCFL_TerraformScripts/scripts/utils"
)

//NodeNames takes a glob string and returns a slice of all of the names contained in the glob
//The glob string is of the format slurm uses
func NodeNames(glob string) []string {
	log.Printf("DEBUG:slurm: expanding glob %s\n", glob)
	//use slurm to expand glob
	out, err := exec.Command(utils.Config("slurm.dir").Self()+"/bin/scontrol", "show", "hostname", glob).Output()
	if err != nil {
		log.Printf("ERROR:slurm: Error expanding nodename glob %s\n", err)
		return nil
	}
	names := strings.Split(string(out), "\n")
	//last element is empty/just whitespace so drop it
	return names[:len(names)-1]
}

//On returns a slice of the currently on, or starting up cloud nodes
//It also resets the state on any "broken" nodes
func On() []string {
	on, _, broken := status()
	fixBroken(broken)
	return on
}

//status returns on, off, broken list of node names
//returns nil if no nodes in that state
//nodes that are powering up are considered on
//nodes that are powering down are considered off
func status() ([]string, []string, []string) {
	//run scontrol show node
	//Split on empty lines
	//Ignore nodes that don't have ActiveFeatures=cloud
	//Nodes with 'state=IDLE+CLOUD+POWER ' are off
	//'State=ALLOCATED#+CLOUD ' are turning on
	//'State=IDLE#+CLOUD ' are turning on
	//'State=IDLE+CLOUD' are on
	//'State=ALLOCATED+CLOUD ' are on
	//'State=IDLE+CLOUD+COMPLETING' are on
	//'State=IDLE+CLOUD+POWERING_DOWN ' are turning off
	//everything else should be considered broken

	var on, off, broken []string

	out, err := exec.Command(utils.Config("slurm.dir").Self()+"/bin/scontrol", "show", "node").Output()
	if err != nil {
		log.Printf("ERROR:slurm: Error getting node states %s\n", err)
	}
	//split output by which node it is about
	nodes := strings.Split(string(out), "\nNodeName=")
	// get rid of leading "NodeNames="
	nodes[0] = nodes[0][len("NodeNames"):len(nodes[0])]
	//iter over all nodes found
	for _, n := range nodes {
		//check is cloud node
		if m, err := regexp.Match("ActiveFeatures=cloud", []byte(n)); !m {
			continue
		} else if err != nil {
			log.Printf("ERROR:slurm: Could not check if node is cloud %s, %s\n", n, err)
		}
		//check if off matches powering_down too
		if m, err := regexp.Match("State=IDLE\\+CLOUD\\+POWER", []byte(n)); m {
			off = append(off, strings.SplitN(n, " ", 2)[0])
			continue
		} else if err != nil {
			log.Printf("Error:slurm: Error parsing node %s, %s\n", n, err)
		}

		//check if on
		str := "State=(IDLE|ALLOCATED)[#]?\\+CLOUD"
		if m, err := regexp.Match(str, []byte(n)); m {
			on = append(on, strings.SplitN(n, " ", 2)[0])
			continue
		} else if err != nil {
			log.Printf("Error:slurm: Error parsing node %s, %s\n", n, err)
		}

		broken = append(broken, strings.SplitN(n, " ", 2)[0])
	}
	return on, off, broken
}

//fixBroken resets the state of any slurm cloud nodes that slurm thinks are broken in some way
func fixBroken(nodes []string) {
	for _, n := range nodes {
		log.Printf("INFO:slurm: attempting to fix node '%s'\n", n)
		out, err := exec.Command(utils.Config("slurm.dir").Self()+"/bin/scontrol", "update", "State=power_down", "NodeName="+n).Output()
		if err != nil {
			log.Printf("ERROR:slurm: error powering down node: %s \n %s\n", out, err)
		}
		out, err = exec.Command(utils.Config("slurm.dir").Self()+"/bin/scontrol", "update", "State=resume", "NodeName="+n).Output()
		if err != nil {
			log.Printf("ERROR:slurm: error resuming node: %s \n %s\n", out, err)
		}
	}
}
