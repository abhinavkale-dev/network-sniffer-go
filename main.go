package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func main() {
	deviceName := "en0"
	fmt.Printf("Using interface: %s\n", deviceName)
	fmt.Println("Starting packet capture... (Press Ctrl+C to stop)")
	fmt.Println(strings.Repeat("=", 60))

	handle, err := pcap.OpenLive(deviceName, 1600, true, pcap.BlockForever)
	if err != nil {
		log.Fatalf("Error opening device %s: %v\nTry running with sudo", deviceName, err)
	}
	defer handle.Close()

	if err := handle.SetBPFFilter("icmp or tcp or udp"); err != nil {
		log.Printf("Warning: Could not set filter: %v", err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		for packet := range packetSource.Packets() {
			printPacketInfo(packet)
		}
	}()

	<-c
	fmt.Println("\nStopping packet capture...")
}

func printPacketInfo(packet gopacket.Packet) {
	if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
		timestamp := time.Now().Format("Jan 02 15:04:05")
		ip := ipLayer.(*layers.IPv4)
		fmt.Printf("[%s] %s -> %s ", timestamp, ip.SrcIP, ip.DstIP)

		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			tcp := tcpLayer.(*layers.TCP)
			fmt.Printf("TCP:%d->%d", tcp.SrcPort, tcp.DstPort)
		} else if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
			udp := udpLayer.(*layers.UDP)
			fmt.Printf("UDP:%d->%d", udp.SrcPort, udp.DstPort)
		} else if packet.Layer(layers.LayerTypeICMPv4) != nil {
			fmt.Printf("ICMP")
		}
		fmt.Println()
	}
}
