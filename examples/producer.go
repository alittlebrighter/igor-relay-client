package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/alittlebrighter/igor-relay-client"
	"github.com/alittlebrighter/igor-relay-client/security"
)

func main() {
	id := flag.String("id", "pi-0", "Identification to be used on the server.")
	host := flag.String("host", "localhost:12345", "The relay host to connect to.")
	//key := flag.String("key", "AES256Key-32Characters1234567890", "The symmetric key to use.")
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
		msg := new(Message)

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Destination Controller: ")
		msg.Controller, _ = reader.ReadString('\n')
		msg.Controller = strings.TrimSpace(msg.Controller)
		fmt.Print("Module to control: ")
		msg.Instruction.Module, _ = reader.ReadString('\n')
		msg.Instruction.Module = strings.TrimSpace(msg.Instruction.Module)
		fmt.Print("Set to mode: ")
		msg.Instruction.Mode, _ = reader.ReadString('\n')
		msg.Instruction.Mode = strings.TrimSpace(msg.Instruction.Mode)
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
	Controller  string
	Instruction struct {
		Module string
		Mode   string
	}
}
