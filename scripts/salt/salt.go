package salt

import(
	"log"
	"os/exec"
)

func Apply(hostname string) {
	log.Printf("Salt Apply %s\n", hostname);
}

func Del(hostname string) {
	exec.Command("sudo","/usr/bin/salt-key", "-d", "-y", hostname).Run()
}
