package main

import (
	"log"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func checkDisabled() error {
	// Check for a diabled lockfile and if present remove to allow puppet to run.
	lockFileLocation := ""
	switch goos := runtime.GOOS; goos {
	case "darwin", "linux":
		lockFileLocation = "/opt/puppetlabs/puppet/cache/state/agent_disabled.lock"
	case "windows":
		lockFileLocation = "C:\\ProgramData\\PuppetLabs\\puppet\\cache\\state\\agent_disabled.lock"
	}
	if _, err := os.Stat(lockFileLocation); err == nil {
		log.Println("Disable lock found, removing")
		if err := os.Remove(lockFileLocation); err != nil {
			log.Printf("Unable to remove lockfile %s", lockFileLocation)
			return err
		}
		log.Println("Lock file removed")
		return nil
	}
	return nil
}

func checkIsPuppetInstalled() {
	// Check for the puppet binary.
	puppetBinLocation := ""
	switch goos := runtime.GOOS; goos {
	case "darwin", "linux":
		puppetBinLocation = "/opt/puppetlabs/puppet/bin/puppet"
	case "windows":
		puppetBinLocation = "C:\\Program Files\\Puppet Labs\\Puppet\\bin\\puppet.bat"
	default:
		log.Fatalf("Err OS %s not supported.", goos)
	}

	log.Printf("Puppet binary found at %s", puppetBinLocation)

	if _, err := os.Stat(puppetBinLocation); err != nil {
		log.Fatalf("Puppet binary not found at %s", puppetBinLocation)
	}

}

func checkRunLockFile() error {
	// Check for puppet run lock & remove if > 25 mins old.
	lockFileLocation := ""
	switch goos := runtime.GOOS; goos {
	case "darwin", "linux":
		lockFileLocation = "/opt/puppetlabs/puppet/cache/state/agent_catalog_run.lock"
	case "windows":
		lockFileLocation = "C:\\ProgramData\\PuppetLabs\\puppet\\cache\\state\\agent_catalog_run.lock"
	}
	if filestat, err := os.Stat(lockFileLocation); err == nil {

		now := time.Now()
		cutoff := 25 * time.Minute
		if diff := now.Sub(filestat.ModTime()); diff > cutoff {
			if err := os.Remove(lockFileLocation); err != nil {
				log.Fatalf("Unable to remove lockfile %s", lockFileLocation)
				return err
			}
			log.Printf("Deleting %s which is %s old\n", filestat.Name(), diff)
		} else {
			log.Printf("Found lock file %s which is less than %s old aborting run\n", filestat.Name(), cutoff)
			return err
		}

	} else {
		log.Printf("No run lock found proceeding")
		return nil
	}
	return nil
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func runPuppet() {
	// Sleep until we need to run
	myrand := random(15, 45)
	log.Printf("Delaying puppet-nanny run by %d minutes", myrand)
	time.Sleep(time.Duration(myrand) * time.Minute)
	// Carry out necessary checks to see if we should run puppet at this time.
	err := checkRunLockFile()
	if err != nil {
		return
	}
	// Check if puppet is disabled, checked as part of every run as someone may disable at any point.
	err = checkDisabled()
	if err != nil {
		return
	}

	cmd := exec.Command("")
	switch goos := runtime.GOOS; goos {
	case "darwin", "linux":
		cmd = exec.Command("/opt/puppetlabs/puppet/bin/puppet", "agent", "-t")
	case "windows":
		cmd = exec.Command("C:\\Program Files\\Puppet Labs\\Puppet\\bin\\puppet.bat", "agent", "-t")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	log.Println("Running Puppet")
}

func priveledgeCheck() {
	// Ensure puppet is running with elevated privledge.
	switch goos := runtime.GOOS; goos {
	case "darwin", "linux":
		if os.Getuid() != 0 {
			log.Fatalf("puppet-nanny needs to be ran as root")
		}
	case "windows":
		_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
		if err != nil {
			log.Fatalf("puppet-nanny needs to be ran with admin privledges")
		}
	}
}

func main() {

	priveledgeCheck()
	for {
		checkIsPuppetInstalled()
		runPuppet()
	}
}
