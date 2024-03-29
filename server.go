package juju

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"time"
)

type serverClient struct {
	TimeLastPublishedTo time.Time
	// internals
	conn net.Conn
}

type Server struct {
	Subscriptions map[string][]serverClient
	// internals
	listener  net.Listener
	listening bool
}

func (server *Server) Start(listenAddress string) error {
	var err error

	if len(listenAddress) == 0 {
		listenAddress = "localhost:9261"
	}

	server.listener, err = net.Listen("tcp", listenAddress)

	if err != nil {
		return err
	}

	server.listening = true
	server.Subscriptions = make(map[string][]serverClient)

	for server.listening {
		conn, err := server.listener.Accept()

		if err != nil {
			log.Printf("failed to accept connection: %s", err)
			continue
		}

		log.Printf("[+] %s", conn.RemoteAddr().String())

		go server.handleConnection(conn)
	}

	return err
}

func (server *Server) handleConnection(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close connection: %s", err)
		}

		server.removeClient(conn)

		log.Printf("[-] %s", conn.RemoteAddr().String())
	}()

	for {
		var data []byte

		for {
			dataBuf := make([]byte, IncomingDataBufferSize)
			n, err := conn.Read(dataBuf)

			if err != nil {
				if err != io.EOF {
					log.Printf("failed reading from client: %s", err)
				}

				return
			}

			data = append(data, dataBuf[:n]...)
			processedDataOffset := server.processIncomingData(conn, data)

			if processedDataOffset > 0 {
				data = data[processedDataOffset:]
			}
		}
	}
}

func (server *Server) processIncomingData(conn net.Conn, data []byte) int {
	var offset = 0

	for offset < len(data) {
		cmd := data[offset]

		if (cmd == PublishCommand || cmd == ExclusivePublishCommand) && (offset+CommandTypeBufferSize >= len(data) || offset+CommandTypeBufferSize+CommandDataLengthBufferSize >= len(data) || offset+CommandMetadataTotalBufferSize >= len(data)) {
			break
		}

		dataLengthBuf := data[offset+CommandTypeBufferSize : offset+CommandTypeBufferSize+CommandDataLengthBufferSize]
		dataLength := binary.BigEndian.Uint64(dataLengthBuf)

		identifierBuf := data[offset+CommandTypeBufferSize+CommandDataLengthBufferSize : offset+CommandMetadataTotalBufferSize]
		identifier := DecodeIdentifierBuffer(identifierBuf)

		if offset+CommandMetadataTotalBufferSize+int(dataLength) > len(data) {
			break
		}

		cmdData := data[offset+CommandMetadataTotalBufferSize : offset+CommandMetadataTotalBufferSize+int(dataLength)]

		if err := server.handleCommand(conn, cmd, identifier, cmdData); err != nil {
			log.Printf("failed running command: %s", err)
		}

		offset += CommandMetadataTotalBufferSize + int(dataLength)
	}

	return offset
}

func (server *Server) handleCommand(conn net.Conn, cmd Command, identifier string, data []byte) error {
	if cmd == PublishCommand || cmd == ExclusivePublishCommand {
		if _, exists := server.Subscriptions[identifier]; !exists {
			return nil
		}

		dataLengthBuf := make([]byte, 8)

		binary.BigEndian.PutUint64(dataLengthBuf, uint64(len(data)))

		identifierBuf := EncodeIdentifier(identifier)

		res := []byte{
			SubscribeCommand,
		}

		res = append(res, dataLengthBuf...)
		res = append(res, identifierBuf...)
		res = append(res, data...)

		if cmd == PublishCommand {
			for i := 0; i < len(server.Subscriptions[identifier]); i++ {
				if _, err := server.Subscriptions[identifier][i].conn.Write(res); err != nil {
					return err
				}

				server.Subscriptions[identifier][i].TimeLastPublishedTo = time.Now()
			}

			return nil
		}

		// sort subscriptions by time last published to
		client := server.getLastPublishedToSubscribedClient(identifier)

		if _, err := client.conn.Write(res); err != nil {
			return err
		}

		client.TimeLastPublishedTo = time.Now()
	} else if cmd == SubscribeCommand {
		if _, exists := server.Subscriptions[identifier]; !exists {
			server.Subscriptions[identifier] = []serverClient{}
		}

		server.Subscriptions[identifier] = append(server.Subscriptions[identifier], serverClient{
			conn: conn,
		})
	}

	return nil
}

func (server *Server) removeClient(conn net.Conn) {
	addr := conn.RemoteAddr().String()

	for channelName, clients := range server.Subscriptions {
		for i, client := range clients {
			if client.conn.RemoteAddr().String() == addr {
				server.Subscriptions[channelName][i] = server.Subscriptions[channelName][len(server.Subscriptions[channelName])-1]
				server.Subscriptions[channelName] = server.Subscriptions[channelName][:len(server.Subscriptions[channelName])-1]
			}
		}
	}
}

func (server *Server) getLastPublishedToSubscribedClient(channelName string) *serverClient {
	_, exists := server.Subscriptions[channelName]

	if !exists {
		return nil
	}

	var client = &server.Subscriptions[channelName][0]

	for i := 1; i < len(server.Subscriptions[channelName]); i++ {
		if server.Subscriptions[channelName][i].TimeLastPublishedTo.UnixNano() < client.TimeLastPublishedTo.UnixNano() {
			client = &server.Subscriptions[channelName][i]
		}
	}

	return client
}

func NewServer() Server {
	return Server{
		Subscriptions: make(map[string][]serverClient),
		listening:     false,
	}
}
