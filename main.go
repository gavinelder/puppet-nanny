package main

import (
	"log"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func checkDisabled() {
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
		err := os.Remove(lockFileLocation)
		if err != nil {
			log.Fatalf("Unable to remove lockfile %s", lockFileLocation)
		}
		log.Println("Lock file removed")
	}

}

func checkPuppetInstalled() {
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

func checkRunLockFile() {
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
			log.Printf("Deleting %s which is %s old\n", filestat.Name(), diff)
			err := os.Remove(lockFileLocation)
			if err != nil {
				log.Fatalf("Unable to remove lockfile %s", lockFileLocation)
			}
		} else {
			log.Printf("Found lock file %s which is less than %s old aborting run\n", filestat.Name(), cutoff)
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
	// Run puppet.
	myrand := random(15, 45)
	log.Printf("Delaying puppet-nanny run by %d minutes", myrand)
	time.Sleep(time.Duration(myrand) * time.Minute)
	cmd := exec.Command("")
	switch goos := runtime.GOOS; goos {
	case "darwin", "linux":
		cmd = exec.Command("/opt/puppetlabs/puppet/bin/puppet", "agent", "-t")
	case "windows":
		cmd = exec.Command("C:\\Program Files\\Puppet Labs\\Puppet\\bin\\puppet.bat", "agent", "-t")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Println("Running Puppet")
	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
}

func priveledgeCheck() {
	// Ensure puppet is running with elevated privledge.
	switch goos := runtime.GOOS; goos {
	case "darwin", "linux":
		if os.Getuid() != 0 {
			log.Fatalf("puppet-nanny needs to be ran as root:")
		}
	case "windows":
		_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
		if err != nil {
			log.Fatalf("puppet-nanny needs to be ran as Admin:")
		}
	}
}

func main() {

	for {
		priveledgeCheck()
		checkPuppetInstalled()
		checkRunLockFile()
		checkDisabled()
		runPuppet()
	}
}
