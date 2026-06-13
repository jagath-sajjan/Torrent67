```bash
Torrent67 on  main [!] via 🐹 v1.25.3 took 7s
❯ go run main.go debian.torrent
--- Torrent Parsed Successfully ---
Tracker URL: http://bttracker.debian.org:6969/announce
File name:   debian-13.5.0-arm64-netinst.iso
Info Hash:   3f04ac6b9d14cb7341faff5f8cbc30d565bac416
Scraping tracker: http://bttracker.debian.org:6969/announce
Found 40 peers!

=========================================
Dialing TCP connection to 77.172.167.173:10121...
Connected! Sending BitTorrent Handshake...
Handshake successful! Peer ID: 2d7142353231302d7a637579364d7338436b415f
Sending 'Interested' message...
Listening for peer messages...

Received Bitfield! (Payload size: 351 bytes)

SUCCESS! Peer unchoked us!

--- TIME TO DOWNLOAD PIECES ---
Requested the first 16384 bytes. Waiting for delivery...
Ignoring Unchoke message while waiting for our data...

NOM NOM NOM! Successfully downloaded 16384 bytes of Ubuntu!

 SUCCESS! Saved the raw data to 'ubuntu_block_0.dat'
 ```
