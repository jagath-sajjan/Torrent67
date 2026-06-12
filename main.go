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
	"Torrent67/message"
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
	copy(peerID[:], []byte("-qB4390-"))
	_, err = rand.Read(peerID[8:])
	if err != nil {
		log.Fatalf("Error generating peer ID: %v", err)
	}

	peers, err := tf.RequestPeers(peerID, 6881)
	if err != nil {
		log.Fatalf("Error requesting peers: %v", err)
	}

	fmt.Printf("Found %d peers!\n", len(peers))

	if len(peers) == 0 {
		log.Fatalf("Tracker returned 0 peers. The tracker might be blocking us, or the torrent is dead.")
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

	fmt.Printf("Handshake successful! Peer ID: %x\n", res.PeerID)

	fmt.Println("Sending 'Interested' message...")
	interestedMsg := &message.Message{ID: message.MsgInterested}
	_, err = conn.Write(interestedMsg.Serialize())
	if err != nil {
		log.Fatalf("Failed to send interested message: %v", err)
	}

	unchoked := false

	fmt.Println("Listening for peer messages...")
MessageLoop:
	for {
		msg, err := message.Read(conn)
		if err != nil {
			log.Fatalf("Error reading message: %v", err)
		}

		if msg == nil {
			continue
		}

		switch msg.ID {
		case message.MsgChoke:
			fmt.Println("\nPeer choked us.")
			unchoked = false
		case message.MsgUnchoke:
			fmt.Println("\nSUCCESS! Peer unchoked us!")
			unchoked = true
			break MessageLoop
		case message.MsgBitfield:
			fmt.Printf("\nReceived Bitfield! (Payload size: %d bytes)\n", len(msg.Payload))
		case message.MsgHave:
			fmt.Print(".")
		default:
			fmt.Printf("\nReceived message: %s\n", msg.Name())
		}
	}

	if unchoked {
		fmt.Println("\n--- next nom nom pieces time ---")
	}
}
