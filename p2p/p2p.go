package p2p

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"net"
	"time"

	"Torrent67/handshake"
	"Torrent67/message"
	"Torrent67/torrentfile"
)

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

type Torrent struct {
	Peers       []torrentfile.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

func (t *Torrent) Download() []byte {
	log.Println("Starting Torrent67 Worker Pool...")

	workQueue := make(chan *pieceWork, len(t.PieceHashes))
	results := make(chan *pieceResult)

	for index, hash := range t.PieceHashes {
		length := t.calculatePieceSize(index)
		workQueue <- &pieceWork{index, hash, length}
	}

	for _, peer := range t.Peers {
		go t.startWorker(peer, workQueue, results)
	}

	buf := make([]byte, t.Length)
	donePieces := 0

	for donePieces < len(t.PieceHashes) {
		res := <-results

		copy(buf[res.index*t.PieceLength:], res.buf)
		donePieces++

		percent := float64(donePieces) / float64(len(t.PieceHashes)) * 100
		fmt.Printf("(%0.2f%%) Downloaded piece #%d from swarm\n", percent, res.index)
	}
	close(workQueue)
	return buf
}

func (t *Torrent) startWorker(peer torrentfile.Peer, workQueue chan *pieceWork, results chan *pieceResult) {
	address := fmt.Sprintf("%s:%d", peer.IP, peer.Port)

	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return
	}
	defer conn.Close()

	req := handshake.Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: t.InfoHash,
		PeerID:   t.PeerID,
	}

	_, err = conn.Write(req.Serialize())
	if err != nil {
		return
	}

	res, err := handshake.Read(conn)
	if err != nil || !bytes.Equal(res.InfoHash[:], t.InfoHash[:]) {
		return
	}

	interestedMsg := &message.Message{ID: message.MsgInterested}
	conn.Write(interestedMsg.Serialize())

	for {
		conn.SetDeadline(time.Now().Add(10 * time.Second))
		msg, err := message.Read(conn)
		if err != nil {
			return
		}
		if msg == nil {
			continue
		}
		if msg.ID == message.MsgUnchoke {
			break
		}
	}

	for pw := range workQueue {
		pieceBuf, err := downloadPiece(conn, pw)
		if err != nil {
			workQueue <- pw
			return
		}

		hash := sha1.Sum(pieceBuf)
		if !bytes.Equal(hash[:], pw.hash[:]) {
			fmt.Printf("Piece #%d failed integrity check! Dropping it.\n", pw.index)
			workQueue <- pw
			continue
		}

		results <- &pieceResult{index: pw.index, buf: pieceBuf}
	}
}

func downloadPiece(conn net.Conn, pw *pieceWork) ([]byte, error) {
	buf := make([]byte, pw.length)
	const blockSize = 16384
	downloaded := 0

	for downloaded < pw.length {
		length := blockSize
		if pw.length-downloaded < blockSize {
			length = pw.length - downloaded
		}

		reqMsg := message.FormatRequest(pw.index, downloaded, length)
		_, err := conn.Write(reqMsg.Serialize())
		if err != nil {
			return nil, err
		}

		for {
			conn.SetDeadline(time.Now().Add(15 * time.Second))
			msg, err := message.Read(conn)
			if err != nil {
				return nil, err
			}
			if msg == nil {
				continue
			}
			if msg.ID == message.MsgPiece {
				block, err := message.ParsePiece(pw.index, msg.Payload)
				if err != nil {
					return nil, err
				}
				copy(buf[downloaded:], block)
				downloaded += len(block)
				break
			}
		}
	}
	return buf, nil
}

func (t *Torrent) calculatePieceSize(index int) int {
	begin := index * t.PieceLength
	end := begin + t.PieceLength
	if end > t.Length {
		return t.Length - begin
	}
	return t.PieceLength
}
