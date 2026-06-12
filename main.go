package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jackpal/bencode-go"
)

type bencodeInfo struct {
	Pieces       string `bencode:"pieces"`       // concatenated 20 byte SHA-1 hashes
	PiecesLength int    `bencode:"piece length"` // no of bytes  per piece
	Length       int    `bencode:"length"`       // total len of the file
	Name         string `bencode:"name"`         // name of the file
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"` // tracker url
	Info     bencodeInfo `bencode:"info"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide a path to a .torrent file")
	}

	torrentPath := os.Args[1]
	file, err := os.Open(torrentPath)
	if err != nil {
		log.Fatal("Error opening file: %v", err)
	}
	defer file.Close()

	var bto bencodeTorrent

	err = bencode.Unmarshal(file, &bto)
	if err != nil {
		log.Fatal("Error prasing bencode: %v", err)
	}

	fmt.Println("--- Torrent Parsed Successfully ---")
	fmt.Println("Tracker Url: %s\n", bto.Announce)
	fmt.Println("File Name: %s\n", bto.Info.Name)
	fmt.Println("File Lenght: %d bytes\n", bto.Info.Length)
}
