package main

import (
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func main() {
	conn, err := net.ListenPacket("ip4:udp", "192.168.0.1")
	if err != nil {
		panic(err)
	}
	buf := make([]byte, 1024)
	for {
		n, peer, err := conn.ReadFrom(buf)
		if err != nil {
			panic(err)
		}
		fmt.Printf("received request from %s: %s\n", peer.String(), buf[8:n])
		data, _ := encodeUDPPacket("192.168.0.1", "127.0.0.1", []byte("hello world"))
		_, err = conn.WriteTo(data, &net.IPAddr{IP: net.ParseIP("127.0.0.1")})
		if err != nil {
			panic(err)
		}
	}
}

func encodeUDPPacket(src, dst string, payload []byte) ([]byte, error) {
	ip := &layers.IPv4{
		SrcIP:    net.ParseIP(src),
		DstIP:    net.ParseIP(dst),
		Version:  4,
		Protocol: layers.IPProtocolUDP,
	}

	udp := &layers.UDP{
		SrcPort: layers.UDPPort(0),
		DstPort: layers.UDPPort(8972),
	}

	udp.SetNetworkLayerForChecksum(ip)

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	err := gopacket.SerializeLayers(buf, opts, udp, gopacket.Payload(payload))

	return buf.Bytes(), err
}
