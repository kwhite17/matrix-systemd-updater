package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/kwhite17/matrix-systemd-updater/homeserver"
)

type MartixComponent int64

const (
	SYNAPSE MartixComponent = iota
	WHATSAPP
	KAKAO
)

var (
	matrixComponentMap = map[string]MartixComponent{
		"SYNAPSE":  SYNAPSE,
		"WHATSAPP": WHATSAPP,
		"KAKAO":    KAKAO,
	}
)

func main() {
	var componentToUpdate MartixComponent
	var parseOk bool

	printHelp := flag.Bool("help", false, "Print the help message for the Matrix Systemd Updater")
	flag.Func("component", "Which Matrix component To Update", func(arg string) error {
		componentToUpdate, parseOk = parseMatrixComponent(arg)
		if !parseOk {
			return fmt.Errorf("unknown matrix component type for arg: %s", arg)
		}
		return nil
	})

	serviceName := flag.String("serviceName", "", "Name of the systemd service for the Matrix component to update")
	packageName := flag.String("packageName", "", "Name of the debian package for the Matrix component to update")
	flag.Parse()

	if *printHelp {
		printHelpMessage()
		os.Exit(0)
	}

	log.Printf("Executing Update for Component: %v\n", componentToUpdate.String())
	switch componentToUpdate {
	case SYNAPSE:
		homeserver.ExecuteHomeServerCron(*serviceName, *packageName)
	case WHATSAPP:
		fallthrough
	case KAKAO:
		fallthrough
	default:
		log.Fatalf("Unsure of how to update Matrix component: %v\n", componentToUpdate)
	}
}

func parseMatrixComponent(arg string) (MartixComponent, bool) {
	component, ok := matrixComponentMap[strings.ToUpper(arg)]
	return component, ok
}

func (mc MartixComponent) String() string {
	switch mc {
	case SYNAPSE:
		return "SYNAPSE"
	case WHATSAPP:
		return "WHATSAPP"
	case KAKAO:
		return "KAKAO"
	}
	return "UNKNOWN"
}

func printHelpMessage() {
	fmt.Println("This is a CLI for updating Matrix components registered as systemd services. It assumes apt is installed on server.")
	fmt.Printf("Command: matrix-systemd-updater -component [COMPONENT NAME] -serviceName [SERVICE NAME] -packageName [PACKAGE NAME]\n\n")
	fmt.Println("Options:")
	fmt.Println("-component: Required. The name of the component to update. Options currently are SYNAPSE, WHATSAPP, and KAKAO")
	fmt.Println("-serviceName: Optional. The systemd service to restart. If not specified, the component updater will print the default service being used")
	fmt.Println("-packageName: Optional. The apt package to update. If not specified, the component updater will print the default package being used")
}
