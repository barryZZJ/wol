package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

func buildMagicPacket(mac string) ([]byte, error) {
	mac = strings.ReplaceAll(mac, ":", "")
	mac = strings.ReplaceAll(mac, "-", "")
	mac = strings.ToLower(mac)
	if len(mac) != 12 {
		return nil, fmt.Errorf("Invalid MAC address format")
	}
	macBytes, err := hex.DecodeString(mac)
	if err != nil {
		return nil, err
	}
	packet := make([]byte, 6+16*6)
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}
	for i := 0; i < 16; i++ {
		copy(packet[6+i*6:], macBytes)
	}
	return packet, nil
}

// getBroadcastIP returns the broadcast IP for the given interface name
func getBroadcastIP(ifaceName string) (string, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return "", fmt.Errorf("interface %s not found: %v", ifaceName, err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return "", fmt.Errorf("cannot get addresses for %s: %v", ifaceName, err)
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.To4() == nil {
			continue
		}
		ip := ipNet.IP.To4()
		mask := ipNet.Mask
		broadcast := make(net.IP, 4)
		for i := 0; i < 4; i++ {
			broadcast[i] = ip[i] | ^mask[i]
		}
		return broadcast.String(), nil
	}
	return "", fmt.Errorf("no IPv4 address found on interface %s", ifaceName)
}

func main() {
	port := flag.Int("p", 9, "Port")
	ip := flag.String("i", "255.255.255.255", "Broadcast IP address")
	iface := flag.String("if", "", "Network interface name (optional, e.g. br0)")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Usage: wol [-p PORT=9] [-i IP=255.255.255.255] [--if IFACE] MAC")
		fmt.Println("Example: wol -p 9 -i 192.168.1.255 01:23:45:67:89:ab")
		fmt.Println("      or wol -p 9 --if br0 01:23:45:67:89:ab")
		os.Exit(1)
	}

	mac := flag.Arg(0)

	// If interface is set, overwrite ip with interface's broadcast address
	if *iface != "" {
		bcast, err := getBroadcastIP(*iface)
		if err != nil {
			fmt.Println("Failed to get broadcast IP from interface:", err)
			os.Exit(1)
		}
		*ip = bcast
	}

	packet, err := buildMagicPacket(mac)
	if err != nil {
		fmt.Println("Failed to generate magic packet:", err)
		os.Exit(1)
	}

	addr := fmt.Sprintf("%s:%d", *ip, *port)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		fmt.Println("Failed to connect to target:", err)
		os.Exit(1)
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	if err != nil {
		fmt.Println("Failed to send magic packet:", err)
		os.Exit(1)
	}

	fmt.Printf("Sent WOL magic packet to %s port %d (MAC: %s)\n", *ip, *port, mac)
}
