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
  -t, --timeout timeout to close connection in millisecond. Needed for closing http request. (default: 500)
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
		conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond)) //close http request
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

	go func() { //handle packet without '\n' ending character
		time.Sleep(time.Duration(100) * time.Millisecond)
		residue, err := r.Peek(r.Buffered())
		if err != nil {
			log.Println(err)
		}
		n, err := conn.Write(residue)
		if err != nil {
			log.Println(n, err)
			return
		}
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

		n, err := conn.Write([]byte(msg))
		if err != nil {
			log.Println(n, err)
			return
		}
	}
}
