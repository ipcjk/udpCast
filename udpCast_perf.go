// +build perf

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
)

/* udpCast is cloning incoming udp packets to multiple destinations, focus on easy and fast packet dumping
 */

/* loveley main function */
func main() {
	/* some local variables */
	var err error

	/* create new slice of receiver udpaddrs, maximum 3 */
	set := make([]*net.UDPAddr, maxDestination)

	/* listener for incoming */
	incomingAddr := flag.String("ia", "", "listening address")
	incomingPort := flag.Int("ip", 2009, "listening port")
	sDestinations := flag.String("dest", "127.0.0.1:2010", "destinations, comma seperated")
	localEndPointPort := flag.Int("lp", 12345, "local source port for outgoing packets")

	/* sender endpoint generation for outgoing packets */
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: *localEndPointPort})
	if err != nil {
		log.Fatal("Listen", err)
	}

	/* parse arguments */
	flag.Parse()

	/* split input into clients */
	clients := strings.Split(*sDestinations, ",")

	/* parse the input slice of clients and create new objects */
	for i := range clients {

		/* split host and port part for new destinations */
		singleClient := strings.Split(clients[i], ":")

		/* client with queue and endpoint */
		daddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", singleClient[0], singleClient[1]))

		/* check for any error condition in creating udp endpoint */
		if err != nil {
			log.Fatalf("cant resolve destination endpoint: %s", err)
		}

		/* append endpoint to the destination slice */
		set = append(set, daddr)

		/* happy printing out message */
		fmt.Printf("Created destination for perf version %s:%s\n", singleClient[0], singleClient[1])

	}

	/* create server listener */
	sAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", *incomingAddr, *incomingPort))
	if err != nil {
		log.Fatalf("cant resolve endpoint: %s", err)
	}

	/* start listening socket for incoming connection */
	s, err := net.ListenUDP("udp", sAddr)
	if err != nil {
		log.Fatalf("cant listen endpoint: %s", err)
	}

	/* listener goroutine */
	func(s *net.UDPConn) {
		buffer := [packetSize]byte{}
		for {

			/* read from source */
			n, _ := s.Read(buffer[0:])

			/* write to destinations */
			for _, v := range set {
				conn.WriteTo(buffer[0:n], v)
			}
		}
	}(s)

}
