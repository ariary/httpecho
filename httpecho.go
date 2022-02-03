package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const usage = `Usage of httpecho: echo server accepting malformed HTTP request
  -s --serve serve continuously (default: wait for 1 request)
  -t, --timeout timeout to close connection. Needed for closing http request. (default: 1)
  -d, --dump dump incoming request to a file (default: only print to stdout)
  -p, --port listening on specific port (default: 8888)
  -h, --help dump incoming request to a file (default: only print to stdout) 
`

func main() {
	//-s
	var serve bool
	flag.BoolVar(&serve, "serve", false, "Serve continuously (default: wait for 1 request)")
	flag.BoolVar(&serve, "s", false, "Serve continuously (default: wait for 1 request)")

	// -t
	var timeout int
	flag.IntVar(&timeout, "timeout", 1, "Timeout to close connection. Needed for closing http request. (default: 1)")
	flag.IntVar(&timeout, "t", 1, "Timeout to close connection. Needed for closing http request. (default: 1)")

	//-d
	var dump string
	flag.StringVar(&dump, "dump", "", "Dump incoming request to a file (default: only print to stdout)")
	flag.StringVar(&dump, "d", "", "Dump incoming request to a file (default: only print to stdout)")

	//-p
	var port string
	flag.StringVar(&port, "port", "8888", "Listening on specific port (default: 8888)")
	flag.StringVar(&port, "p", "8888", "Listening on specific port (default: 8888)")

	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	port = ":" + port
	if serve {

		ln, err := net.Listen("tcp", port)
		if err != nil {
			log.Println(err)
			return
		}
		defer ln.Close()
		for {
			conn, err := ln.Accept()
			conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second)) //close http request
			if err != nil {
				log.Println(err)
			}
			go handleConnection(conn, dump)
		}

	} else { //only 1 time
		ln, err := net.Listen("tcp", ":8888")
		if err != nil {
			log.Println(err)
			return
		}
		defer ln.Close()

		conn, err := ln.Accept()
		conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second)) //close http request
		if err != nil {
			log.Println(err)
		}
		handleConnection(conn, dump)
	}

}

func handleConnection(conn net.Conn, dump string) {
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

	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			if !strings.Contains(err.Error(), "timeout") { //avoid timeout error
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

		fmt.Print(msg)

		if writeFile {
			request += msg
		}

		n, err := conn.Write([]byte(msg))
		if err != nil {
			log.Println(n, err)
			return
		}
	}
}