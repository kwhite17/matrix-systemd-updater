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
