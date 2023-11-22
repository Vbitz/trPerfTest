package main

import (
	"encoding/binary"
	"flag"
	"log"
	"net"
	"net/netip"

	"github.com/schollz/progressbar/v3"
)

var (
	listen    = flag.String("listen", "", "listen on a address for connections")
	connect   = flag.String("connect", "", "connect on a address for connections")
	size      = flag.Int64("size", 1024, "the number of blocks to transmit")
	blockSize = flag.Int64("bs", 1024, "the block size to transmit with")
)

type header struct {
	Size      int64
	BlockSize int64
}

func main() {
	flag.Parse()

	if *listen != "" {
		addr, err := netip.ParseAddrPort(*listen)
		if err != nil {
			log.Fatal(err)
		}
		conn, err := net.ListenUDP("udp", net.UDPAddrFromAddrPort(addr))
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		log.Printf("listening on: %s", conn.LocalAddr().String())

		var hdr header

		err = binary.Read(conn, binary.NativeEndian, &hdr)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("got config: size=%d blockSize=%d", hdr.Size, hdr.BlockSize)

		pb := progressbar.DefaultBytes(hdr.Size*hdr.BlockSize, "receiving")
		defer pb.Close()

		buf := make([]byte, hdr.BlockSize)

		var removeAddr *net.UDPAddr

		for i := 0; i < int(hdr.Size); i++ {
			var n int

			n, removeAddr, err = conn.ReadFromUDP(buf)
			if err != nil {
				log.Fatal(err)
			}

			_, err = conn.WriteToUDP(buf, removeAddr)
			if err != nil {
				log.Fatal(err)
			}

			pb.Add(n)
		}

		log.Printf("finished")
	} else if *connect != "" {
		addr, err := netip.ParseAddrPort(*connect)
		if err != nil {
			log.Fatal(err)
		}
		conn, err := net.DialUDP("udp", nil, net.UDPAddrFromAddrPort(addr))
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		log.Printf("connected to: %s", conn.RemoteAddr().String())

		var hdr header

		hdr.BlockSize = *blockSize
		hdr.Size = *size

		err = binary.Write(conn, binary.NativeEndian, &hdr)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("sent config: size=%d blockSize=%d", hdr.Size, hdr.BlockSize)

		pb := progressbar.DefaultBytes(hdr.Size*hdr.BlockSize, "sending")
		defer pb.Close()

		buf := make([]byte, hdr.BlockSize)

		for i := 0; i < int(hdr.Size); i++ {
			n, err := conn.Write(buf)
			if err != nil {
				log.Fatal(err)
			}

			_, err = conn.Read(buf)
			if err != nil {
				log.Fatal(err)
			}

			pb.Add(n)
		}

		log.Printf("finished")
	} else {
		flag.Usage()
	}
}
