package main

import (
	"fmt"
	"os"
	"os/exec"
)

type UpdateConfig struct {
	fileName         string
	WorkingDirectory string           `yaml:"workingDirectory"`
	ExitOnError      bool             `yaml:"exitOnError"`
	ServiceName      string           `yaml:"serviceName"`
	PreUpgradeCmds   []*ConfigCommand `yaml:"preUpgradeCmds"`
	UpgradeCmd       *ConfigCommand   `yaml:"upgradeCmd"`
	PostUpgradeCmds  []*ConfigCommand `yaml:"postUpgradeCmds"`
}

type ConfigCommand struct {
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
}

func (uc *UpdateConfig) executeUpdate() error {
	fmt.Fprintf(os.Stdout, "Executing Update for Service: %v\n", uc.ServiceName)
	if uc.WorkingDirectory != "" {
		wdErr := os.Chdir(uc.WorkingDirectory)
		if wdErr != nil {
			return fmt.Errorf("failed to change working directory to %s due to error %v", uc.WorkingDirectory, wdErr)
		}
	}

	for _, preUpgradeCmd := range uc.PreUpgradeCmds {
		output, err := executeCommand(preUpgradeCmd)
		fmt.Fprintln(os.Stdout, output)
		if err == nil {
			continue
		}

		if uc.ExitOnError {
			return fmt.Errorf("failed to execute pre-upgrade command: %v", err)
		} else {
			fmt.Fprintf(os.Stderr, "ERROR - Failed to execute pre-upgrade command: %v\n", err)
		}
	}
	fmt.Fprintln(os.Stdout, "Pre-upgrade commands complete. Executing upgrade...")

	output, err := executeCommand(uc.UpgradeCmd)
	fmt.Fprintln(os.Stdout, output)
	fmt.Fprintln(os.Stdout, "Upgrade command complete. Performing post-update commands...")
	if err != nil {
		return fmt.Errorf("failed to execute upgrade command: %v", err)
	}

	for _, postUpgradeCmd := range uc.PostUpgradeCmds {
		output, err := executeCommand(postUpgradeCmd)
		fmt.Fprintln(os.Stdout, output)
		if err == nil {
			continue
		}

		fmt.Fprintf(os.Stderr, "ERROR - Failed to execute post-upgrade command: %v\n", err)
	}
	fmt.Fprintln(os.Stdout, "Post-upgrade commands complete. Restarting service...")

	output, err = executeCommand(&ConfigCommand{Command: "systemctl", Args: []string{"restart", uc.ServiceName}})
	fmt.Fprintln(os.Stdout, output)
	if err != nil {
		return fmt.Errorf("failed to execute service restart command: %v", err)
	}

	fmt.Fprintf(os.Stdout, "Upgrade of service %s complete!\n", uc.ServiceName)
	return nil
}

func executeCommand(command *ConfigCommand) (string, error) {
	var updateCmd *exec.Cmd
	if command.Args == nil || len(command.Args) == 0 {
		updateCmd = exec.Command(command.Command)
	} else {
		updateCmd = exec.Command(command.Command, command.Args...)
	}

	output, err := updateCmd.CombinedOutput()
	return string(output), err
}

func (uc UpdateConfig) validate() error {
	if uc.ServiceName == "" {
		return fmt.Errorf("missing service name on config")
	}

	if uc.UpgradeCmd == nil || uc.UpgradeCmd.Command == "" {
		return fmt.Errorf("missing upgrade command on config")
	}

	if uc.WorkingDirectory != "" && !uc.ExitOnError {
		return fmt.Errorf("non-nil or non-empty working directory must exit on any pre-ugrade errors")
	}

	return nil
}
