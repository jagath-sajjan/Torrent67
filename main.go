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

	// Disguise our client as qBittorrent v4.3.9
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

	var downloadedBlock []byte

PeerLoop:
	for _, peer := range peers {
		address := fmt.Sprintf("%s:%d", peer.IP, peer.Port)

		fmt.Printf("\n=========================================\n")
		fmt.Printf("Dialing TCP connection to %s...\n", address)

		conn, err := net.DialTimeout("tcp", address, 3*time.Second)
		if err != nil {
			fmt.Printf("Failed to connect: %v. Moving to next peer...\n", err)
			continue
		}

		fmt.Println("Connected! Sending BitTorrent Handshake...")

		req := handshake.Handshake{
			Pstr:     "BitTorrent protocol",
			InfoHash: tf.InfoHash,
			PeerID:   peerID,
		}

		_, err = conn.Write(req.Serialize())
		if err != nil {
			fmt.Printf("Failed to send handshake: %v\n", err)
			conn.Close()
			continue
		}

		res, err := handshake.Read(conn)
		if err != nil {
			fmt.Printf("Failed to read handshake response: %v\n", err)
			conn.Close()
			continue
		}

		if !bytes.Equal(res.InfoHash[:], tf.InfoHash[:]) {
			fmt.Printf("Expected infohash %x but got %x\n", tf.InfoHash, res.InfoHash)
			conn.Close()
			continue
		}

		fmt.Printf("Handshake successful! Peer ID: %x\n", res.PeerID)

		fmt.Println("Sending 'Interested' message...")
		interestedMsg := &message.Message{ID: message.MsgInterested}

		_, err = conn.Write(interestedMsg.Serialize())
		if err != nil {
			fmt.Printf("Failed to send interested message: %v\n", err)
			conn.Close()
			continue
		}

		unchoked := false

		fmt.Println("Listening for peer messages...")

	MessageLoop:
		for {
			msg, err := message.Read(conn)
			if err != nil {
				fmt.Printf("Connection dropped by peer (%v). Moving to next peer...\n", err)
				break MessageLoop
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
			fmt.Println("\n--- TIME TO DOWNLOAD PIECES ---")

			const blockSize = 16384

			reqMsg := message.FormatRequest(0, 0, blockSize)

			_, err = conn.Write(reqMsg.Serialize())
			if err != nil {
				fmt.Printf("Failed to send block request: %v\n", err)
				conn.Close()
				continue
			}

			fmt.Printf("Requested the first %d bytes. Waiting for delivery...\n", blockSize)

			for {
				msg, err := message.Read(conn)
				if err != nil {
					fmt.Printf("Peer dropped us before sending data: %v\n", err)
					break
				}

				if msg == nil {
					continue
				}

				if msg.ID == message.MsgPiece {
					block, err := message.ParsePiece(0, msg.Payload)
					if err != nil {
						fmt.Printf("Failed to parse piece: %v\n", err)
						break
					}

					fmt.Printf("\nNOM NOM NOM! Successfully downloaded %d bytes of Ubuntu!\n", len(block))

					downloadedBlock = block
					conn.Close()

					break PeerLoop
				}

				fmt.Printf("Ignoring %s message while waiting for our data...\n", msg.Name())
			}
		}

		conn.Close()
	}

	if downloadedBlock != nil {
		err = os.WriteFile("ubuntu_block_0.dat", downloadedBlock, 0644)
		if err != nil {
			log.Fatalf("Failed to write block to disk: %v", err)
		}

		fmt.Println("\n SUCCESS! Saved the raw data to 'ubuntu_block_0.dat'")
	} else {
		log.Fatal("\nAll peers failed or dropped us! Try running again or use a torrent with more seeders.")
	}
}
