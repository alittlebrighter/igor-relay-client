package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/alittlebrighter/igor/modules"

	"github.com/alittlebrighter/igor-relay-client"
	"github.com/alittlebrighter/igor-relay-client/security"
)

func main() {
	id := flag.String("id", "pi-0", "Identification to be used on the server.")
	host := flag.String("host", "localhost:12345", "The relay host to connect to.")
	flag.Parse()

	err := security.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Error generating key pair: %s", err.Error())
	}

	rClient, _ := relayClient.NewRelayClient(*id, *host, json.Marshal, json.Unmarshal)

	err = rClient.OpenSocket()
	if err != nil {
		log.Printf("Could not open websocket to relay server: %s", err.Error())
	}

	go func() {
		msgs, err := rClient.ReadMessages()
		if err != nil {
			fmt.Printf("Error receiving message: %s\n", err.Error())
		}

		for msg := range msgs {
			fmt.Printf("\nReceived message: %+v\n", msg)
		}
	}()

	for {
		msg := &modules.Request{Module: "garage-doors", Method: "trigger", Args: map[string]interface{}{"force": false}}

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Door: ")
		tmp, _ := reader.ReadString('\n')
		msg.Args["door"], _ = strconv.Atoi(strings.TrimSpace(tmp))
		fmt.Print("Force (y/n): ")
		tmp, _ = reader.ReadString('\n')
		if tmp[0] == 'y' {
			msg.Args["force"] = true
		}
		/*
			fmt.Print("TTL: ")
			uresponse, _ := reader.ReadString('\n')
			ttl, err := strconv.Atoi(strings.TrimSpace(uresponse))
			if err != nil {
				fmt.Println("Error converting string to int: " + err.Error())
				ttl = 0
			}
		*/
		env, err := rClient.NewEnvelope(msg.Controller, nil, msg)
		if err != nil {
			log.Println("Error building envelope: " + err.Error())
			continue
		}

		response, err := rClient.SendMessage(env)
		if err != nil {
			log.Println("Error sending message: " + err.Error())
			continue
		} else {
			log.Println("Response received: " + string(response))
		}
	}
}

type Message struct {
	Module  string
	Command interface{}
}
