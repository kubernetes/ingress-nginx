package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

var podName string
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Ignore http origin
	},
}

func init() {
	podName = os.Getenv("podname")
}

func ws(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request", r.RemoteAddr)
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("failed to upgrade:", err)
		return
	}
	defer c.Close()

	s := fmt.Sprintf("Connected to %v", podName)
	c.WriteMessage(websocket.TextMessage, []byte(s))
	handleWSConn(c)
}

func handleWSConn(c *websocket.Conn) {
	stop := make(chan struct{})
	go func() {
		for {
			time.Sleep(5 * time.Second)

			select {
			case <-stop:
				return
			default:
			}

			s := fmt.Sprintf("%s reports time: %v", podName, time.Now().String())
			c.WriteMessage(websocket.TextMessage, []byte(s))
		}
	}()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("Error while reading:", err)
			break
		}
		if err = c.WriteMessage(mt, message); err != nil {
			log.Println("Error while writing:", err)
			break
		}
	}
	close(stop)
}

func root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Write([]byte(`Websocket example. Connect to /ws`))
}

func main() {
	log.Println("Starting")
	http.HandleFunc("/ws", ws)
	http.HandleFunc("/", root)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
