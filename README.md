# RabbitRelayGo
A Go based utility for consuming messages from a specified RabbitMQ queue (master)
and relaying those messages, unchanged, to configured remote RabbitMQ servers (slaves)

## Build Requriements
 * A Go Development Environment (see: https://golang.org/doc/install)

## Prerequisites:

 * go-flags - A command line options parsing library
  * http://godoc.org/github.com/jessevdk/go-flags
 
 * amqp - A go library for accessing amqp servers
  * http://godoc.org/github.com/streadway/amqp
 
 * json-pointer - I library for parsing arbitrary JSON data
  * http://github.com/dustin/go-jsonpointer

 * spew - a variable dumper
  * http://godoc.org/github.com/davecgh/go-spew/spew

 * seelog - a Go logging framework
  * https://github.com/cihub/seelog/wiki

 * profile - A Go Profiling Helper Library
  * https://github.com/davecheney/profile

### Installing Prerequisite Libraries:
```
	go get github.com/jessevdk/go-flags
	go get github.com/streadway/amqp
	go get github.com/davecgh/go-spew/spew
	go get github.com/dustin/go-jsonpointer
  go get github.com/cihub/seelog
  go get github.com/davecheney/profile

```

## Local Prerequisites:
 * rmqpublisher - A convienent wrapper for publishing messages to RabbitMQ
 * rmqconsumer - A convienent wrapper for consuming messages from RabbitMQ

### Installing local Prerequisite Libraries:
Due to the fact that 'go get' doesn't really understand gitlab based repositories 
that don't, in fact, begin with "gitlab." as part of the address, we have to 
download these libraries and install them locally:

```
    git clone https://git.synapse-wireless.com/hcap/rmqpublisher.git
    git clone https://git.synapse-wireless.com/hcap/rmqconsumer.git
    cd rmqpublisher
    go install
    cd ..
    cd rmqconsumer
    go install
    
```
 
## Project Structure
The project is built from various files, all part of the package "main":

 * rabbitrelay.go       - Main file, contains main() function.
 * publisher.go         - Publish to a RabbitMQ exchange.
 * cmdline.go           - Command line parsing
 * config.go            - Config file parsing
 * types.go             - Various data structures
 * consumer.go          - RabbitMQ Consumer - most of the work is done here

## Build Instructions
 * To build the program change to the package directory, then:

```
go build

```

## Install instructions
 * To install the program, change to the package directory, then:

```

go install

``` 
 * This will install the program to your $GOPATH\bin directory.

## Profiling
 * Run the program with the '-p' option to enable profile data output.
 * Profile data files (*.pprof) are written to the current directory.
 * Profiling should not be enabled by default - it will affect performance.
 * To create a useful (pdf) graph from a profile file:

```
    go tool pprof -pdf ./rabbitrelaygo ./cpu.pprof > pprof.CPU.pdf
```


## Logging
 * Logging is handled via the seelog framework and configured by ./seelog.xml
 * Set the 'minlevel' to your minimum logging level.  One of:
  * "trace"
  * "debug"
  * "info"
  * "warn"
  * "error"
  * "critical"


Sample Logging Configuration file

```

<seelog minlevel="trace" maxlevel="critical">
    <outputs>
        <rollingfile type="size" filename="./rabbitrelay.log" maxsize="1000000" maxrolls="50" />
    </outputs>
</seelog>

```
## Configuration
 * By default a configuration file is read from the local directory (rabbitrelaygo.cfg)
 * Configuration file contains the connection information for one Master RabbitMQ server and one or more slave servers
 * Configuration file MUST be valid JSON

 * Notable Configuration Options for SSL
  * certCommonName - The CN Name from your certificate (if your host name doesn't match)
  * SSLSkipVerify  - Often during testing you don't have access to a full CA Chain, this forces the SSL Connection 
   to skip some of that verification.
  * PathToCACert - Path (from the program's PWD) to the CA Cert File.
  * PathToClientCert - Path (from the program's PWD) to the client's Cert File.
  * PathToClientKey  - Path (from the program's PWD) to the client's Key File.

Sample configuration file.

```
{
    "masterRabbitServer": {
        "displayName": "RabbitFed",
        "host": "10.82.15.89",
        "hostPort": 5672,
        "userName": "guest1",
        "password": "guest1",
        "queueName": "LocalServer1.Incoming",
        "exchange": "exchange.topic",
        "exchangeType": "fanout",
        "routingKey": "",
        "virtualHost": "/",
        "certCommonName": "",
        "SSLSkipVerify": false,
        "PathToCACert": "",
        "PathToClientCert":"",
        "PathToClientKey":""
    },

    "slaveRabbitServers": [
        {
            "displayName": "Daisy",
            "host": "daisy.dev.snapcloud.net",
            "hostPort": 5673,
            "userName": "guest",
            "password": "guest",
            "queueName": "LocalServer1.Incoming",
            "exchange": "exchange.topic",
            "exchangeType": "fanout",
            "routingKey": "",
            "virtualHost": "/",
            "certCommonName": "daisy",
            "SSLSkipVerify": true,
            "PathToCACert": "./RabbitSSLCerts/rhel_client/cacert.pem",
            "PathToClientCert":"./RabbitSSLCerts/rhel_client/atomictest.cert.pem",
            "PathToClientKey":"./RabbitSSLCerts/rhel_client/atomictest.key.pem"
        },
        {
            "displayName": "REHLSSL",
            "host": "atomic.dev.snapcloud.net",
            "hostPort": 5673,
            "userName": "guest",
            "password": "guest",
            "queueName": "LocalServer1.Incoming",
            "exchange": "exchange.topic",
            "exchangeType": "fanout",
            "routingKey": "",
            "virtualHost": "/",
            "certCommonName": "atomictest",
            "SSLSkipVerify": true,
            "PathToCACert": "./RabbitSSLCerts/rhel_client/cacert.pem",
            "PathToClientCert":"./RabbitSSLCerts/rhel_client/atomictest.cert.pem",
            "PathToClientKey":"./RabbitSSLCerts/rhel_client/atomictest.key.pem"
        }

    ]
}

```

## Usage
```

rabbitrelaygo (rabbitrelaygo)
usage: RabbitRelay.py [-f configFile]

A utility program to relay messages from a RabbitMQ Server (master) to one or
more slave RabbitMQ Servers

optional arguments:
  -h, --help            show this help message and exit
  -c=CONFIGFILE, --configfile=CONFIGFILE
                        Valid path to alternate config file. Default filename
                        is "./rabbitrelaygo.cfg"
  -p, --profile         Enable Profiling 

```