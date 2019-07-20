package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
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

	log.Print("Welcome to Simple TLS proxy")
	log.Print("https://github.com/AnimusPEXUS/simpletlsproxy")

	saddr, caddr, err := getaddrs()
	if err != nil {
		log.Fatal("error: ", err)
	}

	log.Print("going to listening on ", saddr)
	log.Print("connections will be TLS proxied to ", caddr)

	certificate, err := tls.LoadX509KeyPair("/tls/cert.pem", "/tls/key.pem")
	if err != nil {
		log.Fatal("error loading keyfile or certificate: ", err)
	}

	log.Print("TLS certificate loaded")

	tls_cfg := &tls.Config{
		Certificates: []tls.Certificate{certificate},
	}

	ln, err := net.ListenTCP("tcp", saddr)
	if err != nil {
		log.Fatal("error serving: ", err)
	}

	log.Print("listening socket opened")

	log.Print("accepting loop begins")

	// var waiters_count uint64

	var conn_id uint64 = 0

	for {

		conn, err := ln.Accept()
		if err != nil {
			log.Fatal("error accepting inbound connection: ", err)
		}

		conn_id += 1

		log.Printf("accepted connection (%d) %v", conn_id, conn.RemoteAddr())

		go func(conn_id uint64, conn net.Conn, tls_cfg *tls.Config) {
			tls_srv := tls.Server(conn, tls_cfg)

			client, err := net.DialTCP("tcp", nil, caddr)
			if err != nil {
				log.Fatalf(" (%d) error dialing %v: %v", conn_id, caddr, err)
			}

			defer func() {
				client.Close()
				tls_srv.Close()
			}()

			// g := sync.WaitGroup{}
			// g.Add(2)

			// waiters_count += 2

			// defer func() {
			// 	log.Printf(" waiting %d end (still waiting for %d)", conn_id, waiters_count)
			// 	g.Wait()
			// 	log.Printf(" waiting %d end done (still waiting for %d)", conn_id, waiters_count)
			// }()

			_, err = io.Copy(client, tls_srv)
			if err != nil {
				log.Fatalf(" (%d) error streaming from %v: %v", conn_id, caddr, err)
			}

			// go func() {
			// 	defer func() {
			// 		waiters_count -= 1
			// 		g.Done()
			// 		log.Printf(" (%d) copier exited tls_srv -> client", conn_id)
			// 	}()
			// 	_, err = io.Copy(client, tls_srv)
			// 	if err != nil {
			// 		log.Fatalf(" (%d) error streaming from %v: %v", conn_id, caddr, err)
			// 	}
			// }()

			_, err = io.Copy(tls_srv, client)
			if err != nil {
				log.Fatalf(" (%d) error streaming to %v: %v", conn_id, caddr, err)
			}

			// go func() {
			// 	defer func() {
			// 		waiters_count -= 1
			// 		g.Done()
			// 		log.Printf(" (%d) copier exited client -> tls_srv", conn_id)
			// 	}()
			// 	_, err = io.Copy(tls_srv, client)
			// 	if err != nil {
			// 		log.Fatalf(" (%d) error streaming to %v: %v", conn_id, caddr, err)
			// 	}
			// }()

		}(conn_id, conn, tls_cfg)
	}

}
