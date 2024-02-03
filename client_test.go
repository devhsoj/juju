package juju

import (
	"log"
	"testing"
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

	for i := 0; i < 10; i++ {
		if err := client.Publish("test", []byte("Test message."), true); err != nil {
			t.Fatalf("failed to publish: %s", err)
		}
	}

	if err := client.Disconnect(); err != nil {
		t.Fatalf("failed to disconnect: %s", err)
	}
}

func TestClientSubscribe(t *testing.T) {
	client := Client{}

	if err := client.Connect(""); err != nil {
		t.Fatalf("failed to connect: %s", err)
	}

	err := client.Subscribe("test", func(data []byte) {
		log.Println(string(data))
	})

	if err != nil {
		t.Fatalf("failed to subscribe: %s", err)
	}

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
		if err := client.Publish("test", msg, false); err != nil {
			b.Fatalf("failed to publish: %s", err)
		}
	}

	if err := client.Disconnect(); err != nil {
		b.Fatalf("failed to disconnect: %s", err)
	}
}
