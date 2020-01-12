package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/abradley2/svelte-metro/api"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

func playPingPong(c *websocket.Conn, id string) {
	for {
		if _, _, err := c.NextReader(); err != nil {
			api.RemoveListener(id)
			c.Close()
			break
		}
	}
}

type handler struct{}

var lastPayload []byte

func createSocketConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	c := make(chan api.Result)
	i := uuid.New().String()
	api.AddListener(i, &c)
	go playPingPong(conn, i)
	for {
		conn.WriteMessage(websocket.TextMessage, lastPayload)
		r := <-c
		if r.Err == nil {
			lastPayload = r.Payload
			conn.WriteMessage(websocket.TextMessage, r.Payload)
		}
	}
}

var apiHandler http.Handler

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path

	// only hand off to the mux router with the api prefix
	if strings.Contains(p, "api/") {
		apiHandler.ServeHTTP(w, r)
		return
	}
	if p == "/ws" {
		createSocketConnection(w, r)
		return
	}

	isAbs := path.IsAbs(r.URL.Path)

	if isAbs {
		var contentType string
		var fp string
		wd, _ := os.Getwd()

		fp = path.Join(wd, "./public", p)

		file, fileErr := os.Open(fp)
		if fileErr != nil {
			fp = path.Join(wd, "./public/index.html")
		} else {
			file.Close()
		}

		var data []byte
		var err error
		data, err = ioutil.ReadFile(fp)

		if err != nil {
			fp = path.Join(wd, "./public/index.html")
			data, err = ioutil.ReadFile(fp)
		}

		if err == nil {
			if strings.HasSuffix(p, ".css") {
				contentType = "text/css"
			} else if strings.HasSuffix(p, ".html") {
				contentType = "text/html"
			} else if strings.HasSuffix(p, ".js") {
				contentType = "application/javascript"
			} else if strings.HasSuffix(p, ".png") {
				contentType = "image/png"
			} else if strings.HasSuffix(p, ".svg") {
				contentType = "image/svg+xml"
			} else {
				contentType = "text/html"
			}

			w.Header().Add("Content-Type", contentType)
			w.Write(data)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNotFound)
	return
}

func main() {
	go api.PollMetro()

	h := new(handler)

	apiHandler = cors.Default().Handler(api.Router)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", api.PORT),
		Handler: h,
	}

	server.ListenAndServe()
}
