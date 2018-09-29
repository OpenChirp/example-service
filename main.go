// Craig Hesling
// November 26, 2017
//
// This is an example OpenChirp service that tracks the number of publications
// to the rawrx and rawtx topics and publishes the count to the
// rawrxcount and rawtxcount transducer topics.
// This example demonstates argument/environment variable parsing,
// setting up the service client, and handling device transducer data.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/openchirp/framework"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	version string = "1.0"
)

const (
	// Set this value to true to have the service publish a service status of
	// "Running" each time it receives a device update event
	//
	// This could be used as a service alive pulse if enabled
	// Otherwise, the service status will indicate "Started" at the time the
	// service "Started" the client
	runningStatus = true
)

const (
	// The subscription key used to identify a messages types
	rawRxKey = 0
	rawTxKey = 1
)

// Device holds any data you want to keep around for a specific
// device that has linked your service.
//
// In this example, we will keep track of the rawrx and rawtx message counts
type Device struct {
	rawRxCount int
	rawTxCount int
}

// NewDevice is called by the framework when a new device has been linked.
func NewDevice() framework.Device {
	d := new(Device)
	// The following initialization is redundant in Go
	d.rawRxCount = 0
	d.rawTxCount = 0
	// Change type to the Device interface
	return framework.Device(d)
}

// ProcessLink is called once, during the initial setup of a
// device, and is provided the service config for the linking device.
func (d *Device) ProcessLink(ctrl *framework.DeviceControl) string {
	// This simply sets up console logging for our program.
	// Any time this logitem is use to print messages,
	// the key/value string "deviceid=<device_id>" is prepended to the line.
	logitem := log.WithField("deviceid", ctrl.Id())
	logitem.Debug("Linking with config:", ctrl.Config())

	// Subscribe to subtopic "rawrx"
	ctrl.Subscribe("rawrx", rawRxKey)
	// Subscribe to subtopic "rawtx"
	ctrl.Subscribe("rawtx", rawTxKey)

	logitem.Debug("Finished Linking")

	// This message is sent to the service status for the linking device
	return "Success"
}

// ProcessUnlink is called once, when the service has been unlinked from
// the device.
func (d *Device) ProcessUnlink(ctrl *framework.DeviceControl) {
	logitem := log.WithField("deviceid", ctrl.Id())
	logitem.Debug("Unlinked:")

	// The framework already handles unsubscribing from all
	// Device associted subtopics, so we don't need to call
	// ctrl.Unsubscribe.
}

// ProcessConfigChange is intended to handle a service config updates.
// If your program does not need to handle incremental config changes,
// simply return false, to indicate the config update was unhandled.
// The framework will then automatically issue a ProcessUnlink and then a
// ProcessLink, instead. Note, NewDevice is not called.
//
// For more information about this or other Device interface functions,
// please see https://godoc.org/github.com/OpenChirp/framework#Device .
func (d *Device) ProcessConfigChange(ctrl *framework.DeviceControl, cchanges, coriginal map[string]string) (string, bool) {
	logitem := log.WithField("deviceid", ctrl.Id())

	logitem.Debug("Ignoring Config Change:", cchanges)
	return "", false

	// If we have processed this config change, we should return the
	// new service status message and true.
	//
	//logitem.Debug("Processing Config Change:", cchanges)
	//return "Sucessfully updated", true
}

// ProcessMessage is called upon receiving a pubsub message destined for
// this device.
// Along with the standard DeviceControl object, the handler is provided
// a Message object, which contains the received message's payload,
// subtopic, and the provided Subscribe key.
func (d *Device) ProcessMessage(ctrl *framework.DeviceControl, msg framework.Message) {
	logitem := log.WithField("deviceid", ctrl.Id())
	logitem.Debugf("Processing Message: %v: [ % #x ]", msg.Key(), msg.Payload())

	if msg.Key().(int) == rawRxKey {
		d.rawRxCount++
		subtopic := "rawrxcount"
		ctrl.Publish(subtopic, fmt.Sprint(d.rawRxCount))
	} else if msg.Key().(int) == rawTxKey {
		d.rawTxCount++
		subtopic := "rawtxcount"
		ctrl.Publish(subtopic, fmt.Sprint(d.rawTxCount))
	} else {
		logitem.Errorln("Received unassociated message")
	}
}

// run is the main function that gets called once form main()
func run(ctx *cli.Context) error {
	/* Set logging level (verbosity) */
	log.SetLevel(log.Level(uint32(ctx.Int("log-level"))))

	log.Info("Starting Example Service")

	/* Start framework service client */
	c, err := framework.StartServiceClientManaged(
		ctx.String("framework-server"),
		ctx.String("mqtt-server"),
		ctx.String("service-id"),
		ctx.String("service-token"),
		"Unexpected disconnect!",
		NewDevice)
	if err != nil {
		log.Error("Failed to StartServiceClient: ", err)
		return cli.NewExitError(nil, 1)
	}
	defer c.StopClient()
	log.Info("Started service")

	/* Post service's global status */
	if err := c.SetStatus("Starting"); err != nil {
		log.Error("Failed to publish service status: ", err)
		return cli.NewExitError(nil, 1)
	}
	log.Info("Published Service Status")

	/* Setup signal channel */
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	/* Post service status indicating I started */
	if err := c.SetStatus("Started"); err != nil {
		log.Error("Failed to publish service status: ", err)
		return cli.NewExitError(nil, 1)
	}
	log.Info("Published Service Status")

	/* Wait on a signal */
	sig := <-signals
	log.Info("Received signal ", sig)
	log.Warning("Shutting down")

	/* Post service's global status */
	if err := c.SetStatus("Shutting down"); err != nil {
		log.Error("Failed to publish service status: ", err)
	}
	log.Info("Published service status")

	return nil
}

func main() {
	/* Parse arguments and environmental variable */
	app := cli.NewApp()
	app.Name = "example-service"
	app.Usage = ""
	app.Copyright = "See https://github.com/openchirp/example-service for copyright information"
	app.Version = version
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "framework-server",
			Usage:  "OpenChirp framework server's URI",
			Value:  "http://localhost:7000",
			EnvVar: "FRAMEWORK_SERVER",
		},
		cli.StringFlag{
			Name:   "mqtt-server",
			Usage:  "MQTT server's URI (e.g. scheme://host:port where scheme is tcp or tls)",
			Value:  "tcp://localhost:1883",
			EnvVar: "MQTT_SERVER",
		},
		cli.StringFlag{
			Name:   "service-id",
			Usage:  "OpenChirp service id",
			EnvVar: "SERVICE_ID",
		},
		cli.StringFlag{
			Name:   "service-token",
			Usage:  "OpenChirp service token",
			EnvVar: "SERVICE_TOKEN",
		},
		cli.IntFlag{
			Name:   "log-level",
			Value:  4,
			Usage:  "debug=5, info=4, warning=3, error=2, fatal=1, panic=0",
			EnvVar: "LOG_LEVEL",
		},
	}

	/* Launch the application */
	app.Run(os.Args)
}
