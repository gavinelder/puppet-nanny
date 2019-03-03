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

	lockfilelocation := ""
	switch os := runtime.GOOS; os {
	case "darwin", "linux":
		lockfilelocation = "/opt/puppetlabs/puppet/cache/state/agent_disabled.lock"
	case "windows":
		lockfilelocation = "C:\\ProgramData\\PuppetLabs\\puppet\\cache\\state\\agent_disabled.lock"
	}
	if _, err := os.Stat(lockfilelocation); err == nil {
		log.Println("Diabled Lock found we don't like diabled run")
		os.Remove(lockfilelocation)
		log.Println("Lock file removed")
	}

}

func checkPuppetInstalled() {

	puppetBinLocation := ""
	switch os := runtime.GOOS; os {
	case "darwin", "linux":
		puppetBinLocation = "/opt/puppetlabs/puppet/bin/puppet"
	case "windows":
		puppetBinLocation = "C:\\Program Files\\Puppet Labs\\Puppet\\bin\\puppet.bat"
	}
	log.Printf("Puppet binary found at %s", puppetBinLocation)
	if _, err := os.Stat(puppetBinLocation); err != nil {
		log.Fatalf("Puppet binary not found at %s", puppetBinLocation)
	}

}

func checkLockFile() {

	lockfilelocation := ""
	switch os := runtime.GOOS; os {
	case "darwin", "linux":
		lockfilelocation = "/opt/puppetlabs/puppet/cache/state/agent_catalog_run.lock"
	case "windows":
		lockfilelocation = "C:\\ProgramData\\PuppetLabs\\puppet\\cache\\state\\agent_catalog_run.lock"
	}
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
	checkPuppetInstalled()
	cmd := exec.Command("")
	switch os := runtime.GOOS; os {
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
	runPuppet()
}

func santityChecks() {
	//Check users priv if not root exit
	log.Print("Checking for root")
	switch OS := runtime.GOOS; OS {
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
	runPuppet()
}

func main() {
	santityChecks()
}
