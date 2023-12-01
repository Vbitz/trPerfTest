package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

var (
	listen  = flag.String("listen", "", "listen on address")
	connect = flag.String("connect", "", "connect to address")
	count   = flag.Int64("count", 1000, "the number of packets to send")
	output  = flag.String("output", "", "output a list of times into a text file")
)

const (
	PACKET_SIZE = 1024
)

func main() {
	flag.Parse()

	if *listen != "" {
		listen, err := net.ListenPacket("udp", *listen)
		if err != nil {
			log.Fatal(err)
		}
		defer listen.Close()

		log.Printf("listening on: %s", listen.LocalAddr())

		buf := make([]byte, PACKET_SIZE)

		for {
			n, addr, err := listen.ReadFrom(buf)
			if err != nil {
				log.Fatal(err)
			}

			_, err = listen.WriteTo(buf[:n], addr)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else if *connect != "" {
		var outputFile io.Writer
		if *output != "" {
			f, err := os.Create(*output)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			outputFile = f
		}

		conn, err := net.Dial("udp", *connect)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		log.Printf("connected to: %s", conn.RemoteAddr())

		buf := make([]byte, PACKET_SIZE)

		_, err = rand.Read(buf)
		if err != nil {
			log.Fatal(err)
		}

		var total uint64 = 0

		var i int64
		for i = 0; i < *count; i++ {
			start := time.Now()
			_, err = conn.Write(buf)
			if err != nil {
				log.Fatal(err)
			}

			_, err = conn.Read(buf)
			if err != nil {
				log.Fatal(err)
			}

			end := time.Since(start)

			if outputFile != nil && i%100 == 0 {
				fmt.Fprintf(outputFile, "%d\n", end.Nanoseconds())
			}

			total += uint64(end.Nanoseconds())
		}

		log.Printf("completed: total=%dns average=%fns count=%d", total, float64(total)/float64(*count), *count)
	} else {
		flag.Usage()
	}
}
