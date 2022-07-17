package homeserver

import (
	"log"
	"os/exec"
)

func ExecuteHomeServerCron(serviceName, packageName string) {
	updateCmd := exec.Command("apt", "update")
	output, err := updateCmd.Output()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(string(output))

	var upgradeCmd *exec.Cmd
	if packageName == "" {
		log.Println("WARN: No package name set. Attempting to upgrade matrix-synapse-py3")
		upgradeCmd = exec.Command("apt", "upgrade", "matrix-synapse-py3")
	} else {
		upgradeCmd = exec.Command("apt", "upgrade", packageName)
	}

	output, err = upgradeCmd.Output()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(string(output))

	var restartCmd *exec.Cmd
	if serviceName == "" {
		log.Println("WARN: No service name set. Attempting to restart systemd service matrix-synapse")
		restartCmd = exec.Command("systemctl", "restart", "matrix-synapse")
	} else {
		restartCmd = exec.Command("systemctl", "restart", serviceName)
	}

	_, err = restartCmd.Output()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Matrix Server Updated")
}
