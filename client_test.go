package juju

import (
	"log"
	"testing"
	"time"
)

func TestClientConnection(t *testing.T) {
	client := Client{}

	if err := client.Connect(""); err != nil {
		t.Fatalf("failed to connect: %s", err)
	}

	if err := client.Disconnect(); err != nil {
		t.Fatalf("failed to disconnect: %s", err)
	}
}

func TestClientPublish(t *testing.T) {
	client := Client{}

	if err := client.Connect(""); err != nil {
		t.Fatalf("failed to connect: %s", err)
	}

	msg := []byte("This is a test message!")
	msgCount := 10

	start := time.Now().UnixMilli()
	for i := 0; i < msgCount; i++ {
		if err := client.Publish("test", msg); err != nil {
			t.Fatalf("failed to publish: %s", err)
		}
	}
	end := time.Now().UnixMilli()

	log.Printf("%d ms | %d mb | %d b", end-start, msgCount*(len(msg)+1+8+128)/1_024_000, msgCount*(len(msg)+1+8+128))

	if err := client.Disconnect(); err != nil {
		t.Fatalf("failed to disconnect: %s", err)
	}
}

func BenchmarkClientPublish(b *testing.B) {
	client := Client{}

	if err := client.Connect(""); err != nil {
		b.Fatalf("failed to connect: %s", err)
	}

	msg := []byte("This is a test message!")

	for n := 0; n < b.N; n++ {
		if err := client.Publish("test", msg); err != nil {
			b.Fatalf("failed to publish: %s", err)
		}
	}

	if err := client.Disconnect(); err != nil {
		b.Fatalf("failed to disconnect: %s", err)
	}
}
