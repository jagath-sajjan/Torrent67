package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"Torrent67/handshake"
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

	peers, err := tf.RequestPeers(peerID, 6881)
	if err != nil {
		log.Fatalf("Error requesting peers: %v", err)
	}

	fmt.Printf("Found %d peers!\n", len(peers))
	for _, peer := range peers {
		fmt.Printf(" - %s:%d\n", peer.IP, peer.Port)
	}

	peer := peers[0]
	address := fmt.Sprintf("%s:%d", peer.IP, peer.Port)
	fmt.Printf("\nDialing TCP connection to %s...\n", address)

	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		log.Fatalf("Failed to connect to peer: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connected! Sending BitTorrent Handshake...")

	req := handshake.Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: tf.InfoHash,
		PeerID:   peerID,
	}

	_, err = conn.Write(req.Serialize())
	if err != nil {
		log.Fatalf("Failed to send handshake: %v", err)
	}

	res, err := handshake.Read(conn)
	if err != nil {
		log.Fatalf("Failed to read handshake response: %v", err)
	}

	if !bytes.Equal(res.InfoHash[:], tf.InfoHash[:]) {
		log.Fatalf("Expected infohash %x but got %x", tf.InfoHash, res.InfoHash)
	}
	fmt.Println("Handshake successpul! Peer ID: %x\n", res.PeerID)
}
