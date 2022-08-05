#  Matrix Systemd Updater

## Purpose
The Matrix Systemd Updater was initially created for the purpose of being able to, in conjuction with `crontab` automate the updating of Matrix homeservers and bridges. Ultimately, what was produced may well be a general purpose systemd updater utility.

## Usage
The usage of the updater is pretty straightforward:

matrix-systemd-updater { -help | [-configDirectory] configfilePath }

`-help`: prints the help message describing how to use the utility
`configDirectory`: indicates that `configfilePath` corresponds to a directory
`configFilePath`: fully qualified or relative path indicating the location of the configuration that details how to update a systemd service. If the `configDirectory` flag is not present, `configfilePath` must correspond to YAML file. When the flag is present, `configfilePath` corresponds to a directory containing valid YAML files. 

## Configuration File

### Properties
The configuration file contains the following YAML mappings:

* `workingDirectory` (optional) - The directory from which the updater should execute the commands specified in the configuration file
* `exitOnError` (optional) - Indicates if an upgrade of a service should be attempted when one of the pre-upgrade commands fails. The default value is `false` but must be set to `true` if `workingDirectory` is populated.
* `serviceName` (required) - the name of the systemd service being upgraded
* `preUpgradeCmds` (optional) - A sequence of `ConfigCommand`s that need to be executed before the the updater performs the service upgrade (e.g. `git pull`)
* `upgradeCmd` (required) - The `ConfigCommand` responsible for building, installing, or downloading the upgraded executable for the systemd service.
* `postUpgradeCmds` (optional) - A sequence of `ConfigCommand`s that need to be executed after the upgrade is complete but _before_ the updater restarts the systemd service

### The `ConfigCommand` Structure

The `preUpgradeCmds`, `upgradeCmd`, and `postUpgradeCmds` all take a singular mapping or a sequence of mappings representing `ConfigCommand`s. `ConfigCommand` is a structure that has the following mappings:

* `command` (required) - The command to be executed by the updater. It should only contain the name of the command and [no arguments](https://pkg.go.dev/os/exec@go1.19#Command). The command must exist in the current `PATH`.
* `args` (optional) - A sequence of arguments to the command. Given the format `$command $arg1 $flag1 $arg2 $flag2...`, all arguments, flags, and flag-bound arguments should exist as separate entries.

### Caveats

Go's [exec package](https://pkg.go.dev/os/exec@go1.19#pkg-overview) doesn't allow for the use of common built-in shell functions such as `cd` (from my experimentation, even `help` isn't an option). This creates problems when updating python-based systemd services that were set up using `virtualenv`. In order to support upgrades of these services, the upgrade command `ConfigCommand` struct must:

1. have a `command` property taking the form `bash -c "source $virtualenvPath/activate && $yourUpgradeCommand $yourUpgradeCommandArgs"`
2. have no `args`

### Example
An example of valid configuration file is located in the [test config directory](https://github.com/kwhite17/matrix-systemd-updater/blob/main/test_configs/good_config.yaml).

