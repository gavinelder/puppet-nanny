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

type puppetRunConfig struct {
	runLockFileLocation     string
	disableLockFileLocation string
	puppetBinLocation       string
	environment             *string
	now                     *bool
}

func removeAgentDisableLock(disableLockFileLocation string) error {
	// Check for a disabled lockfile and if present remove to allow puppet to run.
	if _, err := os.Stat(disableLockFileLocation); err == nil {
		log.Print("Disable lock found, removing.\n")
		if err := os.Remove(disableLockFileLocation); err != nil {
			log.Printf("Unable to remove lockfile %s\n", disableLockFileLocation)
			return err
		}
		log.Print("Disable lock removed.\n")
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
		log.Print("No run lock found proceeding. \n")
	}
	return nil
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func runPuppet(runConfig puppetRunConfig) {
	if _, err := os.Stat(runConfig.puppetBinLocation); err != nil {
		log.Fatalf("Puppet binary not found at %s. \n", runConfig.puppetBinLocation)
	}
	// Sleep until we need to run
	delay := 0
	if !*runConfig.now {
		delay = random(15, 45)
	}
	log.Printf("Delaying puppet-nanny run by %d minutes", delay)
	time.Sleep(time.Duration(delay) * time.Minute)
	// Carry out necessary checks to see if we should run puppet at this time.
	if err := checkRunState(runConfig.runLockFileLocation); err != nil {
		return
	}
	// Check if puppet is disabled, checked as part of every run as someone may disable at any point.
	if err := removeAgentDisableLock(runConfig.disableLockFileLocation); err != nil {
		return
	}
	runArgs := []string{"agent", "-t"}
	if *runConfig.environment != "" {
		runArgs = append(runArgs, "--environment", *runConfig.environment)
	}
	cmd := exec.Command(runConfig.puppetBinLocation, runArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("%v failed with %s.\n", cmd, err)
	}
	log.Print("Puppet run succeeded.\n")
}

func main() {
	runConfig := puppetRunConfig{}

	runConfig.environment = flag.String("environment", "", "Specifies the environment puppet should run using.")
	runConfig.now = flag.Bool("now", false, "Runs puppet now.")
	flag.Parse()

	switch goos := runtime.GOOS; goos {
	case "darwin", "linux":
		if os.Getuid() != 0 {
			log.Fatal("puppet-nanny needs to be ran as root")
		}
		runConfig.puppetBinLocation = "/opt/puppetlabs/bin/puppet"
		runConfig.runLockFileLocation = "/opt/puppetlabs/puppet/cache/state/agent_catalog_run.lock"
		runConfig.disableLockFileLocation = "/opt/puppetlabs/puppet/cache/state/agent_disabled.lock"
	case "windows":
		if _, err := os.Open("\\\\.\\PHYSICALDRIVE0"); err != nil {
			log.Fatal("puppet-nanny needs to be ran with admin privledges. \n")
		}
		runConfig.puppetBinLocation = "C:\\Program Files\\Puppet Labs\\Puppet\\bin\\puppet.bat"
		runConfig.runLockFileLocation = "C:\\ProgramData\\PuppetLabs\\puppet\\cache\\state\\agent_catalog_run.lock"
		runConfig.disableLockFileLocation = "C:\\ProgramData\\PuppetLabs\\puppet\\cache\\state\\agent_disabled.lock"
	default:
		log.Fatalf("%v is not officially supported please raise an issue at https://github.com/gavinelder/puppet-nannyy .\n", goos)
	}
	for {
		runPuppet(runConfig)
	}
}
