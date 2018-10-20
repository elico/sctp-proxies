package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/asaskevich/govalidator"

	"github.com/ishidawataru/sctp"
	"github.com/pires/go-proxyproto"
)

var (
	dialWithPROXY bool
	remoteTCPIP   string
	remoteTCPPort string
	listenIP      string
	listenPort    string
	debug         int
)

func parseIP(s string) (net.IP, uint16, error) {
	ip, port, err := net.SplitHostPort(s)
	if err != nil {
		return nil, uint16(0), err
	}

	ip2 := net.ParseIP(ip)
	if ip2 != nil {
		return nil, uint16(0), errors.New("invalid IP")
	}

	i, _ := strconv.Atoi(port)
	return ip2, uint16(i), nil
}

func copyConn(src net.Conn) {
	remoteAddress := src.RemoteAddr()
	localAddresss := src.LocalAddr()

	dst, err := net.Dial("tcp", remoteTCPIP+":"+remoteTCPPort)
	if err != nil {
		if debug > 0 {
			log.Println("Dial TCP Error:", err.Error(), "To:", remoteTCPIP+":"+remoteTCPPort)
		}
		src.Close()
		return
	}
	if debug > 0 {
		log.Println("Dialed TCP to", remoteTCPIP+":"+remoteTCPPort)
	}

	if dialWithPROXY {
		// Write PROXY Protocl Header first

		remoteIP, remotePort, _ := parseIP(remoteAddress.String())
		localIP, localPort, _ := parseIP(localAddresss.String())
		if govalidator.IsIPv4(remoteIP.String()) {
			proxyProtocolV1Header := &proxyproto.Header{Version: byte(1), TransportProtocol: proxyproto.TCPv4, SourceAddress: remoteIP, SourcePort: remotePort, DestinationAddress: localIP, DestinationPort: localPort}
			wrote, err := proxyProtocolV1Header.WriteTo(src)
			if err != nil {
				if debug > 0 {
					log.Println("Dial TCP Error writing PROXYProtocol header:", err.Error(), "To:", remoteTCPIP+":"+remoteTCPPort)
				}
				src.Close()
				return
			}
			if debug > 0 {
				log.Println("Wrote PROXY Protocol header bytes:", wrote, "To:", remoteTCPIP+":"+remoteTCPPort)
			}
		} else {
			src.Write([]byte("This service supports only Version 4 sources\n"))
			src.Close()
			return
		}
	}

	done := make(chan struct{})

	go func() {
		defer src.Close()
		defer dst.Close()
		io.Copy(dst, src)
		done <- struct{}{}
	}()

	go func() {
		defer src.Close()
		defer dst.Close()
		io.Copy(src, dst)
		done <- struct{}{}
	}()

	<-done
	<-done
	if debug > 0 {
		log.Println("Ended connection to:", remoteTCPIP+":"+remoteTCPPort, "From:", remoteAddress, "At:", localAddresss)
	}
}

func init() {
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			//handle SIGINT
			os.Exit(0)
		case syscall.SIGTERM:
			//handle SIGTERM
			os.Exit(0)
		}
	}()

	dialWithPROXY = false
	// remoteTCPIP = "10.0.0.138"
	// remoteTCPPort = "8080"
	// listenIP = ""
	// listenPort = "3128"

	flag.StringVar(&listenIP, "listen-ip", "127.0.0.1", "The IP which the service will listen to")
	flag.StringVar(&listenPort, "listen-port", "3128", "The Port which the service will listen to")

	flag.StringVar(&remoteTCPIP, "connect-ip", "10.0.0.138", "The IP which the service will proxy to")
	flag.StringVar(&remoteTCPPort, "connect-port", "8080", "The Port which the service will proxy to")

	// flag.IntVar(&retries, "retries", 4, "The number of http and connect retries")
	flag.IntVar(&debug, "debug", 0, "The Debug level of the service")

	flag.BoolVar(&dialWithPROXY, "proxy-protocol-connect", false, "Connect to the TCP host with a PROXY protocol header")

	flag.Parse()
}

func main() {

	log.Println("Starting SCTP-to-TCP service")

	// sctpAddr, _ := sctp.ResolveSCTPAddr("sctp", listenIP+":"+intListenPort)

	// type SCTPAddr struct {
	// 	IPAddrs []net.IPAddr
	// 	Port    int
	// }

	q := net.ParseIP(listenIP)
	addr := net.IPAddr{IP: q, Zone: ""}
	var intListenPort int
	if govalidator.IsPort(listenPort) {
		intListenPort, _ = strconv.Atoi(listenPort)
	} else {

	}
	var addresses []net.IPAddr
	addresses = append(addresses, addr)
	sctpAddr := &sctp.SCTPAddr{IPAddrs: addresses, Port: intListenPort}

	ln, err := sctp.ListenSCTP("sctp", sctpAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("Listen on %s", ln.Addr())

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Failed to SCTP Accept: ", err.Error())
			continue
		}
		if debug > 0 {
			log.Printf("Accepted Connection from RemoteAddr: %s", conn.RemoteAddr())
		}
		go copyConn(conn)
	}
}
