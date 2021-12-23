package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace"
	"solace.dev/go/messaging/pkg/solace/config"
	"solace.dev/go/messaging/pkg/solace/message"
	"solace.dev/go/messaging/pkg/solace/resource"
)

// Message Handler
func MessageHandler(message message.InboundMessage) {
	fmt.Printf("Message Dump %s \n", message)
}

func ReconnectionHandler(e solace.ServiceEvent) {
	e.GetTimestamp()
	e.GetBrokerURI()
	err := e.GetCause()
	if err != nil {
		fmt.Println(err)
	}
}

func main() {

	// Define Topic Subscriptions
	TOPIC_PREFIX := "solace/samples/go"

	// Configuration parameters
	brokerConfig := config.ServicePropertyMap{
		config.TransportLayerPropertyHost:                "tcp://localhost:55554",
		config.ServicePropertyVPNName:                    "default",
		config.AuthenticationPropertySchemeBasicPassword: "default",
		config.AuthenticationPropertySchemeBasicUserName: "default",
	}

	// Build A messaging service with a reconnection strategy of 20 retries over an interval of 3 seconds
	// Note: The reconnections strategy could also be configured using the broker properties object
	messagingService, err := messaging.NewMessagingServiceBuilder().FromConfigurationProvider(brokerConfig).Build()

	if err != nil {
		panic(err)
	}

	messagingService.AddReconnectionListener(ReconnectionHandler)

	// Connect to the messaging serice
	if err := messagingService.Connect(); err != nil {
		panic(err)
	}

	fmt.Println("Connected to the broker? ", messagingService.IsConnected())

	// Define Topic Subscriptions

	// topics := [...]string{TOPIC_PREFIX + "/>"}
	topics := [...]string{TOPIC_PREFIX + "/direct/sub/>", TOPIC_PREFIX + "/direct/sub/*", "solace/samples/>"}
	topics_sup := make([]resource.Subscription, len(topics))

	// Create topic objects
	for i, topicString := range topics {
		topics_sup[i] = resource.TopicSubscriptionOf(topicString)
	}

	// Print out list of strings to subscribe to
	for _, ts := range topics_sup {
		fmt.Println("Subscribed to: ", ts.GetName())
	}

	// Build a Direct message receivers with given topics
	directReceiver, err := messagingService.CreateDirectMessageReceiverBuilder().
		WithSubscriptions(topics_sup...).
		Build()

	if err != nil {
		panic(err)
	}

	// Start Direct Message Receiver
	if err := directReceiver.Start(); err != nil {
		panic(err)
	}

	fmt.Println("Direct Receiver running? ", directReceiver.IsRunning())

	// Register Message callback handler to the Message Receiver
	if regErr := directReceiver.ReceiveAsync(MessageHandler); regErr != nil {
		panic(regErr)
	}

	fmt.Println("\n===Interrupt (CTR+C) to handle graceful terminaltion of the subscriber===\n")

	// Run forever until an interrupt signal is received
	// Handle interrupts

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until a signal is received.
	<-c

	// Terminate the Direct Receiver
	directReceiver.Terminate(1 * time.Second)
	fmt.Println("\nDirect Receiver Terminated? ", directReceiver.IsTerminated())
	// Disconnect the Message Service
	messagingService.Disconnect()
	fmt.Println("Messaging Service Disconnected? ", !messagingService.IsConnected())

}
