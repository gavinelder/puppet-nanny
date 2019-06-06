package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func removeAgentDisableLock(disableLockFileLocation string) error {
	// Check for a disabled lockfile and if present remove to allow puppet to run.
	if _, err := os.Stat(disableLockFileLocation); err == nil {
		log.Printf("Disable lock found, removing.\n")
		if err := os.Remove(disableLockFileLocation); err != nil {
			log.Printf("Unable to remove lockfile %s\n", disableLockFileLocation)
			return err
		}
		log.Printf("Disable lock removed.\n")
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
				log.Fatalf("Unable to remove lockfile %s.\n", runLockFileLocation)
			}
			log.Printf("Deleting run lock %s which is %s old.\n", filestat.Name(), diff)
		} else {
			log.Printf("Found lock file %s which is less than %s old aborting run.\n", filestat.Name(), cutoff)
			return err
		}
	} else {
		log.Printf("No run lock found proceeding. \n")
	}
	return nil
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func runPuppet(puppetBinLocation, runLockFileLocation, disableLockFileLocation string, runNowFlag *bool, envFlag *string) {
	if _, err := os.Stat(puppetBinLocation); err != nil {
		log.Fatalf("Puppet binary not found at %s. \n", puppetBinLocation)
	}
	// Sleep until we need to run

	delay := 0
	if !*runNowFlag {
		delay = random(15, 45)
	}
	log.Printf("Delaying puppet-nanny run by %d minutes", delay)
	time.Sleep(time.Duration(delay) * time.Minute)
	// Carry out necessary checks to see if we should run puppet at this time.
	if err := checkRunState(runLockFileLocation); err != nil {
		return
	}
	// Check if puppet is disabled, checked as part of every run as someone may disable at any point.
	if err := removeAgentDisableLock(disableLockFileLocation); err != nil {
		return
	}
	runArgs := []string{"agent", "-t"}
	if *envFlag != "" {
		runArgs = append(runArgs, "--environment", *envFlag)
	}
	cmd := exec.Command(puppetBinLocation, runArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("cmd.Run() failed with %s.\n", err)
	}
	log.Printf("Puppet run succeeded.\n")
}

func main() {
	environmentFlag := flag.String("environment", "", "Specifies the environment puppet should run using.")
	runNowFlag := flag.Bool("now", false, "Runs puppet now.")
	flag.Parse()
	runLockFileLocation := ""
	disableLockFileLocation := ""
	puppetBinLocation := ""

	switch goos := runtime.GOOS; goos {
	case "darwin", "linux":
		if os.Getuid() != 0 {
			log.Fatalf("puppet-nanny needs to be ran as root")
		}
		puppetBinLocation = "/opt/puppetlabs/bin/puppet"
		runLockFileLocation = "/opt/puppetlabs/puppet/cache/state/agent_catalog_run.lock"
		disableLockFileLocation = "/opt/puppetlabs/puppet/cache/state/agent_disabled.lock"
	case "windows":
		if _, err := os.Open("\\\\.\\PHYSICALDRIVE0"); err != nil {
			log.Fatalf("puppet-nanny needs to be ran with admin privledges. \n")
		}
		puppetBinLocation = "C:\\Program Files\\Puppet Labs\\Puppet\\bin\\puppet.bat"
		runLockFileLocation = "C:\\ProgramData\\PuppetLabs\\puppet\\cache\\state\\agent_catalog_run.lock"
		disableLockFileLocation = "C:\\ProgramData\\PuppetLabs\\puppet\\cache\\state\\agent_disabled.lock"
	default:
		log.Fatalf("OS not supported.\n")
	}
	log.Printf("Puppet binary set as %s.\n", puppetBinLocation)
	for {
		runPuppet(puppetBinLocation, runLockFileLocation, disableLockFileLocation, runNowFlag, environmentFlag)
	}
}
