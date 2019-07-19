package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

func getaddrs() (*net.TCPAddr, *net.TCPAddr, error) {

	serve_addr := "0.0.0.0:443"
	connect_addr := "127.0.0.1:80"

	args := os.Args[1:]

	args_l := len(args)

	if args_l == 1 && args[0] == "help" {
		fmt.Println("./simpletlsproxy [connect_to_addr [serve_on_addr]]")
		log.Fatal("showing help")
	}

	if args_l > 0 {
		connect_addr = args[0]
	}

	if args_l > 1 {
		serve_addr = args[1]
	}

	if args_l > 2 {
		return nil, nil, errors.New("too many arguments")
	}

	saddr, err := net.ResolveTCPAddr("tcp", serve_addr)
	if err != nil {
		return nil, nil, errors.New("error parsing serve_on_address: " + err.Error())
	}

	caddr, err := net.ResolveTCPAddr("tcp", connect_addr)
	if err != nil {
		return nil, nil, errors.New("error parsing connect_to_addr: " + err.Error())
	}

	return saddr, caddr, nil
}

func main() {

	saddr, caddr, err := getaddrs()
	if err != nil {
		log.Fatal("error: ", err)
	}

	certificate, err := tls.LoadX509KeyPair("/tls/cert.pem", "/tls/key.pem")
	if err != nil {
		log.Fatal("error loading keyfile or certificate: ", err)
	}

	tls_cfg := &tls.Config{
		Certificates: []tls.Certificate{certificate},
	}

	ln, err := net.ListenTCP("tcp", saddr)
	if err != nil {
		log.Fatal("error serving: ", err)
	}

	m := sync.Mutex{}

	for {

		m.Lock()

		conn, err := ln.Accept()
		if err != nil {
			log.Fatal("error accepting inbound connection: ", err)
		}

		m.Unlock()

		go func(conn net.Conn, tls_cfg *tls.Config) {
			tls_srv := tls.Server(conn, tls_cfg)

			client, err := net.DialTCP("tcp", nil, caddr)
			if err != nil {
				log.Fatalf("error dialing %v: %v", caddr, err)
			}

			defer func() {
				client.Close()
				conn.Close()
			}()

			g := sync.WaitGroup{}
			g.Add(2)

			defer g.Wait()

			go func() {
				defer g.Done()
				_, err = io.Copy(tls_srv, client)
				if err != nil {
					log.Fatalf("error streaming to %v: %v", caddr, err)
				}
			}()

			go func() {
				defer g.Done()
				_, err = io.Copy(client, tls_srv)
				if err != nil {
					log.Fatalf("error streaming from %v: %v", caddr, err)
				}
			}()

		}(conn, tls_cfg)
	}

}
