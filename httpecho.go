package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const usage = `Usage of httpecho: echo server accepting malformed HTTP request
  -s --serve      serve continuously (default: wait for 1 request)
  -t, --timeout   timeout to close connection in millisecond. Needed for closing http request. (default: 500)
  -d, --dump      dump incoming request to a file (default: only print to stdout)
  -p, --port      listening on specific port (default: 8888)
  --tls           use TLS encryption for communication
  -h, --help      dump incoming request to a file (default: only print to stdout) 
`

func main() {
	//-s
	var serve bool
	flag.BoolVar(&serve, "serve", false, "Serve continuously (default: wait for 1 request)")
	flag.BoolVar(&serve, "s", false, "Serve continuously (default: wait for 1 request)")

	// -t
	var timeout int
	flag.IntVar(&timeout, "timeout", 200, "Timeout to close connection. Needed for closing http request. (default: 200)")
	flag.IntVar(&timeout, "t", 200, "Timeout to close connection. Needed for closing http request. (default: 200)")

	//-d
	var dump string
	flag.StringVar(&dump, "dump", "", "Dump incoming request to a file (default: only print to stdout)")
	flag.StringVar(&dump, "d", "", "Dump incoming request to a file (default: only print to stdout)")

	//-p
	var port string
	flag.StringVar(&port, "port", "8888", "Listening on specific port (default: 8888)")
	flag.StringVar(&port, "p", "8888", "Listening on specific port (default: 8888)")

	//--tls
	var encrypted bool
	flag.BoolVar(&encrypted, "tls", false, "Use TLS encryption for communication")

	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	port = ":" + port

	var ln net.Listener
	var err error

	if encrypted {
		home := os.Getenv("HOME")
		cer, err := tls.LoadX509KeyPair(home+"/.httpecho/server.crt", home+"/.httpecho/server.key")
		if err != nil {
			log.Println(err)
			return
		}
		config := &tls.Config{Certificates: []tls.Certificate{cer}}
		ln, err = tls.Listen("tcp", port, config)
	} else {
		ln, err = net.Listen("tcp", port)
	}
	if err != nil {
		log.Println(err)
		return
	}
	defer ln.Close()
	if serve {

		for {
			conn, err := ln.Accept()
			conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second)) //close http request
			if err != nil {
				log.Println(err)
			}
			//fmt.Println("-----------------")
			go handleConnection(conn, dump, timeout)
		}

	} else { //only 1 time

		conn, err := ln.Accept()
		//conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond)) //close http request
		if err != nil {
			log.Println(err)
		}
		handleConnection(conn, dump, timeout)
	}

}

func handleConnection(conn net.Conn, dump string, timeout int) {
	defer conn.Close()
	writeFile := false
	var request string

	if dump != "" {
		writeFile = true
		f, err := os.Create(dump)

		if err != nil {
			log.Fatal(err)
		}

		defer f.Close()

	}
	r := bufio.NewReader(conn)

	go func() { //handle packet without '\n' ending character
		time.Sleep(time.Duration(timeout) * time.Millisecond)
		residue, err := r.Peek(r.Buffered())
		if err != nil {
			log.Println(err)
		}
		n, err := conn.Write(residue)
		if err != nil {
			log.Println(n, err)
			return
		}
		conn.Close()
	}()

	for {
		msg, err := r.ReadString('\n')
		//print log
		fmt.Print(msg)
		//write to file
		if writeFile {
			request += msg
		}
		//handle read error
		if err != nil && err != io.EOF {
			if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "closed network connection") { //avoid timeout error
				log.Println(err)
			}
			if writeFile { //Write request received in file
				f, err := os.Create(dump)
				if err != nil {
					log.Fatal(err)
				}
				defer f.Close()

				_, err2 := f.WriteString(request)
				if err2 != nil {
					log.Fatal(err2)
				} else {
					fmt.Println("dump request in:", dump)
				}
			}

			return
		}

		n, err := conn.Write([]byte(msg))
		if err != nil {
			log.Println(n, err)
			return
		}
	}
}
