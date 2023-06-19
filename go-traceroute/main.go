package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"syscall"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	protocolICMP = 1
	maxHops      = 64
)

var (
	sport = flag.Int("sport", 12345, "source port")
	dport = flag.Int("p", 33434, "destination port")
)

func main() {
	flag.Parse()
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s host", os.Args[0])
	}
	dst := os.Args[1]
	timeout := 3 * time.Second
	dstAddr := &syscall.SockaddrInet4{}
	copy(dstAddr.Addr[:], net.ParseIP(dst).To4())
	// 得到本机的地址
	local := localAddr()
	// 生成一个icmp conn, 用来读取ICMP回包
	rconn, err := icmp.ListenPacket("ip4:icmp", local)
	if err != nil {
		log.Fatalf("Failed to create ICMP listener: %v", err)
	}
	defer rconn.Close()
	// 得到进程ID
	id := uint16(os.Getpid() & 0xffff)
	// 生成一个用来写udp的raw socket,这里使用syscall.IPPROTO_RAW,因为我们需要自己设置IP Header
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer syscall.Close(fd)
	// 设置此项，我们自己手工组装IP header
	err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)
	if err != nil {
		fmt.Println(err)
		return
	}
	// TTL递增探测
loop_ttl:
	for ttl := 1; ttl <= maxHops; ttl++ {
		*dport++
		// 拼装一个IP+UDP的包， IP header使用指定的id和ttl, udp 的payload使用一段字符串
		data, err := encodeUDPPacket(local, dst, id, uint8(ttl), []byte("Hello, are you there?"))
		if err != nil {
			log.Printf("%d: %v", ttl, err)
			continue
		}

		// 发送UDP包
		start := time.Now()
		err = syscall.Sendto(fd, data, 0, dstAddr)
		if err != nil {
			log.Printf("%d: %v", ttl, err)
			continue
		}
		// listen for the reply
		replyBytes := make([]byte, 1500)
		if err := rconn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			log.Fatalf("Failed to set read deadline: %v", err)
		}
		// 尝试读取3次
		// 你也可以使用死循环+一个超时来控制
		for i := 0; i < 3; i++ {
			n, peer, err := rconn.ReadFrom(replyBytes)
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					fmt.Printf("%d: *\n", ttl)
					continue loop_ttl
				} else {
					log.Printf("%d: Failed to parse ICMP message: %v", ttl, err)
				}
				continue
			}
			// 解析 ICMP message
			replyMsg, err := icmp.ParseMessage(protocolICMP, replyBytes[:n])
			if err != nil {
				log.Printf("%d: Failed to parse ICMP message: %v", ttl, err)
				continue
			}
			// 如果是 DestinationUnreachable,说明探测到了目的主机
			if replyMsg.Type == ipv4.ICMPTypeDestinationUnreachable {
				te, ok := replyMsg.Body.(*icmp.DstUnreach)
				if !ok {
					continue
				}
				// 抽取匹配项
				ipAndPayload, err := extractIPAndPayload(te.Data)
				if err != nil {
					continue
				}
				// 判断这个回包是否是本次请求匹配？
				if ipAndPayload.Dst != dst || ipAndPayload.Src != local || ipAndPayload.SrcPort != *sport || ipAndPayload.DstPort != *dport || ipAndPayload.ID != id {
					continue
				}
				// 如果匹配，这已经到达目的主机了，把时延打印出来，返回
				fmt.Printf("%d: %v %v\n", ttl, peer, time.Since(start))
				return
			}
			// 如果是中间设备而回包
			if replyMsg.Type == ipv4.ICMPTypeTimeExceeded {
				te, ok := replyMsg.Body.(*icmp.TimeExceeded)
				if !ok {
					continue
				}
				// 抽取匹配项
				ipAndPayload, err := extractIPAndPayload(te.Data)
				if err != nil {
					continue
				}
				// 判断这个回包是否是本次请求匹配？
				if ipAndPayload.Dst != dst || ipAndPayload.Src != local || ipAndPayload.SrcPort != *sport || ipAndPayload.DstPort != *dport || ipAndPayload.ID != id {
					continue
				}
				// 打印中间设备IP和时延
				fmt.Printf("%d: %v %v\n", ttl, peer, time.Since(start))
				if peer.String() == dst {
					return
				}
				continue loop_ttl
			}
		}
	}
}

// 构造IP包和UDP包
func encodeUDPPacket(localIP, dstIP string, id uint16, ttl uint8, payload []byte) ([]byte, error) {
	ip := &layers.IPv4{
		Id:       uint16(id), // ID
		SrcIP:    net.ParseIP(localIP),
		DstIP:    net.ParseIP(dstIP),
		Version:  4,
		TTL:      ttl, // ttl
		Protocol: layers.IPProtocolUDP,
	}
	udp := &layers.UDP{
		SrcPort: layers.UDPPort(*sport),
		DstPort: layers.UDPPort(*dport),
	}
	udp.SetNetworkLayerForChecksum(ip)
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	err := gopacket.SerializeLayers(buf, opts, ip, udp, gopacket.Payload(payload))
	return buf.Bytes(), err
}

type ipAndPayload struct {
	Src     string
	Dst     string
	SrcPort int
	DstPort int
	ID      uint16
	TTL     int
	Payload []byte
}

// 抽取匹配项
func extractIPAndPayload(body []byte) (*ipAndPayload, error) {
	if len(body) < ipv4.HeaderLen {
		return nil, fmt.Errorf("ICMP packet too short: %d bytes", len(body))
	}
	ipHeader, payload := body[:ipv4.HeaderLen], body[ipv4.HeaderLen:] // 抽取ip header和payload(UDP packet的前8个字节)
	iph, err := ipv4.ParseHeader(ipHeader)
	if err != nil {
		return nil, fmt.Errorf("Error parsing IP header: %s", err)
	}
	srcPort := binary.BigEndian.Uint16(payload[0:2]) // 前两个字节是源端口
	dstPort := binary.BigEndian.Uint16(payload[2:4]) // 接下来两个字节是目的端口
	return &ipAndPayload{
		Src:     iph.Src.String(),
		Dst:     iph.Dst.String(),
		SrcPort: int(srcPort),
		DstPort: int(dstPort),
		ID:      uint16(iph.ID),
		TTL:     iph.TTL,
		Payload: payload,
	}, nil
}
func localAddr() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	panic("no local IP address found")
}
