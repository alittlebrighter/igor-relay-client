package main

import (
	"flag"
	"log"

	"github.com/alittlebrighter/treehouse-relay-client"
)

func main() {
	id := flag.String("id", "pi-0", "Identification to be used on the server.")
	host := flag.String("host", "localhost:12345", "The relay host to connect to.")
	key := flag.String("key", "AES256Key-32Characters1234567890", "The symmetric key to use.")
	flag.Parse()

	rClient := relayClient.NewRelayClient(*id, *host, []byte(*key))

	relayChan, err := rClient.OpenSocket()
	if err != nil {
		log.Fatalf("Could not open websocket to relay server: %s", err.Error())
	}

	for {
		msg := <-relayChan
		if string(msg) == "EXIT" {
			break
		}
		log.Printf("Received command: %s", string(msg))
	}
}
