package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"

	"Torrent67/p2p"
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

	fmt.Println("--- Torrent Parsed Successfully ---")
	fmt.Printf("Tracker URL: %s\n", tf.Announce)
	fmt.Printf("File name:   %s\n", tf.Name)
	fmt.Printf("Info Hash:   %x\n", tf.InfoHash)

	var peerID [20]byte
	copy(peerID[:], []byte("-q84390-"))
	_, err = rand.Read(peerID[8:])
	if err != nil {
		log.Fatalf("Error generating peer ID: %v", err)
	}

	peers, err := tf.RequestPeers(peerID, 6881)
	if err != nil {
		log.Fatalf("Error requesting peers: %v", err)
	}

	torrent := p2p.Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    tf.InfoHash,
		PieceHashes: tf.PieceHashes,
		PieceLength: tf.PieceLength,
		Length:      tf.Length,
		Name:        tf.Name,
	}

	downloadedFile := torrent.Download()

	err = os.WriteFile(tf.Name, downloadedFile, 0644)
	if err != nil {
		log.Fatalf("Failed to save file: %v", err)
	}

	fmt.Printf("\n [SUCCESS]! fully verified & downloaded '%s' to disk! \n", tf.Name)
}
