# Example OpenChirp Service

## Overview
This is a minimal OpenChirp service that tracks the number of publications
to the `rawrx` and `rawtx` topics and publishes the count to the
`rawrxcount` and `rawtxcount` transducer topics.

We demonstrate argument/environment variable parsing,
setting up the service client, and handling device transducer data.

The core of this service revolves around the 2 Device interface
functions `ProcessLink` and `ProcessMessage`. In these handlers, you are
responsible for subscribing to subtopics that your are interested in
and then handling messages that are received from them.
```go
type Device struct {
	rawRxCount int
	rawTxCount int
}

// ProcessLink is called once, during the initial setup of a
// device, and is provided the service config for the linking device.
func (d *Device) ProcessLink(ctrl *framework.DeviceControl) string {
	// ctrl.Id() - The device id
	// ctrl.Config() - The device's current config for your service
	// ctrl.Publish() - Directly publish a message to a subtopic
	// ctrl.Subscribe() - Subscribe to a device subtopic
	// ctrl.Unsubscribe() - Unsubscribe from a device subtopic
	fmt.Println("I got a new device config!:", ctrl.Config())

	// Subscribe to subtopic "transducer/rawrx"
	ctrl.Subscribe(framework.TransducerPrefix+"/rawrx", rawRxKey)
	// Subscribe to subtopic "transducer/rawtx"
	ctrl.Subscribe(framework.TransducerPrefix+"/rawtx", rawTxKey)

	// This message is sent to the service status for the linking device
	return "Success"
}

// ProcessMessage is called upon receiving a pubsub message destined for
// this device.
func (d *Device) ProcessMessage(ctrl *framework.DeviceControl, msg framework.Message) {
	// msg.Topic() - The device's subtopic the message was received on
	// msg.Key() - The key that was given to Subscribe
	// msg.Payload() - The received message payload
	fmt.Println("Got the message", string(msg.Payload()), "from", ctrl.Id(), "!")

	if msg.Key().(int) == rawRxKey {
		d.rawRxCount++
		subtopic := framework.TransducerPrefix + "/rawrxcount"
		ctrl.Publish(subtopic, fmt.Sprint(d.rawRxCount))
	} else if msg.Key().(int) == rawTxKey {
		d.rawTxCount++
		subtopic := framework.TransducerPrefix + "/rawtxcount"
		ctrl.Publish(subtopic, fmt.Sprint(d.rawTxCount))
	} else {
		logitem.Errorln("Received unassociated message")
	}
}
```
For more advanced scenarios, you have the flexibility to implement
`ProcessConfigChange` and/or `ProcessUnlink`.

For a more detailed explanation of when these methods are called,
please see https://godoc.org/github.com/OpenChirp/framework#Device .

To dig right into the code, checkout the only source file, [main.go](main.go).

## Concept

The overarching idea is that the framework instantiates your Device object for each new device that has been linked to your service. All your runtime data pertaining to that device can be saved inside your Device object.
```go
// Device holds any data you want to keep around for a specific
// device that has linked your service.
//
// In this example, we will keep track of the rawrx and rawtx message counts
type Device struct {
	rawRxCount int
	rawTxCount int
}
```
The framework brings your Device object through all phases of it's life cycles by invoking your Device object's `ProcessLink`, `ProcessMessage`, `ProcessConfigChange`, and `ProcessUnlink` methods.

Within these handlers, the framework provides a per device context, which
abstracts away additional server requests and pointless glue logic, such as determining a device's pubsub prefix or having implementing a synchronous message to device action router.
Using the services framework, you only need to spend time implementing the logic that makes your service unique.

To give you a better feel for what the framework is doing, you can think of
it like your personal assistant.
It strips away redundant information from parameters, supplies you with independently tracked differences to service configuration, infers pubsub prefixes from device context, pushes statuses to devices, handles config changes when you don't want to, cleans up your subscriptions when the device is no longer needed, pushes a service status if your service crashes, and handles runtime errors.

This example uses the [OpenChirp Golang framework library](https://github.com/OpenChirp/framework).
