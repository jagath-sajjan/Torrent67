package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"

	"Torrent67/torrentfile"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide a path to a .torrent file")
	}

	torrentPath := os.Args[1]

	tf, err := torrentfile.Open(torrentPath)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}

	fmt.Println("--- Torrent Parsed Successpully ---")
	fmt.Printf("Tracker URL: %s\n", tf.Announce)
	fmt.Printf("File name:    %s\n", tf.Name)
	fmt.Printf("Info Hash:  %x\n", tf.InfoHash)

	var peerID [20]byte
	_, err = rand.Read(peerID[:])
	if err != nil {
		log.Fatalf("Error generating peer ID: %v", err)
	}

	err = tf.RequestPeers(peerID, 6881)
	if err != nil {
		log.Fatalf("Error requesting peers: %v", err)
	}
}
