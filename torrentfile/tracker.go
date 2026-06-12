package torrentfile

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/jackpal/bencode-go"
)

type bencodeTrackerResp struct {
	Interval      int    `bencode:"interval"`
	Peers         string `bencode:"peers"` // binary str of IPs & ports
	FailureReason string `bencode:"failure reason"`
}

func (t *TorrentFile) RequestPeers(peerID [20]byte, port uint16) ([]Peer, error) {
	base, err := url.Parse(t.Announce)
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
	trackerURL := base.String()

	fmt.Printf("Connecting to tracker: %s\n", t.Announce)

	req, err := http.NewRequest("GET", trackerURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "qBittorrent/4.3.9")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tracker returned HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	trackerResp := bencodeTrackerResp{}
	err = bencode.Unmarshal(resp.Body, &trackerResp)
	if err != nil {
		return nil, err
	}

	if trackerResp.FailureReason != "" {
		return nil, fmt.Errorf("tracker rejected request: %s", trackerResp.FailureReason)
	}

	peers, err := UnmarshalPeers([]byte(trackerResp.Peers))
	if err != nil {
		return nil, err
	}
	return peers, nil
}
