package main

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestMissingServiceBadConfig(t *testing.T) {
	configPath := "test_configs/bad_config_missing_service.yaml"
	configs := buildUpdateConfigFiles(configPath, false)
	if len(configs) != 0 {
		t.Logf("Expected config path %s to generate %d configs. Actual %d\n", configPath, 0, len(configs))
		t.Fail()
	}
}

func TestMissingUpgradeBadConfig(t *testing.T) {
	configPath := "test_configs/bad_config_missing_upgrade.yaml"
	configs := buildUpdateConfigFiles(configPath, false)
	if len(configs) != 0 {
		t.Logf("Expected config path %s to generate %d configs. Actual %d\n", configPath, 0, len(configs))
		t.Fail()
	}
}

func TestGoodConfig(t *testing.T) {
	configPath := "test_configs/good_config.yaml"
	configs := buildUpdateConfigFiles(configPath, false)
	if len(configs) != 1 {
		t.Logf("Expected config path %s to generate %d configs. Actual %d\n", configPath, 1, len(configs))
		t.FailNow()
	}

	assertGoodConfig(configs[0], configPath, t)
}

func TestDirectoryConfigParsing(t *testing.T) {
	directoryPath := "test_configs"
	goodConfigPath := "test_configs/good_config.yaml"
	configs := buildUpdateConfigFiles(directoryPath, true)
	if len(configs) != 1 {
		t.Logf("Expected config path %s to generate %d configs. Actual %d\n", directoryPath, 1, len(configs))
		t.FailNow()
	}

	assertGoodConfig(configs[0], goodConfigPath, t)
}

func TestParsesFullyQualifiedCommand(t *testing.T) {
	qualifiedCommand := "/bin/echo 'first pre-upgrade command'"
	parsedCommand := parseCommand(qualifiedCommand)
	cmd, cmdOk := parsedCommand["cmd"]
	args, argOk := parsedCommand["args"]
	if !cmdOk {
		t.Logf("Expected 'cmd' key to exist in map for command: %s\n", qualifiedCommand)
		t.FailNow()
	}

	if cmd != "/bin/echo" {
		t.Logf("Expected 'cmd' key to correspond to value %s: Actual: %s\n", "/bin/echo", cmd)
		t.Fail()
	}

	if !argOk {
		t.Logf("Expected 'args' key to exist in map for command: %s\n", qualifiedCommand)
		t.FailNow()
	}

	if args != "'first pre-upgrade command'" {
		t.Logf("Expected 'args' key to correspond to value %s: Actual: %s\n", "'first pre-upgrade command'", args)
		t.Fail()
	}
}

func TestParsesFullyQualifiedCommandNoArgs(t *testing.T) {
	qualifiedCommand := "/bin/echo"
	parsedCommand := parseCommand(qualifiedCommand)
	cmd, cmdOk := parsedCommand["cmd"]
	_, argOk := parsedCommand["args"]
	if !cmdOk {
		t.Logf("Expected 'cmd' key to exist in map for command: %s\n", qualifiedCommand)
		t.FailNow()
	}

	if cmd != "/bin/echo" {
		t.Logf("Expected 'cmd' key to correspond to value %s: Actual: %s\n", "/bin/echo", cmd)
		t.Fail()
	}

	if argOk {
		t.Logf("Expected no 'args' key in map for command: %s\n", qualifiedCommand)
		t.FailNow()
	}
}

func TestParsesShorthandCommandNoArgs(t *testing.T) {
	qualifiedCommand := "echo"
	parsedCommand := parseCommand(qualifiedCommand)
	cmd, cmdOk := parsedCommand["cmd"]
	_, argOk := parsedCommand["args"]
	if !cmdOk {
		t.Logf("Expected 'cmd' key to exist in map for command: %s\n", qualifiedCommand)
		t.FailNow()
	}

	if cmd != "echo" {
		t.Logf("Expected 'cmd' key to correspond to value %s: Actual: %s\n", "/bin/echo", cmd)
		t.Fail()
	}

	if argOk {
		t.Logf("Expected no 'args' key in map for command: %s\n", qualifiedCommand)
		t.FailNow()
	}
}

func TestParsesShorthandCommand(t *testing.T) {
	qualifiedCommand := "echo 'first pre-upgrade command'"
	parsedCommand := parseCommand(qualifiedCommand)
	cmd, cmdOk := parsedCommand["cmd"]
	args, argOk := parsedCommand["args"]
	if !cmdOk {
		t.Logf("Expected 'cmd' key to exist in map for command: %s\n", qualifiedCommand)
		t.FailNow()
	}

	if cmd != "echo" {
		t.Logf("Expected 'cmd' key to correspond to value %s: Actual: %s\n", "echo", cmd)
		t.Fail()
	}

	if !argOk {
		t.Logf("Expected 'args' key to exist in map for command: %s\n", qualifiedCommand)
		t.FailNow()
	}

	if args != "'first pre-upgrade command'" {
		t.Logf("Expected 'args' key to correspond to value %s: Actual: %s\n", "'first pre-upgrade command'", args)
		t.Fail()
	}
}

func TestParsesShorthandCommandMultiArgs(t *testing.T) {
	qualifiedCommand := "apt upgrade matrix-synapse-py3"
	parsedCommand := parseCommand(qualifiedCommand)
	cmd, cmdOk := parsedCommand["cmd"]
	args, argOk := parsedCommand["args"]
	if !cmdOk {
		t.Logf("Expected 'cmd' key to exist in map for command: %s\n", qualifiedCommand)
		t.FailNow()
	}

	if cmd != "apt" {
		t.Logf("Expected 'cmd' key to correspond to value %s: Actual: %s\n", "apt", cmd)
		t.Fail()
	}

	if !argOk {
		t.Logf("Expected 'args' key to exist in map for command: %s\n", qualifiedCommand)
		t.FailNow()
	}

	if args != "upgrade matrix-synapse-py3" {
		t.Logf("Expected 'args' key to correspond to value %s: Actual: %s\n", "upgrade matrix-synapse-py3", args)
		t.Fail()
	}
}

func assertGoodConfig(actualConfig *UpdateConfig, configPath string, t *testing.T) {
	if actualConfig.ExitOnError {
		t.Logf("Expected config to not exit on error but it does")
		t.Fail()
	}

	if actualConfig.ServiceName != "test-matrix-component" {
		t.Logf("Expected config service name to to be test-matrix-component. Actual: %s\n", actualConfig.ServiceName)
		t.Fail()
	}

	if actualConfig.UpgradeCmd != "echo 'this is an upgradeCmd'" {
		t.Logf("Expected config upgrade command to be %s. Actual: %s\n", "echo 'this is an upgradeCmd'", actualConfig.UpgradeCmd)
		t.Fail()
	}

	if actualConfig.fileName != filepath.Clean(configPath) {
		t.Logf("Expected file name to be %s. Actual: %s\n", configPath, actualConfig.fileName)
		t.Fail()
	}

	expectedPreUpgradeCmds := []string{"echo 'first pre-upgrade command'", "echo 'second pre-upgrade command'"}
	if !reflect.DeepEqual(actualConfig.PreUpgradeCmds, expectedPreUpgradeCmds) {
		t.Logf("Expected pre-upgrade commands to be %v. Actual: %v\n", expectedPreUpgradeCmds, actualConfig.PreUpgradeCmds)
		t.Fail()
	}

	expectedPostUpgradeCmds := []string{"echo 'only post-upgrade command'"}
	if !reflect.DeepEqual(actualConfig.PostUpgradeCmds, expectedPostUpgradeCmds) {
		t.Logf("Expected post-upgrade commands to be %v. Actual: %v\n", expectedPostUpgradeCmds, actualConfig.PostUpgradeCmds)
		t.Fail()
	}
}
