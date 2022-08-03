package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

var CMD_REGEX = regexp.MustCompile(`(?P<cmd>\S*)\s?(?P<args>.*)`)

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
		err := configFile.ExecuteUpdate()
		if err != nil {
			log.Printf("ERROR - Failed up to update systemd component %s due to error: %v\n", configFile.ServiceName, err)
		}
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

func (uc *UpdateConfig) ExecuteUpdate() error {
	log.Printf("Executing Update for Service: %v\n", uc.ServiceName)
	for _, preUpgradeCmd := range uc.PreUpgradeCmds {
		output, err := executeCommand(preUpgradeCmd)
		log.Println(output)
		if err == nil {
			continue
		}

		if uc.ExitOnError {
			return fmt.Errorf("failed to execute pre-upgrade command: %v", err)
		} else {
			log.Printf("ERROR - Failed to execute pre-upgrade command: %v\n", err)
		}
	}

	output, err := executeCommand(uc.UpgradeCmd)
	log.Println(output)
	if err != nil {
		return fmt.Errorf("failed to execute upgrade command: %v", err)
	}

	for _, postUpgradeCmd := range uc.PostUpgradeCmds {
		output, err := executeCommand(postUpgradeCmd)
		log.Println(output)
		if err == nil {
			continue
		}

		log.Printf("ERROR - Failed to execute post-upgrade command: %v\n", err)
	}

	output, err = executeCommand("systemctl restart " + uc.ServiceName)
	log.Println(output)
	if err != nil {
		return fmt.Errorf("failed to execute service restart command: %v", err)
	}
	return nil
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

func parseCommand(command string) map[string]string {
	matchesByKeyword := make(map[string]string)
	if !CMD_REGEX.MatchString(command) {
		return matchesByKeyword
	}

	matches := CMD_REGEX.FindStringSubmatch(command)
	cmdIndex := CMD_REGEX.SubexpIndex("cmd")
	argsIndex := CMD_REGEX.SubexpIndex("args")

	if cmdIndex > -1 && matches[cmdIndex] != "" {
		matchesByKeyword["cmd"] = matches[cmdIndex]
	}

	if argsIndex > -1 && matches[argsIndex] != "" {
		matchesByKeyword["args"] = matches[argsIndex]
	}

	return matchesByKeyword
}

func executeCommand(command string) (string, error) {
	var updateCmd *exec.Cmd
	parsedCommand := parseCommand(command)
	cmd, cmdOk := parsedCommand["cmd"]
	if !cmdOk {
		return "", fmt.Errorf("unable to parse command: %s", command)
	}

	args, argsOk := parsedCommand["args"]
	if argsOk {
		updateCmd = exec.Command(cmd, args)
	} else {
		updateCmd = exec.Command(cmd)
	}

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
