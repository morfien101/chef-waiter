package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kardianos/service"
	"github.com/morfien101/chef-waiter/logs"
)

// VERSION holds the version of the program
// Don't change this as the build server tags the builds.
var VERSION = "1.0.0"

// Flags for the application launch
var (
	versionCheck = flag.Bool("v", false, "Outputs the version of the program.")
	helpFlag     = flag.Bool("h", false, "Shows the help menu")
	svcFlag      = flag.String("service", "", "Control the system service.")
	logger       logs.SysLogger
)

func main() {
	// Deal with flags
	digestFlags()

	svcConfig := &service.Config{
		Name:        "chefwaiter",
		DisplayName: "Chef Waiter",
		Description: "Chef Waiter: API that allows you to control the chef client remotely.",
	}

	prg := &program{}

	serviceController, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Create a channel to house the errors from the service.
	// These are then printed out on the go func below.
	errsChan := make(chan error, 5)
	logger, err = serviceController.Logger(errsChan)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			err := <-errsChan
			if err != nil {
				log.Print(err)
			}
		}
	}()

	// Look for the --service flag and do what we need to here.
	if len(*svcFlag) != 0 {
		err := service.Control(serviceController, *svcFlag)
		if err != nil {
			log.Printf("Valid actions: %q\n", service.ControlAction)
			log.Fatal(err)
		}
		return
	}
	err = serviceController.Run()
	if err != nil {
		logger.Error(err)
	}
}

func digestFlags() {
	// Parse the flags to get the state we need to run in.
	flag.Parse()

	if *versionCheck {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	if *helpFlag {
		flag.PrintDefaults()
		os.Exit(0)
	}
}
