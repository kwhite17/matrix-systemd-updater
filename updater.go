package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

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
