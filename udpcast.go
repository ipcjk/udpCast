package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
)

/* udpcast is cloning incoming udp packets to multiple destinations, focus on easy and fast packet dumping
 */

/* fixed packetSize for better results for optimizer */
const packetSize = 1520

/* loveley main function */
func main() {
	/* some local variables */
	var err error

	/* signal channel */
	//sigs := make(chan os.Signal, 1)
	//signal.Notify(sigs, os.Interrupt, syscall.SIGTERM&syscall.SIGINT)

	/* create new slice of receiver udpaddrs, maximum 3 */
	set := make([]*net.UDPAddr, 0)

	/* listener for incoming */
	incomingAddr := flag.String("ia", "", "listening address")
	incomingPort := flag.Int("ip", 2009, "listening port")
	interfaceName := flag.String("i", "", "interface name to receive e.g. multicast")
	ssm := flag.String("ssm", "", "source specific multicast")
	sDestinations := flag.String("dest", "127.0.0.1:2010", "destinations, comma seperated")
	ttl := flag.Int("ttl", 24, "time to live for IP packets")
	debug := flag.Bool("debug", false, "print incoming packet loop counter")

	/* parse arguments */
	flag.Parse()

	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		log.Fatal("Listen", err)
	}

	/* set ttl */
	f, err := conn.File()
	if err != nil {
		log.Fatal("File descriptor", err)
	}

	/* if its normal outbound */
	err = syscall.SetsockoptInt(int(f.Fd()), syscall.IPPROTO_IP, syscall.IP_TTL, *ttl)
	if err != nil {
		log.Println("Change ttl for unicast", err)
	}
	/* if its multicast */
	err = syscall.SetsockoptInt(int(f.Fd()), syscall.IPPROTO_IP, syscall.IP_MULTICAST_TTL, *ttl)
	if err != nil {
		log.Println("Change ttl for multicast", err)
	}

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
		if *debug {
			fmt.Printf("Created destination %s:%s\n", singleClient[0], singleClient[1])
		}

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

	/* if source is multicast, join a group */
	if iAddr.IP.IsMulticast() || iAddr.IP.IsLinkLocalMulticast() {
		if *interfaceName == "" {
			log.Fatal("Source is multicast, but interface name was not given, please correct")
			os.Exit(1)
		}
		iface, err := net.InterfaceByName(*interfaceName)
		if err != nil {
			log.Fatal(err)
		}

		p := ipv4.NewPacketConn(s)

		if *ssm != "" {
			ssmAddress, err := net.ResolveUDPAddr("udp", *ssm)
			if err != nil {
				log.Fatal(ssmAddress, err)
			}
			if *debug {
				fmt.Println("Joing SSM", *ssm, "on", iAddr)
			}
			err = p.JoinSourceSpecificGroup(iface, iAddr, ssmAddress)
		} else {
			err = p.JoinGroup(iface, iAddr)
		}

		if err != nil {
			log.Fatal(err)
		}

	}

	/* Fixme, don't ignore the p channel where multicast information is flooded */

	var packetCounter int64 = 1000
	/* listener goroutine */
	func(s *net.UDPConn) {

		buffer := [packetSize]byte{}
		for {
			if *debug {
				packetCounter++
				if packetCounter%1000 == 0 {
				}
			}

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
