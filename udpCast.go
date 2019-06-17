package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

func main() {
	/* listener */
	lAddr := flag.String("ia", "", "listening address")
	lPort := flag.Int("ip", 2009, "listening port")

	/* destination */
	oAddr := flag.String("oa", "127.0.0.1", "destination server")
	oPort := flag.Int("op", 2010, "destination port")

	/* parse arguments */
	flag.Parse()

	/* destination endpoint */
	daddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", *oAddr, *oPort))
	if err != nil {
		log.Fatalf("cant resolve destination endpoint: %s", err)
	}

	/* server endpoint */
	sAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", *lAddr, *lPort))
	if err != nil {
		log.Fatalf("cant resolve endpoint: %s", err)
	}

	/* start listening socket for incoming connection */
	s, err := net.ListenUDP("udp", sAddr)
	if err != nil {
		log.Fatalf("cant listen endpoint: %s", err)
	}

	/* pseudo dial output for outgoing connection */
	dConn, err := net.DialUDP("udp", nil, daddr)
	if err != nil {
		log.Fatalf("cant build udp socket to server: %s", err)
	}

	func(s *net.UDPConn, d *net.UDPAddr) {
		buffer := [1518]byte{}

		for {

			/* read from source */
			n, err := s.Read(buffer[0:])
			if err != nil {
				log.Println(err)
			}

			/* relay to destination */
			_, err = dConn.Write(buffer[0:n])
			if err != nil {
				log.Println(err)
			}

		}
	}(s, daddr)

	fmt.Println("End of the line")

}
