package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

/* structure we use for copying the incoming stream */
type transmit struct {
	data [1560]byte
	len  int
}

/* simple destination type for looping */
type destinations struct {
	host string
	port int
}

/* structure of each udp destination*/
type udpDest struct {
	daddr   *net.UDPAddr
	dconn   *net.UDPConn
	network string
	queue   chan transmit
}

/* read data and transmit over the UDP socket */
func (u *udpDest) poller() {
	for {
		select {
		case toTransmit := <-u.queue:
			u.dconn.Write(toTransmit.data[0:toTransmit.len])
		}
	}
}

/* newUdpDest creates new initialized udp destination structure */
func newUdpDest(oAddr string, oPort int) udpDest {
	var err error

	/* client with queue and endpoint */
	udp := udpDest{network: fmt.Sprintf("%s:%d", oAddr, oPort)}
	udp.daddr, err = net.ResolveUDPAddr("udp", udp.network)

	/* check for error condition */
	if err != nil {
		log.Fatalf("cant resolve destination endpoint: %s", err)
	}

	/* pseudo dial output for outgoing connection */
	udp.dconn, err = net.DialUDP("udp", nil, udp.daddr)

	/* check for error condition */
	if err != nil {
		log.Fatalf("cant build udp socket to server: %s", err)
	}

	/* new buffered channel for receiving data */
	udp.queue = make(chan transmit, 1000)

	/* return our freshly created udp object */
	return udp

}

/* loveley main function */
func main() {
	/* some local variables */
	var err error
	reader := make(chan transmit)

	/* create new slices of receiver channels, maximum 10 */
	set := make([]chan transmit, 0)

	/* listener for incoming */
	incomingAddr := flag.String("ia", "", "listening address")
	incomingPort := flag.Int("ip", 2009, "listening port")
	sDestinations := flag.String("dest", "127.0.0.1:2010,127.0.0.1:2011", "destinations, comma seperated")

	/* parse arguments */
	flag.Parse()

	/* split input into clients */
	clients := strings.Split(*sDestinations, ",")

	/* parse the input slice of clients and create new objects */
	for i := range clients {

		/* create  and startup new destinations */
		singleClient := strings.Split(clients[i], ":")

		/* check if we got a host and a port entry */
		if len(singleClient) < 2 {
			log.Println("Cant parse client ", singleClient)
			continue
		}

		/* parse to Int for later usage inside newUdpDest */
		port, err := strconv.Atoi(singleClient[1])
		if err != nil {
			log.Fatalf("Cant parse client %s", singleClient)
		}

		udp := newUdpDest(singleClient[0], port)

		/* put the receivement channel of the newly created udp object into the slice */
		set = append(set, udp.queue)

		/* start listener thread */
		go func() {
			fmt.Println("Created destination ", udp.network)
			udp.poller()
		}()

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

	/* reader goroutine is pushing
	duplicating data to the waiting, udp-objects */
	go func(set []chan transmit) {
		for {
			select {
			case toTransmit := <-reader:
				{
					for _, v := range set {
						v <- toTransmit
					}
				}
			}
		}
	}(set)

	/* listener goroutine */
	func(s *net.UDPConn) {
		buffer := [1560]byte{}
		for {

			/* read from source */
			n, err := s.Read(buffer[0:])
			if err != nil {
				log.Println(err)
			}
			/* Write to channel */
			reader <- transmit{data: buffer, len: n}
		}
	}(s)

}
