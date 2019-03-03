package main

import (
	"log"
	"math/rand"
	"os"
	"os/exec"
	"time"
)

func checkDisabled() {

	lockfilelocation := "/opt/puppetlabs/puppet/cache/state/agent_disabled.lock"
	if _, err := os.Stat(lockfilelocation); err == nil {
		println("Diabled Lock found we don't like diabled run")
		os.Remove(lockfilelocation)
		log.Println("Lock file removed")
	}

}

func checkLockFile() {

	lockfilelocation := "/opt/puppetlabs/puppet/cache/state/agent_catalog_run.lock"
	if filestat, err := os.Stat(lockfilelocation); err == nil {

		now := time.Now()
		cutoff := 25 * time.Minute
		if diff := now.Sub(filestat.ModTime()); diff > cutoff {
			log.Printf("Deleting %s which is %s old\n", filestat.Name(), diff)
			os.Remove(lockfilelocation)
		} else {
			log.Printf("Found lock file %s which is less than %s old aborting run\n", filestat.Name(), cutoff)
			runPuppet()
		}

	} else {
		log.Printf("No run lock found proceeding")
	}
}
func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func runPuppet() {
	log.Print("Running Puppet")
	myrand := random(15, 45)
	log.Printf("Delaying puppet-nanny run by %d minutes", myrand)
	time.Sleep(time.Duration(myrand) * time.Minute)
	checkDisabled()
	checkLockFile()

	cmd := exec.Command("/opt/puppetlabs/puppet/bin/puppet", "agent", "-t")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	runPuppet()
}

func santityChecks() {
	//Check users priv if not root exit
	log.Print("Checking for root")
	if os.Getuid() != 0 {
		log.Fatalf("puppet-nanny needs to be ran as root:")
	}
	runPuppet()
}

func main() {
	santityChecks()
}
