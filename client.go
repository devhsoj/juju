package juju

import (
	"encoding/binary"
	"net"
)

type Client struct {
	// internals
	conn      net.Conn
	connected bool
}

func (client *Client) Connect(serverAddress string) error {
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

	client.connected = false

	return nil
}

func (client *Client) Publish(channelName string, data []byte) error {
	identifierBuf := EncodeIdentifier(channelName)
	dataLengthBuf := make([]byte, 8)

	binary.BigEndian.PutUint64(dataLengthBuf, uint64(len(data)))

	req := []byte{
		PublishCommand,
	}

	req = append(req, dataLengthBuf...)
	req = append(req, identifierBuf...)
	req = append(req, data...)

	if _, err := client.conn.Write(req); err != nil {
		return err
	}

	return nil
}
