package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"Torrent67/p2p"
	"Torrent67/torrentfile"
)

func main() {
	infoCmd := flag.String("i", "", "Print info about a .torrent file")
	peersCmd := flag.String("cp", "", "List active peers for a .torrent file")
	downloadCmd := flag.String("d", "", "Download the .torrent file")
	outDir := flag.String("loc", ".", "Download directory location (defaults to current folder)")

	flag.Usage = func() {
		fmt.Println("Tor67 > A lightweight, custom BitTorrent Client")
		fmt.Println("\nUsage:")
		fmt.Println("  tor67 [flags]")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  tor67 -i debian.torrent              (Show info)")
		fmt.Println("  tor67 -cp debian.torrent             (List peers)")
		fmt.Println("  tor67 -d debian.torrent -loc /tmp    (Download to specific folder)")
	}

	flag.Parse()

	if *infoCmd == "" && *peersCmd == "" && *downloadCmd == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *infoCmd != "" {
		printInfo(*infoCmd)
		return
	}

	var peerID [20]byte
	copy(peerID[:], []byte("-qB4390-"))
	_, err := rand.Read(peerID[8:])
	if err != nil {
		log.Fatalf("Error generating peer ID: %v", err)
	}

	if *peersCmd != "" {
		listPeers(*peersCmd, peerID)
		return
	}

	if *downloadCmd != "" {
		downloadTorrent(*downloadCmd, peerID, *outDir)
		return
	}
}

// Helper Funcs

func printInfo(path string) {
	tf, err := torrentfile.Open(path)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	fmt.Println("---Torrent Info---")
	fmt.Printf("File Name:    %s\n", tf.Name)
	fmt.Printf("File Size:    %d bytes\n", tf.Length)
	fmt.Printf("Piece Lenght: %d bytes\n", tf.PieceLength)
	fmt.Printf("Total Pieces: %d\n", len(tf.PieceHashes))
	fmt.Printf("Info Hash:    %x\n", tf.InfoHash)
	fmt.Println("Trackers:")
	for _, tr := range tf.TrackerList {
		fmt.Printf("  - %s\n", tr)
	}
}

func listPeers(path string, peerID [20]byte) {
	tf, err := torrentfile.Open(path)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	fmt.Println("Scraping trackers for peers...")
	peers, err := tf.RequestPeers(peerID, 6881)
	if err != nil {
		log.Fatalf("Error requesting peers: %v", err)
	}
	fmt.Printf("\nFound %d Active Peers:\n", len(peers))
	for _, peer := range peers {
		fmt.Printf("  - %s:%d\n", peer.IP, peer.Port)
	}
}

func downloadTorrent(path string, peerID [20]byte, outDir string) {
	tf, err := torrentfile.Open(path)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}

	fmt.Printf("Scraping swarm for '%s'...\n", tf.Name)
	peers, err := tf.RequestPeers(peerID, 6881)
	if err != nil {
		log.Fatalf("Error requesting peers: %v", err)
	}
	fmt.Printf("Found %d peers. Starting download engine...\n", len(peers))

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

	outPath := filepath.Join(outDir, tf.Name)
	err = os.WriteFile(outPath, downloadedFile, 0644)
	if err != nil {
		log.Fatalf("Failed to save file to %s: %v", outPath, err)
	}

	fmt.Printf("\n[SUCCESS] File downloaded & verified at: %s\n", outPath)
}
