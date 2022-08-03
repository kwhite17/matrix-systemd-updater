package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
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

func main() {
	printHelp := flag.Bool("help", false, "Print the help message for the Matrix Systemd Updater")
	isDirectory := flag.Bool("configDirectory", false, "Provided file path for configurations correspond to directories")
	flag.Parse()

	configPath := flag.Arg(0)
	if *printHelp {
		printHelpMessage()
		os.Exit(0)
	}

	configFiles := buildUpdateConfigFiles(configPath, *isDirectory)
	for _, configFile := range configFiles {
		err := configFile.executeUpdate()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR - Failed up to update systemd component %s due to error: %v\n", configFile.ServiceName, err)
		}
	}
}

func buildUpdateConfigFiles(configPath string, isDirectory bool) []*UpdateConfig {
	configFileNames := make([]string, 0)
	if isDirectory {
		entries, err := os.ReadDir(configPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			} else {
				configFileNames = append(configFileNames, filepath.Clean(configPath+string(os.PathSeparator)+entry.Name()))
			}
		}
	} else {
		configFileNames = []string{filepath.Clean(configPath)}
	}

	configFiles := make([]*UpdateConfig, 0)
	for _, filename := range configFileNames {
		file, fileErr := os.Open(filename)
		if fileErr != nil {
			fmt.Fprintf(os.Stderr, "ERROR - Failed to open file: %s - %v\n", filename, fileErr)
			continue
		}

		var configFile UpdateConfig
		decodeErr := yaml.NewDecoder(file).Decode(&configFile)
		configFile.fileName = filename
		if decodeErr != nil {
			fmt.Fprintf(os.Stderr, "ERROR - Failed to decode config file: %s - %v\n", filename, decodeErr)
			continue
		}

		validateErr := configFile.validate()
		if validateErr == nil {
			configFiles = append(configFiles, &configFile)
		} else {
			fmt.Fprintf(os.Stderr, "ERROR - Config file is invalid: %v\n", validateErr)
		}
	}

	return configFiles
}

func (uc *UpdateConfig) executeUpdate() error {
	fmt.Fprintf(os.Stdout, "Executing Update for Service: %v\n", uc.ServiceName)
	if uc.WorkingDirectory != "" {
		wdErr := os.Chdir(uc.WorkingDirectory)
		if wdErr != nil {
			return wdErr
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

	output, err := executeCommand(uc.UpgradeCmd)
	fmt.Fprintln(os.Stdout, output)
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

	output, err = executeCommand(&ConfigCommand{Command: "systemctl", Args: []string{"restart", uc.ServiceName}})
	fmt.Fprintln(os.Stdout, output)
	if err != nil {
		return fmt.Errorf("failed to execute service restart command: %v", err)
	}
	return nil
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

func executeCommand(command *ConfigCommand) (string, error) {
	var updateCmd *exec.Cmd
	if command.Args == nil || len(command.Args) == 0 {
		updateCmd = exec.Command(command.Command)
	} else {
		updateCmd = exec.Command(command.Command, command.Args...)
	}

	output, err := updateCmd.Output()
	return string(output), err
}

func printHelpMessage() {
	fmt.Fprintln(os.Stdout, "This is a CLI for updating Matrix components registered as systemd services. It assumes apt is installed on server.")
	fmt.Fprintf(os.Stdout, "Command: matrix-systemd-updater { -help | [-configDirectory] filepath }\n\n")
	fmt.Fprintln(os.Stdout, "Arguments:")
	fmt.Fprintln(
		os.Stdout,
		"filepath: `filepath` specifies the path to the YAML file that contains the configuration "+
			"for this update. This YAML file must contain a field called `serviceName` (the systemd service to update) "+
			"and a field called `upgradeCmd` (the command to run to update the service).",
	)
	fmt.Fprintln(os.Stdout, "Options:")
	fmt.Fprintln(os.Stdout, "-configDirectory: Optional. indicates the filepath is a directory containing the update configfiles")
	fmt.Fprintln(os.Stdout, "-help: Optional. Print this messgae and exit")
}
