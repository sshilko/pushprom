package main

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"log"
	"testing"
	"time"

	"github.com/messagebird/pushprom/delta"
	"github.com/messagebird/pushprom/metrics"

	"github.com/stretchr/testify/assert"
)

var conn *net.UDPConn

func TestMain(m *testing.M) {
	*udpListenAddress = ":3007"

	go listenUDP(context.Background())

	// wait for it to "boot"
	time.Sleep(time.Millisecond * 1000)

	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1"+*udpListenAddress)
	if err != nil {
		log.Fatal("Could not resolve server UPD address")
	}

	localAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		log.Fatal("Could not resolve local UDP address")
	}

	conn, err = net.DialUDP("udp", localAddr, serverAddr)
	if err != nil {
		log.Fatal("Could not establish UDP connection")
	}

	defer conn.Close()

	os.Exit(m.Run())
}

func TestUDP(t *testing.T) {
	delta := &delta.Delta{
		Type:   delta.GAUGE,
		Method: "set",
		Name:   "tree_width",
		Help:   "the width in meters of the tree",
		Value:  2.2,
	}

	buf, _ := json.Marshal(delta)

	_, err := conn.Write(buf)
	assert.Nil(t, err)

	time.Sleep(time.Millisecond * 500)

	ms, err := metrics.Fetch()
	assert.Nil(t, err)
	result, err := metrics.Read(ms, delta.Name, delta.Labels)
	if assert.Nil(t, err) {
		assert.Equal(t, delta.Value, result)
	}
}

func TestIncorrectJson(t *testing.T) {
	// First, let's write the correct value
	oldDelta := &delta.Delta{
		Type:   delta.GAUGE,
		Method: "set",
		Name:   "tree_width",
		Help:   "the width in meters of the tree",
		Value:  2.2,
	}
	buf, _ := json.Marshal(oldDelta)

	_, err := conn.Write(buf)
	assert.Nil(t, err)

	oldMetrics, err := metrics.Fetch()
	assert.Nil(t, err)
	oldResult, err := metrics.Read(oldMetrics, oldDelta.Name, oldDelta.Labels)
	if assert.Nil(t, err) {
		assert.Equal(t, oldDelta.Value, oldResult)
	}

	// Now, let's write the incorrect value
	buf = []byte("hello")

	_, err = conn.Write(buf)
	assert.Nil(t, err)

	// Last, let's write a new value
	newDelta := &delta.Delta{
		Type:   delta.GAUGE,
		Method: "set",
		Name:   "tree_width",
		Help:   "the width in meters of the tree",
		Value:  2.5,
	}
	buf, _ = json.Marshal(newDelta)

	_, err = conn.Write(buf)
	assert.Nil(t, err)

	newMetrics, err := metrics.Fetch()
	assert.Nil(t, err)
	newResult, err := metrics.Read(newMetrics, newDelta.Name, newDelta.Labels)
	if assert.Nil(t, err) {
		assert.Equal(t, newDelta.Value, newResult)
	}
}
