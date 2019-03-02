package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func checkDisabled() {

	lockfilelocation := "/opt/puppetlabs/puppet/cache/state/agent_disabled.lock"
	if _, err := os.Stat(lockfilelocation); err == nil {
		// path/to/whatever exists
		println("Diabled Lock found we don't like diabled run")
		os.Remove(lockfilelocation)
		fmt.Println("Lock file removed")
	}

}

func checkLockFile() {

	lockfilelocation := "/opt/puppetlabs/puppet/cache/state/agent_catalog_run.lock"
	if _, err := os.Stat(lockfilelocation); err == nil {
		// path/to/whatever exists
		println("Run lock found removing")
		os.Remove(lockfilelocation)
		fmt.Println("Lock file removed")
	}
}

func runPuppet() {

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

}

func main() {
	checkDisabled()
	checkLockFile()
	runPuppet()

}
