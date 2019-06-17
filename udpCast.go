package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

type transmit struct {
	data [1518]byte
	len  int
}

type udpRecv struct {
	daddr   *net.UDPAddr
	dconn   *net.UDPConn
	network string
	queue   chan transmit
}

func main() {
	/* some local variables */
	var err error
	reader := make(chan transmit)

	/* listener */
	lAddr := flag.String("ia", "", "listening address")
	lPort := flag.Int("ip", 2009, "listening port")

	/* destination */
	oAddr := flag.String("oa", "127.0.0.1", "destination server")
	oPort := flag.Int("op", 2010, "destination port")

	/* parse arguments */
	flag.Parse()

	/* server listener */
	sAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", *lAddr, *lPort))
	if err != nil {
		log.Fatalf("cant resolve endpoint: %s", err)
	}

	/* start listening socket for incoming connection */
	s, err := net.ListenUDP("udp", sAddr)
	if err != nil {
		log.Fatalf("cant listen endpoint: %s", err)
	}

	/* client with queue and endpoint */
	udp := udpRecv{network: fmt.Sprintf("%s:%d", *oAddr, *oPort)}
	udp.queue = make(chan transmit, 100)
	udp.daddr, err = net.ResolveUDPAddr("udp", udp.network)
	if err != nil {
		log.Fatalf("cant resolve destination endpoint: %s", err)
	}

	/* pseudo dial output for outgoing connection */
	udp.dconn, err = net.DialUDP("udp", nil, udp.daddr)
	if err != nil {
		log.Fatalf("cant build udp socket to server: %s", err)
	}

	go func(client udpRecv) {
		for {
			select {
			case toTransmit := <-client.queue:
				log.Print("Packet angekommen")
				udp.dconn.Write(toTransmit.data[0:toTransmit.len])
			}
		}
	}(udp)

	/* reader goroutine is pushing
	duplicating data to waiting outgoing threads */
	go func() {
		for {
			select {
			case toTransmit := <-reader:
				{
					udp.queue <- toTransmit
				}
			}
		}

	}()

	func(s *net.UDPConn, d *net.UDPAddr) {
		buffer := [1518]byte{}
		for {
			/* read from source */
			n, err := s.Read(buffer[0:])
			if err != nil {
				log.Println(err)
			}
			/* Write to channel */
			reader <- transmit{data: buffer, len: n}
		}
	}(s, udp.daddr)

}
