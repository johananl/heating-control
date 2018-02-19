package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/johananl/heating-control/controller"
)

var (
	brokerURI     string
	readingsTopic string
)

func main() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)

	flag.StringVar(&brokerURI, "brokeruri", "tcp://localhost:1883", "URI of MQTT broker. Example: tcp://mybroker:1883")
	flag.StringVar(&readingsTopic, "readingstopic", "/readings/temperature", "MQTT topic to subscribe to for readings.")
	flag.Parse()

	c := controller.NewController(brokerURI, readingsTopic)

	stop, err, wg := c.Run()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)

	// Exit on error or shutdown signal
	select {
	case e := <-err:
		log.Println("Got error from controller: ", e.Error())
	case <-shutdown:
		log.Println("Got shutdown signal")
		stop <- true
		wg.Wait()
	}

	log.Println("Shutdown complete")
}
