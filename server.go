package juju

import (
	"encoding/binary"
	"io"
	"log"
	"net"
)

const IncomingDataBufferSize int = 32_768

type Server struct {
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

		log.Printf("[-] %s", conn.RemoteAddr().String())
	}()

	for {
		var data []byte

		for {
			dataBuf := make([]byte, IncomingDataBufferSize)
			n, err := conn.Read(dataBuf)

			if err != nil {
				if err != io.EOF {
					log.Printf("failed reading from client: %s %d", err, n)
				}

				return
			}

			data = append(data, dataBuf[:n]...)
			processedDataLength := server.processIncomingData(conn, data)

			if processedDataLength > 0 {
				data = data[processedDataLength:]
			}
		}
	}
}

func (server *Server) processIncomingData(conn net.Conn, data []byte) int {
	var offset = 0
	var processedLength = 0

	for offset < len(data) {
		cmd := data[offset]

		if cmd != PublishCommand || offset+CommandTypeBufferSize >= len(data) || offset+CommandTypeBufferSize+CommandDataLengthBufferSize >= len(data) || offset+CommandMetadataTotalBufferSize >= len(data) {
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

		go server.handleCommand(conn, cmd, identifier, cmdData)

		offset = offset + CommandMetadataTotalBufferSize + int(dataLength)
		processedLength += CommandMetadataTotalBufferSize + int(dataLength)
	}

	return processedLength
}

func (server *Server) handleCommand(conn net.Conn, cmd Command, identifier string, data []byte) {
	// TODO
}
