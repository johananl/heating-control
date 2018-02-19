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

	stop, wg := c.Run()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)

	<-shutdown

	log.Println("Shutting down")
	stop <- true
	wg.Wait()
	log.Println("Shutdown complete")
}
