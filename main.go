package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/johananl/heating-control/controller"
)

func main() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)

	c := controller.NewController()

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
