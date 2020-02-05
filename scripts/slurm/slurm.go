package slurm

import(
	"log"
	"github.com/Will-Shanks/terraformScripts/utils"
	"os/exec"
	"strings"
)

const SLURM_DIR="/opt/slurm/latest/"

func Add(inst utils.Instance){
	log.Printf("INFO:slurm: adding %s\n", inst)
}

func Del(inst string){
	log.Printf("INFO:slurm: removing %s\n", inst)
}

func NodeNames(glob string) []string{
	log.Printf("DEBUG:slurm: expanding glob %s\n", glob)
	out, err := exec.Command(SLURM_DIR+"/bin/scontrol", "show", "hostname", glob).Output()
	if err != nil{
		log.Printf("ERROR:slurm: Error expanding nodename glob %s\n", err)
	}
	names := strings.Split(string(out), "\n")
	return names[0:len(names)-1]
}
