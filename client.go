package juju

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
)

type Client struct {
	// internals
	conn       net.Conn
	connected  bool
	subscribed bool
}

var ErrClientSubscribed = errors.New("client is subscribed, command cannot be executed until the client unsubscribes")

func (client *Client) Connect(serverAddress string) error {
	if client.subscribed {
		return ErrClientSubscribed
	}

	var err error

	if len(serverAddress) == 0 {
		serverAddress = "localhost:9261"
	}

	client.conn, err = net.Dial("tcp", serverAddress)

	if err != nil {
		return err
	}

	client.connected = true

	return nil
}

func (client *Client) Disconnect() error {
	if err := client.conn.Close(); err != nil {
		return err
	}

	client.subscribed = false
	client.connected = false

	return nil
}

func (client *Client) Publish(channelName string, data []byte, exclusive bool) error {
	if client.subscribed {
		return ErrClientSubscribed
	}

	identifierBuf := EncodeIdentifier(channelName)
	dataLengthBuf := make([]byte, 8)

	binary.BigEndian.PutUint64(dataLengthBuf, uint64(len(data)))

	var cmd Command = PublishCommand

	if exclusive {
		cmd = ExclusivePublishCommand
	}

	req := []byte{
		cmd,
	}

	req = append(req, dataLengthBuf...)
	req = append(req, identifierBuf...)
	req = append(req, data...)

	if _, err := client.conn.Write(req); err != nil {
		return err
	}

	return nil
}

func (client *Client) Subscribe(channelName string, callback func(data []byte)) error {
	if client.subscribed {
		return ErrClientSubscribed
	}

	identifierBuf := EncodeIdentifier(channelName)
	dataLengthBuf := make([]byte, 8)

	req := []byte{
		SubscribeCommand,
	}

	req = append(req, dataLengthBuf...)
	req = append(req, identifierBuf...)

	if _, err := client.conn.Write(req); err != nil {
		return err
	}

	client.subscribed = true

	for client.subscribed {
		var data []byte

		for {
			dataBuf := make([]byte, IncomingDataBufferSize)
			n, err := client.conn.Read(dataBuf)

			if err != nil {
				if err != io.EOF {
					log.Printf("failed reading from juju server: %s", err)
				}

				return err
			}

			data = append(data, dataBuf[:n]...)
			processedDataLength := client.processIncomingSubscriptionData(data, callback)

			if processedDataLength > 0 {
				data = data[processedDataLength:]
			}
		}
	}

	return nil
}

func (client *Client) Unsubscribe() {
	client.subscribed = false
}

func (client *Client) processIncomingSubscriptionData(data []byte, callback func(data []byte)) int {
	var offset = 0
	var processedLength = 0

	for offset < len(data) {
		//cmd := data[offset]

		if offset+CommandTypeBufferSize >= len(data) || offset+CommandTypeBufferSize+CommandDataLengthBufferSize >= len(data) || offset+CommandMetadataTotalBufferSize >= len(data) {
			break
		}

		dataLengthBuf := data[offset+CommandTypeBufferSize : offset+CommandTypeBufferSize+CommandDataLengthBufferSize]
		dataLength := binary.BigEndian.Uint64(dataLengthBuf)

		//identifierBuf := data[offset+CommandTypeBufferSize+CommandDataLengthBufferSize : offset+CommandMetadataTotalBufferSize]
		//identifier := DecodeIdentifierBuffer(identifierBuf)

		if offset+CommandMetadataTotalBufferSize+int(dataLength) > len(data) {
			break
		}

		publishedData := data[offset+CommandMetadataTotalBufferSize : offset+CommandMetadataTotalBufferSize+int(dataLength)]

		go callback(publishedData)

		offset = offset + CommandMetadataTotalBufferSize + int(dataLength)
		processedLength += CommandMetadataTotalBufferSize + int(dataLength)
	}

	return processedLength
}

func NewClient() Client {
	return Client{
		connected:  false,
		subscribed: false,
	}
}
