package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func checkDisabled() {
	fileStat, err := os.Stat("/opt/puppetlabs/puppet/bin/puppet")

	if err != nil {
		log.Fatal("File does not exist")
	}
	fileStat.ModTime()
}

func checkLockFile() {

}

func main() {
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

}
