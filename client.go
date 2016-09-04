package relayClient

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"golang.org/x/net/websocket"

	"github.com/alittlebrighter/treehouse-relay-client/security"
)

type RelayClient struct {
	id         string
	relayHost  string
	socketConn chan []byte
	key        []byte
}

func NewRelayClient(id, relayHost string, key []byte) *RelayClient {
	return &RelayClient{id: id, relayHost: relayHost, key: key}
}

// OpenConn opens a websocket on host arg identifying itself with controllerID arg and
// returns a channel that relays messages coming down from the server
// it currently only receives commands at the moment, later we can use the channel and ws both ways
func (rc *RelayClient) OpenSocket() (relayChan chan []byte, err error) {
	// origin can be a bogus URL so we'll just use it to identify the connection on the server
	origin := "http://" + rc.id
	url := "ws://" + rc.relayHost + "/socket"

	ws, err := websocket.Dial(url, "", origin)

	relayChan = make(chan []byte)
	go func() {
		for {
			var msg = make([]byte, 512)
			n, err := ws.Read(msg)
			if err != nil {
				log.Printf("Error reading incoming message: %s", err.Error())
			}

			decrypted, err := security.Decrypt(rc.key, msg[:n])
			if err != nil {
				log.Printf("Error decrypting incoming message: %s", err.Error())
			} else {
				relayChan <- decrypted
			}
		}
	}()

	return
}

func (rc *RelayClient) SendMessage(destination string, msg interface{}, marshaller func(interface{}) ([]byte, error)) (msgResponse []byte, err error) {
	return rc.sendMessageHTTP(destination, msg, marshaller)
}

func (rc *RelayClient) sendMessageHTTP(destination string, msg interface{}, marshaller func(interface{}) ([]byte, error)) (
	msgResponse []byte, err error) {
	marshalled, err := marshaller(msg)
	if err != nil {
		return
	}

	encryptedMarshalled, err := security.Encrypt(rc.key, marshalled)
	if err != nil {
		return
	}

	request, err := http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s/messages?destination=%s", rc.relayHost, destination),
		bytes.NewBuffer(encryptedMarshalled))
	if err != nil {
		return
	}
	request.Header.Set("Content-Type", "application/octet-stream")
	request.Header.Set("Content-Length", strconv.Itoa(len(encryptedMarshalled)))
	request.ContentLength = int64(len(encryptedMarshalled))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return
	}

	_, err = ioutil.ReadAll(io.LimitReader(response.Body, 1048576))
	return
}
