package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"strconv"

	log "github.com/cihub/seelog"
	"github.com/davecheney/profile"
	"github.com/streadway/amqp"
	rc "synapse-wireless.com/rmqconsumer"
	rp "synapse-wireless.com/rmqpublisher"
)

var pList []*rp.Publisher
var consumer *rc.Consumer

func main() {

	// we want to run in parallel:
	runtime.GOMAXPROCS(4) // 1 for main thread, 2 for slave servers, 1 for master server

	// Set up logging first
	logger, err := log.LoggerFromConfigAsFile("./seelog.xml")
	if err != nil {
		panic(err)
	} else {
		// Make the current logger configuration the default
		log.ReplaceLogger(logger)
		defer log.Flush()
	}

	Opts := ParseCommandLineOptions()
	config, err := readConfigFile(Opts.ConfigFile)

	if Opts.Profile { // is the profiling switch passed to the program?
		profConfig := profile.Config{
			CPUProfile:   true,
			MemProfile:   true,
			BlockProfile: true,
			ProfilePath:  ".", // Output profile data to the local directory.
		}
		fmt.Println("Profiling is enabled.")
		defer profile.Start(&profConfig).Stop()
	}

	if err == nil {
		//fmt.Println("Setting up connection(s) to slave server(s)")
		pList = setupPublishers(&config.slaveServers)

		// Setup the connection to the master server
		//fmt.Println("Setting up connection to master server")
		consumer = consumeMessages(&config.masterServer, consumerHandler)

	} else {
		log.Error("Error reading/parsing configuration file: ", err)
		fmt.Println("Error reading/parsing configuration file: ", err)
	}

	log.Info("Consumer started, waiting for messages to publish.")
	fmt.Println("Waiting for messages to publish.  Press Ctrl-C to exit")

	// The anonymous function looks for a signal from Ctrl-C, then
	// calls shutdown on our consumer (which also shuts down the publishers)

	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		log.Trace("listening for Ctrl-C")
		for _ = range signalChan {
			log.Trace("Received Ctrl-C, exiting.")
			if consumer != nil {
				consumer.Shutdown()
				log.Flush()
			}
			for _, p := range pList {
				p.Quit <- true
			}
			cleanupDone <- true
		}
	}()
	// Main blocks waiting for input on this channel
	<-cleanupDone
}

func consumerHandler(deliveries <-chan amqp.Delivery, done chan error) {
	log.Trace("Started")
	for d := range deliveries {
		//log.Trace("Processing deliveries")
		d.Ack(false) // No matter what we ack the message.
		// Publish the messages to the slave server(s)
		for _, p := range pList {

			// Add the message to this publishers Queue
			p.InMsg <- d
			//log.Trace("Message Sent")
		}
	}
	log.Trace("Deliveries channel closed.")
	done <- nil
}

// Handler for messages to be published
func publisherHandler(msgq <-chan amqp.Delivery, quit <-chan bool, done chan<- error, p *rp.Publisher) {

	for {
		select {
		case msg := <-msgq:
			p.PublishMessage(msg)
		case <-quit:
			msg := fmt.Sprintf("Shutting down handler for publisher (%s)", p.URL)
			log.Trace(msg)
			done <- nil
			return
		}
	}
}

// Handler for messages to be consumed
func consumeMessages(ms *Server, ch rc.ConsumerHandler) *rc.Consumer {

	// Build the server URL
	url := buildServerURL(ms)
	log.Info("Server URL: ", url)
	fmt.Println("Setting up connection to master server ", ms.DisplayName, " @ ", ms.Host)

	var sslcfg *tls.Config
	var err error

	if len(ms.PathToClientCert) > 0 {
		sslcfg, err = getSSLConfig(*ms)
	}
	if err != nil {
		log.Error("SSL Related error creating consumer: ", err)
		fmt.Println("SSL Related error creating consumer: ", err)
	}

	// Create a new consumer of messages clientCA string, clientCert string, clientKey string
	result, err := rc.NewConsumer(url, ms.Exchange, ms.ExchangeType, ms.QueueName, "", "rabbitrelaygo",
		ch, sslcfg)
	if err != nil {
		log.Error("Error creating consumer: ", err)
		fmt.Println("Error creating consumer: ", err)
	}
	return result
}

func buildServerURL(s *Server) string {

	var result string
	result = "amqp://" + s.UserName + ":" + s.Password + "@" + s.Host + ":" + strconv.Itoa(s.HostPort) + "/"
	return result
}

func setupPublishers(ss *[]Server) []*rp.Publisher {
	fmt.Println("Starting slave server connections...")
	var result []*rp.Publisher
	for _, svr := range *ss {
		var p *rp.Publisher
		var err error
		fmt.Println("Connecting to slave server: ", svr.DisplayName, " at: ", svr.Host)
		log.Info("Connecting to slave server: ", svr.DisplayName, " at: ", svr.Host)

		url := buildServerURL(&svr)
		sslcfg, err := getSSLConfig(svr)
		if err != nil {
			msg := fmt.Sprintf("Error setting up slave server %s (%s): %s", svr.DisplayName, svr.Host, err)
			log.Error(msg)

		} else {

			p, err = rp.NewPublisher(url, svr.Exchange, svr.ExchangeType, svr.QueueName, svr.RoutingKey,
				"rabbitrelay", publisherHandler, sslcfg)
			if err != nil {
				msg := fmt.Sprintf("Error setting up slave server %s (%s): %s", svr.DisplayName, svr.Host, err)
				log.Error(msg)

			} else {
				log.Info("... Connected to slave server: ", svr.DisplayName)
				result = append(result, p)
			}
		}
	}
	return result
}

func getSSLConfig(svr Server) (*tls.Config, error) {

	var err error
	cfg := new(tls.Config)
	if len(svr.PathToClientCert) > 0 {
		if svr.SSLSkipVerify == true {
			cfg.InsecureSkipVerify = true
		}
		if len(svr.CertCommonName) > 0 {
			cfg.ServerName = svr.CertCommonName

		}
		//cfg.InsecureSkipVerify = true
		cfg.RootCAs = x509.NewCertPool()

		if ca, err := ioutil.ReadFile(svr.PathToCACert); err == nil {
			cfg.RootCAs.AppendCertsFromPEM(ca)
		}

		if cert, err := tls.LoadX509KeyPair(svr.PathToClientCert, svr.PathToClientKey); err == nil {
			cfg.Certificates = append(cfg.Certificates, cert)
		}
	}
	/*
		fmt.Println("============================================")
		spew.Dump(cfg)
		spew.Dump(cfg.RootCAs)
		spew.Dump(cfg.Certificates)
		spew.Dump(cfg.ServerName)
		spew.Dump(cfg.InsecureSkipVerify)
		fmt.Println("-------------------------------------------")
	*/
	return cfg, err
}

/*
 Go Rabbit SSL Example
--------------------------------
package main

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/streadway/amqp"
	"io/ioutil"
	"log"
)

func main() {
	// To get started with SSL/TLS follow the instructions for adding SSL/TLS
	// support in RabbitMQ with a private certificate authority here:
	//
	// http://www.rabbitmq.com/ssl.html
	//
	// Then in your rabbitmq.config, disable the plain AMQP port, verify clients
	// and fail if no certificate is presented with the following:
	//
	//   [
	//   {rabbit, [
	//     {tcpListeners, []},     % listens on 127.0.0.1:5672
	//     {ssl_listeners, [5671]}, % listens on 0.0.0.0:5671
	//     {ssl_options, [{cacertfile,"/path/to/your/testca/cacert.pem"},
	//                    {certfile,"/path/to/your/server/cert.pem"},
	//                    {keyfile,"/path/to/your/server/key.pem"},
	//                    {verify,verify_peer},
	//                    {fail_if_no_peer_cert,true}]}
	//     ]}
	//   ].

	cfg := new(tls.Config)

	// The self-signing certificate authority's certificate must be included in
	// the RootCAs to be trusted so that the server certificate can be verified.
	//
	// Alternatively to adding it to the tls.Config you can add the CA's cert to
	// your system's root CAs.  The tls package will use the system roots
	// specific to each support OS.  Under OS X, add (drag/drop) your cacert.pem
	// file to the 'Certificates' section of KeyChain.app to add and always
	// trust.
	//
	// Or with the command line add and trust the DER encoded certificate:
	//
	//   security add-certificate testca/cacert.cer
	//   security add-trusted-cert testca/cacert.cer
	//
	// If you depend on the system root CAs, then use nil for the RootCAs field
	// so the system roots will be loaded.

	cfg.RootCAs = x509.NewCertPool()

	if ca, err := ioutil.ReadFile("testca/cacert.pem"); err == nil {
		cfg.RootCAs.AppendCertsFromPEM(ca)
	}

	// Move the client cert and key to a location specific to your application
	// and load them here.

	if cert, err := tls.LoadX509KeyPair("client/cert.pem", "client/key.pem"); err == nil {
		cfg.Certificates = append(cfg.Certificates, cert)
	}

	// Server names are validated by the crypto/tls package, so the server
	// certificate must be made for the hostname in the URL.  Find the commonName
	// (CN) and make sure the hostname in the URL matches this common name.  Per
	// the RabbitMQ instructions for a self-signed cert, this defautls to the
	// current hostname.
	//
	//   openssl x509 -noout -in server/cert.pem -subject
	//
	// If your server name in your certificate is different than the host you are
	// connecting to, set the hostname used for verification in
	// ServerName field of the tls.Config struct.

	conn, err := amqp.DialTLS("amqps://server-name-from-certificate/", cfg)

	log.Printf("conn: %v, err: %v", conn, err)
}
*/
