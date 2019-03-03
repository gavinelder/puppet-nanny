package main

import (
	"bytes"
	"fmt"
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
		fmt.Println("Lock file removed")
	}

}

func checkLockFile() {

	lockfilelocation := "/opt/puppetlabs/puppet/cache/state/agent_catalog_run.lock"
	if filestat, err := os.Stat(lockfilelocation); err == nil {

		now := time.Now()
		cutoff := 25 * time.Minute
		if diff := now.Sub(filestat.ModTime()); diff > cutoff {
			fmt.Printf("Deleting %s which is %s old\n", filestat.Name(), diff)
			os.Remove(lockfilelocation)
		} else {
			fmt.Printf("Found lock file %s which is less than %s old aborting run\n", filestat.Name(), cutoff)
			runPuppet()
		}

	} else {
		println("No run lock found proceeding")
	}
}
func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func runPuppet() {
	checkDisabled()
	checkLockFile()
	cmd := exec.Command("/opt/puppetlabs/puppet/bin/puppet", "agent", "-t")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)
	runPuppet()
}

func checkForAdmin() {
	//Check users priv if not root exit
	println("Checking for root")

}

func main() {
	checkForAdmin()
	myrand := random(5, 30)
	fmt.Printf("Delaying puppet-nanny run by %d minutes", myrand)
	time.Sleep(time.Duration(myrand) * time.Minute)
	runPuppet()
}
