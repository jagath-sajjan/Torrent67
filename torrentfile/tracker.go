package torrentfile

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/jackpal/bencode-go"
)

type bencodeTrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"` // binary str of IPs & ports
}

func (t *TorrentFile) RequestPeers(peerID [20]byte, port uint16) error {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return err
	}

	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"}, // barks at tracker to send compact peer list
		"left":       []string{strconv.Itoa(t.Length)},
	}
	base.RawQuery = params.Encode()
	trackerURL := base.String()

	fmt.Printf("Connecting to tracker: %s\n", t.Announce)

	resp, err := http.Get(trackerURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	trackerResp := bencodeTrackerResp{}
	err = bencode.Unmarshal(resp.Body, &trackerResp)
	if err != nil {
		return err
	}

	fmt.Printf("Success! Tracker interval: %d secs\n", trackerResp.Interval)
	fmt.Printf("Raw peers string length: %d bytes\n", len(trackerResp.Peers))

	return nil
}
