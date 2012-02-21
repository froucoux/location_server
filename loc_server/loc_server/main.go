package main

import (
	"flag"
	"location_server/logutil"
	"location_server/loc_server/server"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"websocket"
)

const logPath = "/var/log/locserver/server.log"

var iFile []byte
var minTreeMax *int64 = flag.Int64("treeSize", 1000, "The initialisation size of the quadtree")
var trackMovement *bool = flag.Bool("m", false, "Broadcast fine grained movement of users")
var threads *int = flag.Int("t", 1, "The number of threads available to the runtime")

func init() {
	flag.Parse()
	runtime.GOMAXPROCS(*threads)
}

func main() {
	logutil.LogFree("Location Server Started")
	http.Handle("/loc", websocket.Handler(locserver.WebsocketUser))
	go locserver.TreeManager(*minTreeMax, *trackMovement)
	http.ListenAndServe(":8002", nil)
}
