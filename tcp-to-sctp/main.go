package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ishidawataru/sctp"
)

var (
	// dialWithPROXY bool
	remoteSCTPIP   string
	remoteSCTPPort string
	listenIP       string
	listenPort     string
	debug          int
)

func copyConn(src net.Conn) {
	remoteAddress := src.RemoteAddr()
	localAddresss := src.LocalAddr()

	addr, _ := sctp.ResolveSCTPAddr("sctp", remoteSCTPIP+":"+remoteSCTPPort)
	dst, err := sctp.DialSCTP("sctp", nil, addr)
	if err != nil {
		if debug > 0 {
			log.Println("DialSCTP Error:", err.Error(), "To:", remoteSCTPIP+":"+remoteSCTPPort)
		}
		src.Close()
		return
	}
	if debug > 0 {
		log.Printf("Dialed SCTP to %s", dst.RemoteAddr())
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
		log.Println("Ended connection to:", remoteSCTPIP+":"+remoteSCTPPort, "From:", remoteAddress, "At:", localAddresss)
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

	// dialWithPROXY = false
	// remoteSCTPIP = "127.0.0.1"
	// remoteSCTPPort = "3128"
	// listenIP = ""
	// listenPort = "3128"

	flag.StringVar(&listenIP, "listen-ip", "127.0.0.1", "The IP which the service will listen to")
	flag.StringVar(&listenPort, "listen-port", "3128", "The Port which the service will listen to")

	flag.StringVar(&remoteSCTPIP, "connect-ip", "127.0.0.1", "The IP which the service will proxy to")
	flag.StringVar(&remoteSCTPPort, "connect-port", "3128", "The Port which the service will proxy to")

	// flag.IntVar(&retries, "retries", 4, "The number of http and connect retries")
	flag.IntVar(&debug, "debug", 0, "The Debug level of the service")

	// flag.BoolVar(&dialWithPROXY, "proxy-protocol-connect", false, "Connect to the SCTP host with a PROXY protocol header")

	flag.Parse()
}

func main() {

	log.Println("Starting TCP-to-SCTP proxy service")

	ln, err := net.Listen("tcp", listenIP+":"+listenPort)
	if err != nil {
		log.Fatal("connection error:" + err.Error())
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("failed to accept: %v", err)
			continue
		}
		if debug > 0 {
			log.Printf("Accepted Connection from RemoteAddr: %s", conn.RemoteAddr())
		}
		go copyConn(conn)
	}
}
