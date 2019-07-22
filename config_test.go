package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func createTestConfigFile() (*os.File, error) {

	file, err := ioutil.TempFile(".", "config_test")
	return file, err
}

func populateTestConfigFile(file *os.File) {

	file.WriteString(testConfig)
}

func Test_readConfigFile_1(t *testing.T) {

	// Setup test config file
	file, _ := createTestConfigFile()
	// Populate the test.config file
	populateTestConfigFile(file)
	file.Close()
	// defer it`s deletion
	defer os.Remove(file.Name())

	// Read the test configuration file
	config, _ := readConfigFile(file.Name())
	if config.masterServer.Host != "10.82.17.175" {
		spew.Dump(config)
		msg := fmt.Sprintf("Master Server Host String Incorrect.  \n\tExpected: 10.82.17.175\n\tGot: %s", config.masterServer.Host)
		t.Error(msg)
	} else {
		t.Log("Master Server Host Correct")
	}
}

func Test_readConfigFile_2(t *testing.T) {

	// Setup test config file
	file, _ := createTestConfigFile()
	// Populate the test.config file
	populateTestConfigFile(file)
	file.Close()
	// defer it`s deletion
	defer os.Remove(file.Name())

	// Read the test configuration file
	config, _ := readConfigFile(file.Name())
	if len(config.slaveServers) != 2 {
		t.Error("Wrong number of slave servers configured.")
	} else {
		t.Log("Correct number of slave servers read.")
	}
}

var testConfig = ` {
    "masterRabbitServer": {
        "displayName": "Test Master",
        "host": "10.82.17.175",
        "hostPort": 5672,
        "userName": "guest1",
        "password": "guest1",
        "queueName": "LocalServer1.Incoming",
        "exchange": "exchange.topic",
        "exchangeType": "fanout",
        "routingKey": "",
        "virtualHost": "/",
        "SSLSkipVerify": false,
        "pathToCACert": "",
        "pathToClientCert": "",
        "pathToClientKey": ""
    },
    "slaveRabbitServers": [
        {
            "displayName": "Test Slave 1",
            "hostIP": "10.82.15.52",
            "hostPort": 5672,
            "userName": "guest1",
            "password": "guest1",
            "queueName": "LocalServer1.Incoming",
            "exchange": "exchange.topic",
            "exchangeType": "fanout",
            "routingKey": "",
            "virtualHost": "/",
            "certCommonName": "testslave1",
            "SSLSkipVerify": false,
            "pathToCACert": "",
            "pathToClientCert": "",
            "pathToClientKey": ""
        },
        {
            "displayName": "Test Slave 2",
            "hostIP": "10.82.5.151",
            "hostPort": 5672,
            "userName": "guest1",
            "password": "guest1",
            "queueName": "LocalServer1.Incoming",
            "exchange": "exchange.topic",
            "exchangeType": "fanout",
            "routingKey": "",
            "virtualHost": "/",
            "certCommonName": "testslave2",
            "SSLSkipVerify": false,
            "pathToCACert": "",
            "pathToClientCert": "",
            "pathToClientKey": ""
        }
    ]
}`
