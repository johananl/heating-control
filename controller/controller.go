package controller

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/eclipse/paho.mqtt.golang"
)

// Reading represents a temperator reading sent to the controller from a sensor.
type Reading struct {
	SensorID    string  `json:"sensorID"`
	ReadingType string  `json:"type"`
	Value       float64 `json:"value"`
}

// ValveOpenness represents a message sent to an actuator by the controller.
type ValveOpenness struct {
	Level int `json:"level"`
}

// Controller represents a heating controller.
type Controller struct {
	brokerURI      string
	readingsTopic  string
	actuatorsTopic string
	targetTemp     float64
	client         mqtt.Client
}

// ProcessReading receives a Reading and executes an appropriate action, if any, based on it.
// This function contains the controller's "policy" or logic, and is therefore one of the main
// candidates for optimization.
func (c *Controller) ProcessReading(r Reading) {
	log.Printf("Received reading: sensor %v temp %v", r.SensorID, r.Value)

	// Set valve openness
	if r.Value < c.targetTemp {
		log.Println("Too cold! Opening valve.")
		c.SetValveOpenness(100)
	} else if r.Value > c.targetTemp {
		log.Println("Too hot! Closing valve.")
		c.SetValveOpenness(0)
	} else {
		log.Println("Target temperature achieved!")
		return
	}
}

// SetValveOpenness sets the openness of a valve by publishing to an actuators topic.
func (c *Controller) SetValveOpenness(v int) {
	vo := ValveOpenness{Level: v}
	json, err := json.Marshal(vo)
	if err != nil {
		log.Println(err.Error())
		return
	}

	log.Printf("Setting valve openness to %d", vo.Level)
	token := c.client.Publish(c.actuatorsTopic, 0, false, json)
	token.Wait()
}

// Initialize the MQTT client, connect to the broker and subscribe to the readings topic.
func (c *Controller) start() (mqtt.Client, error) {
	// Set MQTT client options
	opts := mqtt.NewClientOptions()
	opts.AddBroker(c.brokerURI)

	// Connect to MQTT broker
	client := mqtt.NewClient(opts)
	log.Println("Connecting to MQTT broker")
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return client, token.Error()
	}
	c.client = client
	log.Println("Connected to MQTT broker")

	// Subscribe to readings topic
	var handler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		go func() {
			r := Reading{}
			err := json.Unmarshal(msg.Payload(), &r)
			if err != nil {
				log.Println("Error parsing JSON message: ", err)
				return
			}

			c.ProcessReading(r)
		}()
	}

	log.Println("Subscribing to readings topic")
	if token := client.Subscribe(c.readingsTopic, 0, handler); token.Wait() && token.Error() != nil {
		return client, token.Error()
	}

	return client, nil
}

// Unsubscribe from the readings topic and disconnect the MQTT client.
func (c *Controller) stop(client mqtt.Client) error {
	log.Println("Unsubscribing from readings topic")
	if token := client.Unsubscribe(c.readingsTopic); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	log.Println("Disconnecting from MQTT broker")
	client.Disconnect(1000)

	return nil
}

// Run starts the controller goroutine. It returns a quit channel, an error channel and a waitgroup
// for graceful shutdown.
func (c *Controller) Run() (chan<- bool, <-chan error, *sync.WaitGroup) {
	stop := make(chan bool)
	errChan := make(chan error)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		log.Println("Controller started")
		defer wg.Done()

		client, err := c.start()
		if err != nil {
			errChan <- err
			return
		}
		log.Println("Controller ready")

		// Wait for stop signal
		<-stop
		log.Println("Stopping controller")
		err = c.stop(client)
		if err != nil {
			errChan <- err
		}
		log.Println("Controller stopped")
	}()

	return stop, errChan, &wg
}

// NewController creates a new controller and returns a pointer to it.
func NewController(brokerURI, readingsTopic, actuatorsTopic string, targetTemp float64) *Controller {
	return &Controller{
		brokerURI:      brokerURI,
		readingsTopic:  readingsTopic,
		actuatorsTopic: actuatorsTopic,
		targetTemp:     targetTemp,
	}
}
