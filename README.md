# Example OpenChirp Service
This is a minimal OpenChirp service that shows the main service device subscription event runtime loop.
This exampe uses the [OpenChirp Golang framework library](https://github.com/OpenChirp/framework).

The main runtime of a service revolves around processing device events, such as when a device links, unlinks, or updates their config for your service. This example service sets up argument parsing, program termination, and the main runtime loop where you will process device events.
The idea is to process a device event by handing it off to some asynchronous task, publishing the status/result of handling the event, and repeat.

Please see [main.go](main.go) to understand more about how services operate.
