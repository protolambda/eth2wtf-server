package main

import (
	"eth2wtf-server/server"
	"eth2wtf-server/world/headers"
	bh "github.com/protolambda/zrnt/eth2/beacon/header"
	. "eth2wtf-server/common/contenttyp"
	"flag"
	"log"
	"net/http"
	"os"
)

var addr = flag.String("addr", ":4000", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func main() {
	flag.Parse()

	s := server.NewServer()

	hp := &headers.HeadersProducer{
		Headers: make(chan *bh.BeaconBlockHeader, 10),
		Logger:  log.New(os.Stdout, "header producer: ", log.LUTC | log.Lshortfile),
		Closed:  false,
	}
	go hp.Produce(s.World().ChunkGetter(HeaderType), s.GetViewers)
	// enable clients to connect
	go s.Run()

	httpServer := http.NewServeMux()
	// route to http handlers. There's a home page and a websocket entry.
	httpServer.HandleFunc("/", serveHome)
	httpServer.HandleFunc("/ws", s.ServeWs)

	// accept connections
	if err := http.ListenAndServe(*addr, httpServer); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
