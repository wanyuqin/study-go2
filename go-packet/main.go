package main

import (
	"fmt"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func main() {

	Recoding("en0")

}

var (
	snaplen = int32(65535)
	promisc = true
	timeout = -1 * time.Second
)

func Recoding(device string) {
	handle, err := pcap.OpenLive(device, snaplen, promisc, timeout)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer handle.Close()

	source := gopacket.NewPacketSource(handle, handle.LinkType())

	pc := source.Packets()

	for {
		select {
		case packet := <-pc:
			tcpLayer := packet.Layer(layers.LayerTypeTCP)
			if tcpLayer != nil {
				tcp := tcpLayer.(*layers.TCP)
				fmt.Printf("TCP Source Port: %d\n\n", tcp.SrcPort)
				fmt.Printf("TCP Dest Port %d\n\n", tcp.DstPort)
				fmt.Println()
			}

			ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
			if ipv4Layer != nil {
				iPv4 := ipv4Layer.(*layers.IPv4)
				fmt.Printf("IPV4 Source IP: %s\n\n", iPv4.SrcIP)
				fmt.Printf("IPV4 Dest IP %s\n\n", iPv4.DstIP)
				fmt.Println()
			}
		}
	}
}
