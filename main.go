package main

import (
	"log"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func removeAgentDisableLock(disableLockFileLocation string) error {
	// Check for a diabled lockfile and if present remove to allow puppet to run.

	if _, err := os.Stat(disableLockFileLocation); err == nil {
		log.Println("Disable lock found, removing")
		if err := os.Remove(disableLockFileLocation); err != nil {
			log.Printf("Unable to remove lockfile %s", disableLockFileLocation)
			return err
		}
		log.Println("Lock file removed")
		return nil
	}
	return nil
}

func checkRunState(runLockFileLocation string) error {
	// Check for puppet running state by inspecting the lockfile & remove if > 25 mins old.
	if filestat, err := os.Stat(runLockFileLocation); err == nil {
		now := time.Now()
		cutoff := 25 * time.Minute
		if diff := now.Sub(filestat.ModTime()); diff > cutoff {
			if err := os.Remove(runLockFileLocation); err != nil {
				log.Fatalf("Unable to remove lockfile %s", runLockFileLocation)
			}
			log.Printf("Deleting %s which is %s old\n", filestat.Name(), diff)
		} else {
			log.Printf("Found lock file %s which is less than %s old aborting run\n", filestat.Name(), cutoff)
			return err
		}
	} else {
		log.Printf("No run lock found proceeding")
	}
	return nil
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func runPuppet(puppetBinLocation, execPuppetCMD, runLockFileLocation, disableLockFileLocation string) {
	if _, err := os.Stat(puppetBinLocation); err != nil {
		log.Fatalf("Puppet binary not found at %s", puppetBinLocation)
	}
	// Sleep until we need to run
	myrand := random(15, 45)
	log.Printf("Delaying puppet-nanny run by %d minutes", myrand)
	time.Sleep(time.Duration(myrand) * time.Minute)
	// Carry out necessary checks to see if we should run puppet at this time.
	if err := checkRunState(runLockFileLocation); err != nil {
		return
	}
	// Check if puppet is disabled, checked as part of every run as someone may disable at any point.
	if err := removeAgentDisableLock(disableLockFileLocation); err != nil {
		return
	}
	cmd := exec.Command(execPuppetCMD)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	log.Println("Running Puppet")
}

func main() {
	// Set Environment variables
	execPuppetCMD := ""
	runLockFileLocation := ""
	disableLockFileLocation := ""
	puppetBinLocation := ""
	switch goos := runtime.GOOS; goos {
	case "darwin", "linux":
		if os.Getuid() != 0 {
			log.Fatalf("puppet-nanny needs to be ran as root")
		}
		puppetBinLocation = "/opt/puppetlabs/puppet/bin/puppet"
		execPuppetCMD = `"/opt/puppetlabs/puppet/bin/puppet", "agent", "-t"`
		runLockFileLocation = "/opt/puppetlabs/puppet/cache/state/agent_catalog_run.lock"
		disableLockFileLocation = "/opt/puppetlabs/puppet/cache/state/agent_disabled.lock"
	case "windows":
		if _, err := os.Open("\\\\.\\PHYSICALDRIVE0"); err != nil {
			log.Fatalf("puppet-nanny needs to be ran with admin privledges")
		}
		puppetBinLocation = "C:\\Program Files\\Puppet Labs\\Puppet\\bin\\puppet.bat"
		execPuppetCMD = `"C:\\Program Files\\Puppet Labs\\Puppet\\bin\\puppet.bat", "agent", "-t"`
		runLockFileLocation = "C:\\ProgramData\\PuppetLabs\\puppet\\cache\\state\\agent_catalog_run.lock"
		disableLockFileLocation = "C:\\ProgramData\\PuppetLabs\\puppet\\cache\\state\\agent_disabled.lock"
	default:
		log.Fatal("OS Not supported")
	}
	log.Printf("Puppet binary found at %s", puppetBinLocation)
	for {
		runPuppet(puppetBinLocation, execPuppetCMD, runLockFileLocation, disableLockFileLocation)
	}
}
