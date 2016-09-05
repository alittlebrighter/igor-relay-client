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
	id           string
	relayHost    string
	socketConn   *websocket.Conn
	key          []byte
	marshaller   func(interface{}) ([]byte, error)
	unmarshaller func(data []byte, v interface{}) error
}

func (rc *RelayClient) Marshaller() func(interface{}) ([]byte, error) {
	return rc.marshaller
}

func (rc *RelayClient) Unmarshaller() func(data []byte, v interface{}) error {
	return rc.unmarshaller
}

func NewRelayClient(id, relayHost string, key []byte, marshaller func(interface{}) ([]byte, error), unmarshaller func(data []byte, v interface{}) error) *RelayClient {
	return &RelayClient{id: id, relayHost: relayHost, key: key, marshaller: marshaller, unmarshaller: unmarshaller}
}

func (rc *RelayClient) OpenSocket() error {
	// origin can be a bogus URL so we'll just use it to identify the connection on the server
	origin := "http://" + rc.id
	url := "ws://" + rc.relayHost + "/socket"

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		return err
	}
	rc.socketConn = ws
	return nil
}

// ReadMessages opens a websocket or polls on host arg identifying itself with controllerID arg and
// returns a channel that relays messages coming down from the server
func (rc *RelayClient) ReadMessages() (relayChan chan []byte) {
	relayChan = make(chan []byte)

	if rc.socketConn != nil {
		go func() {
			for {
				var msg = make([]byte, 512)
				n, err := rc.socketConn.Read(msg)
				if err != nil {
					log.Printf("Error reading incoming message: %s", err.Error())
					close(relayChan)
					break
				}

				decrypted, err := security.DecryptFromString(rc.key, string(msg[:n]))
				if err != nil {
					log.Printf("Error decrypting incoming message: %s", err.Error())
				} else {
					relayChan <- decrypted
				}
			}
		}()
	} else {
		go func() {
			request, err := http.NewRequest(
				"GET",
				fmt.Sprintf("http://%s/messages", rc.relayHost),
				nil)
			if err != nil {
				log.Println("Error building request: " + err.Error())
				return
			}
			response, err := http.DefaultClient.Do(request)
			if err != nil {
				log.Println("Error making request: " + err.Error())
				return
			}

			// download mailbox contents
			msgResponse, err := ioutil.ReadAll(io.LimitReader(response.Body, 1048576))
			var msgs [][]byte
			err = rc.unmarshaller(msgResponse, msgs)
			if err != nil {
				log.Println("Error parsing request: " + err.Error())
				return
			}

			for _, msg := range msgs {
				decrypted, err := security.DecryptFromString(rc.key, string(msg))
				if err != nil {
					log.Printf("Error decrypting incoming message: %s", err.Error())
				} else {
					relayChan <- decrypted
				}
			}
			close(relayChan)
		}()
	}

	return
}

func (rc *RelayClient) SendMessage(env *Envelope) (msgResponse []byte, err error) {
	if rc.socketConn != nil {
		return rc.sendMessageWS(env)
	}

	return rc.sendMessageHTTP(env)
}

func (rc *RelayClient) sendMessageHTTP(env *Envelope) (msgResponse []byte, err error) {
	reqBody, err := rc.marshaller(env)

	request, err := http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s/messages", rc.relayHost),
		bytes.NewBuffer(reqBody))
	if err != nil {
		return
	}
	request.Header.Set("Content-Type", "application/octet-stream")
	request.Header.Set("Content-Length", strconv.Itoa(len(reqBody)))
	request.ContentLength = int64(len(reqBody))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return
	}

	msgResponse, err = ioutil.ReadAll(io.LimitReader(response.Body, 1048576))
	return
}

func (rc *RelayClient) sendMessageWS(env *Envelope) ([]byte, error) {
	reqBody, err := rc.marshaller(env)

	_, err = rc.socketConn.Write(reqBody)
	return []byte("Message sent and received."), err
}

type Envelope struct {
	Destination string
	TTL         int64
	Contents    string
}

func NewEnvelope(to string, ttl int64, contents interface{}, client *RelayClient) (env *Envelope, err error) {
	env = &Envelope{Destination: to, TTL: ttl}

	marshalled, err := client.Marshaller()(contents)
	if err != nil {
		return
	}

	encryptedMarshalled, err := security.EncryptToString(client.key, marshalled)
	if err != nil {
		return
	}
	env.Contents = encryptedMarshalled
	return
}
