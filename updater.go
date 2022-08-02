package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type UpdateConfig struct {
	fileName        string
	ExitOnError     bool     `yaml:"exitOnError"`
	ServiceName     string   `yaml:"serviceName"`
	PreUpgradeCmds  []string `yaml:"preUpgradeCmds"`
	UpgradeCmd      string   `yaml:"upgradeCmd"`
	PostUpgradeCmds []string `yaml:"postUpgradeCmds"`
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
		configFile.ExecuteUpdate()
	}
}

func buildUpdateConfigFiles(configPath string, isDirectory bool) []*UpdateConfig {
	configFileNames := make([]string, 0)
	if isDirectory {
		entries, err := os.ReadDir(configPath)
		if err != nil {
			log.Fatalln(err)
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
			log.Printf("ERROR - Failed to open file: %s - %v\n", filename, fileErr)
			continue
		}

		var configFile UpdateConfig
		decodeErr := yaml.NewDecoder(file).Decode(&configFile)
		configFile.fileName = filename
		if decodeErr != nil {
			log.Printf("ERROR - Failed to decode config file: %s - %v\n", filename, decodeErr)
			continue
		}

		validateErr := configFile.validate()
		if validateErr == nil {
			configFiles = append(configFiles, &configFile)
		} else {
			log.Printf("ERROR - Config file is invalid: %v\n", validateErr)
		}
	}

	return configFiles
}

func (uc *UpdateConfig) ExecuteUpdate() {
	log.Printf("Executing Update for Service: %v\n", uc.ServiceName)
	exitOnErr := false
	for _, preUpgradeCmd := range uc.PreUpgradeCmds {
		output, err := executeCommand(preUpgradeCmd)
		log.Println(output)
		if err == nil {
			continue
		}

		log.Printf("ERROR - Failed to execute pre-upgrade command: %v\n", err)
		if uc.ExitOnError {
			exitOnErr = true
			break
		}
	}

	if exitOnErr {
		output, err := executeCommand("systemctl restart " + uc.ServiceName)
		log.Println(output)
		if err != nil {
			log.Printf("ERROR - Failed to execute service restart command: %v\n", err)
		}

		return
	}

	output, err := executeCommand(uc.UpgradeCmd)
	log.Println(output)
	if err != nil {
		log.Fatalf("ERROR - Failed to execute upgrade command: %v\n", err)
	}

	for _, postUpgradeCmd := range uc.PostUpgradeCmds {
		output, err := executeCommand(postUpgradeCmd)
		if err == nil {
			log.Println(output)
			return
		}

		log.Printf("ERROR - Failed to execute post-upgrade command: %v\n", err)
	}

	output, err = executeCommand("systemctl restart " + uc.ServiceName)
	log.Println(output)
	if err != nil {
		log.Printf("ERROR - Failed to execute service restart command: %v\n", err)
	}
}

func (uc UpdateConfig) validate() error {
	if uc.ServiceName == "" {
		return fmt.Errorf("missing service name on config")
	}

	if uc.UpgradeCmd == "" {
		return fmt.Errorf("missing upgrade command on config")
	}

	return nil
}

func executeCommand(command string) (string, error) {
	updateCmd := exec.Command(command)
	output, err := updateCmd.Output()
	if err == nil {
		return "", err
	}
	return string(output), err
}

func printHelpMessage() {
	fmt.Println("This is a CLI for updating Matrix components registered as systemd services. It assumes apt is installed on server.")
	fmt.Printf("Command: matrix-systemd-updater { -help | [-directory] filepath }\n\n")
	fmt.Println("Arguments:")
	fmt.Println(
		"filepath: `filepath` specifies the path to the YAML file that contains the configuration " +
			"for this update. This YAML file must contain a field called `serviceName` (the systemd service to update) " +
			"and a field called `upgradeCmd` (the command to run to update the service).",
	)
	fmt.Println("Options:")
	fmt.Println("-directory: Optional. indicates the filepath is a directory containing the update configfiles")
	fmt.Println("-help: Optional. Print this messgae and exit")
}
