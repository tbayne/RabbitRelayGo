package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	//"github.com/davecgh/go-spew/spew"
)

/*
====================================================================
	Command Line Parsing
====================================================================
*/

// Options ... Parsed command line options structure
type Options struct {
	ConfigFile   string `short:"c" long:"configFile" optional:"yes" description:"Alternate Configuration File" default:"./rabbitrelaygo.cfg" `
	DisplayUsage bool   `short:"u" long:"usage" optional:"yes" description:"Display detailed usage message"`
	Profile      bool   `short:"p" long:"profile" optional:"yes" description:"Name of the profile output file"`
}

// ParseCommandLineOptions ... Returns parsed command line options
func ParseCommandLineOptions() Options {
	var opts Options
	var parser = flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		displayUsage()
		os.Exit(1)
	}
	//spew.Dump(opts)
	return opts
}

func displayUsage() {
	fmt.Println("rabbitrelay (rabbitrelaygo)")
	fmt.Println("usage: rabbitrelay [-c configFile] [-u usage]")
	fmt.Println("")
	fmt.Println("A utility program to relay messages from a RabbitMQ Server (master) to one or")
	fmt.Println("more slave RabbitMQ Servers")
	fmt.Println("")
	fmt.Println("optional arguments:")
	fmt.Println("")
	fmt.Println("\t-h, --help            show this help message and exit.")
	fmt.Println("\t-p, --profile         enable profiling output")
	fmt.Println("\t-c=CONFIGFILE, --configfile=CONFIGFILE")
	fmt.Println("\t\tValid path to alternate config file. Default filename is ./rabbitrelaygo.cfg")
}
