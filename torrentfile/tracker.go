package torrentfile

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackpal/bencode-go"
)

type bencodeTrackerResp struct {
	Interval      int    `bencode:"interval"`
	Peers         string `bencode:"peers"` // binary str of IPs & ports
	FailureReason string `bencode:"failure reason"`
}

func (t *TorrentFile) RequestPeers(peerID [20]byte, port uint16) ([]Peer, error) {
	var allPeers []Peer
	seenPeers := make(map[string]bool)

	for _, tr := range t.TrackerList {
		if !strings.HasPrefix(tr, "http") {
			continue
		}

		fmt.Printf("Scraping tracker: %s\n", tr)
		peers, err := t.querySingleTracker(tr, peerID, port)
		if err != nil {
			fmt.Printf(" -> Tracker failed: %v\n", err)
			continue
		}

		for _, p := range peers {
			addr := fmt.Sprintf("%s:%d", p.IP, p.Port)
			if !seenPeers[addr] {
				seenPeers[addr] = true
				allPeers = append(allPeers, p)
			}
		}
	}
	if len(allPeers) == 0 {
		return nil, fmt.Errorf("all HTTP trackers failed or timed out")
	}
	return allPeers, nil
}

func (t *TorrentFile) querySingleTracker(trackedURL string, peerID [20]byte, port uint16) ([]Peer, error) {
	base, err := url.Parse(trackedURL)
	if err != nil {
		return nil, err
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

		req, err := http.NewRequest("GET", base.String(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "qBittorrent/4.3.9")

		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
		}

		trackerResp := bencodeTrackerResp{}
		err = bencode.Unmarshal(resp.Body, &trackerResp)
		if err != nil {
			return nil, err
		}
		if trackerResp.FailureReason != "" {
			return nil, fmt.Errorf(trackerResp.FailureReason)
		}

		return UnmarshalPeers([]byte(trackerResp.Peers))
}
