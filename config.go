package main

import (
	"fmt"
	"io/ioutil"

	log "github.com/cihub/seelog"
	"github.com/dustin/go-jsonpointer"
)

// Server - Structure containing server connection parameters
type Server struct {
	DisplayName      string
	CertCommonName   string
	Host             string
	HostPort         int
	UserName         string
	Password         string
	QueueName        string
	Exchange         string
	ExchangeType     string
	RoutingKey       string
	VirtualHost      string
	SSLSkipVerify    bool
	PathToCACert     string
	PathToClientCert string
	PathToClientKey  string
}

// Config - Server configuration, master and one or more slave servers
type Config struct {
	masterServer Server
	slaveServers []Server
}

// Notice that this function returns TWO values -
// this is a very common idiom in Go.

func readConfigFile(cfgFile string) (Config, error) {

	log.Info("RabbitRelay: reading configuration from: ", cfgFile)

	config := Config{}

	var err error
	fileContents, err := ioutil.ReadFile(cfgFile)

	if err != nil {

		log.Critical("Error reading config file: ", err)

	} else {

		// This doesn't work, only the top level items are populated
		// err = jsonpointer.FindDecode(fileContents,"",&config)
		err = jsonpointer.FindDecode(fileContents, "/masterRabbitServer", &config.masterServer)
		if err != nil {
			fmt.Println("Error reading master server config: ", err)
			return config, err
		}
		err = jsonpointer.FindDecode(fileContents, "/slaveRabbitServers", &config.slaveServers)
		if err != nil {
			fmt.Println("Error reading slave server config: ", err)
			return config, err
		}
	}
	return config, err
}
