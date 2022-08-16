package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/ariary/go-utils/pkg/color"
	"github.com/ariary/quicli/pkg/quicli"
)

var verbose bool

func main() {
	log.SetFlags(log.Lshortfile)

	cli := quicli.Cli{
		Usage:       "httpecho [flags]",
		Description: "Echo server accepting malformed HTTP request",
		Flags: quicli.Flags{
			{Name: "serve", Description: "Serve continuously. If not only wait for 1 request"},
			{Name: "timeout", Default: 200, Description: "Timeout to close connection. Needed for closing http request."},
			{Name: "dump", Default: "", Description: "Dump incoming request to a file. If not used then print to stdout"},
			{Name: "port", Default: "8888", Description: "Listening on specific port"},
			{Name: "tls", Description: "Use TLS encryption for communication"},
			{Name: "verbose", Description: "Display request with special characters"},
		},
		CheatSheet: quicli.Examples{
			{Title: "Quicly register a request", CommandLine: "httpecho -d request"},
			{Title: "Wait request indefinitely", CommandLine: "httpecho -s"},
			{Title: "Observe special characters in request", CommandLine: "httpecho -v"},
		},
	}
	cfg := cli.Parse()

	serve := cfg.GetBoolFlag("serve")
	timeout := cfg.GetIntFlag("timeout")
	dump := cfg.GetStringFlag("dump")
	port := cfg.GetStringFlag("port")
	encrypted := cfg.GetBoolFlag("tls")
	verbose = cfg.GetBoolFlag("verbose")
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
	//HTTP1.1 OK!
	n, err := conn.Write([]byte("HTTP/1.1 200 OK\n\n"))
	if err != nil {
		log.Println(n, err)
		return
	}
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
		if err != nil && err != io.EOF {
			log.Println(err)
		}
		n, err := conn.Write(residue)
		if err != nil && err != io.EOF {
			log.Println(n, err)
			return
		}
		conn.Close()
	}()

	for {
		msg, err := r.ReadString('\n')
		//print log
		if verbose {
			msgDebug := strings.ReplaceAll(string(msg), "\r", color.Green("\\r"))
			msgDebug = strings.ReplaceAll(string(msgDebug), "\n", color.Green("\\n\n"))
			fmt.Print(msgDebug)
		} else {
			fmt.Print(msg)
		}

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
