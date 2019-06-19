package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

/* udpCast is cloning incoming udp packets to multiple destinations, focus on easy and fast packet dumping
 */

/* fixed packetSize for better results for optimizer */
const packetSize = 1520

/* loveley main function */
func main() {
	/* some local variables */
	var err error

	/* create new slice of receiver udpaddrs, maximum 3 */
	set := make([]*net.UDPAddr, 0)

	/* listener for incoming */
	incomingAddr := flag.String("ia", "", "listening address")
	incomingPort := flag.Int("ip", 2009, "listening port")
	sDestinations := flag.String("dest", "127.0.0.1:2010", "destinations, comma seperated")
	outPort := flag.Int("lp", 12345, "local source port for outgoing packets")
	outAddr := flag.String("la", "", "local source address for outgoing packets")

	/* source generation for outgoing packets */
	sAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", *outAddr, *outPort))
	if err != nil {
		log.Fatalf("cant resolve endpoint: %s", err)
	}

	conn, err := net.ListenUDP("udp", sAddr)
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

		/* resolve client endpoint */
		daddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", singleClient[0], singleClient[1]))

		/* check for any error condition in creating udp endpoint */
		if err != nil {
			log.Fatalf("cant resolve destination endpoint: %s", err)
		}

		/* append endpoint to the destination slice */
		set = append(set, daddr)

		/* happy printing out message */
		fmt.Printf("Created destination %s:%s\n", singleClient[0], singleClient[1])

	}

	/* create server listener */
	iAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", *incomingAddr, *incomingPort))
	if err != nil {
		log.Fatalf("cant resolve endpoint: %s", err)
	}

	/* start listening socket for incoming connection */
	s, err := net.ListenUDP("udp", iAddr)
	if err != nil {
		log.Fatalf("cant listen endpoint: %s", err)
	}

	/* listener goroutine */
	func(s *net.UDPConn) {
		buffer := [packetSize]byte{}
		for {

			/* read from source */
			n, err := s.Read(buffer[0:])
			if err != nil {
				log.Println(err)
			}

			/* write to destinations */
			for _, v := range set {
				_, err = conn.WriteTo(buffer[0:n], v)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}(s)

	os.Exit(0)

}
