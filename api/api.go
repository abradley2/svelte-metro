package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

// Router _
var Router *mux.Router

var client *http.Client

var apiKey string

// PORT _
var PORT int

// cache
var memCache = make(map[string]*[]byte)

var cacheMutex = &sync.Mutex{}
var listenerMutex = &sync.Mutex{}

const (
	stationURL = "https://api.wmata.com/Rail.svc/json/jStations"
	linesURL   = "https://api.wmata.com/Rail.svc/json/jLines"
)

func cleanCache() {
	for {
		c := time.After(60 * time.Second)

		_ = <-c

		cacheMutex.Lock()
		delete(memCache, stationURL)
		delete(memCache, linesURL)
		cacheMutex.Unlock()
	}
}

// Result _
type Result struct {
	Err     error
	Payload []byte
}

// Listener _
type Listener struct {
	ID      string
	Channel *chan Result
}

var listeners []*Listener

// AddListener _
func AddListener(id string, c *chan Result) {
	l := Listener{
		ID:      id,
		Channel: c,
	}
	listenerMutex.Lock()
	listeners = append(listeners, &l)
	listenerMutex.Unlock()
}

// RemoveListener _
func RemoveListener(id string) {
	listenerMutex.Lock()
	var filtered []*Listener
	for i := 0; i < len(listeners); i++ {
		l := listeners[i]
		if l.ID != id {
			filtered = append(filtered, l)
		}
	}
	listeners = filtered
	listenerMutex.Unlock()
}

func updateListeners(r Result) {
	for i := 0; i < len(listeners); i++ {
		l := listeners[i]
		*l.Channel <- r
	}
}

func init() {
	wd, _ := os.Getwd()
	viper.SetConfigName("config")
	viper.AddConfigPath(wd)
	viper.AutomaticEnv()
	viper.ReadInConfig()

	apiKey = viper.GetString("API_KEY")
	PORT = viper.GetInt("PORT")

	client = &http.Client{Timeout: 10 * time.Second}

	Router = mux.NewRouter()

	Router.HandleFunc("/api/stations", stations)
	Router.HandleFunc("/api/lines", lines)

	go cleanCache()
}

// PollMetro _
func PollMetro() {
	for {
		c := time.After(15 * time.Second)

		_ = <-c

		url := "https://api.wmata.com/StationPrediction.svc/json/GetPrediction/All"

		r, _ := http.NewRequest("GET", url, nil)
		r.Header.Set("api_key", apiKey)
		resp, err := client.Do(r)

		if err != nil {
			fmt.Println(err)
			updateListeners(Result{Err: err})
			return
		}

		body, _ := ioutil.ReadAll(resp.Body)

		updateListeners(Result{Err: nil, Payload: body})
	}
}

func stations(w http.ResponseWriter, r *http.Request) {
	// first check if this is cached
	cacheMutex.Lock()
	if cached, ok := memCache[stationURL]; ok {
		w.WriteHeader(200)
		w.Write(*cached)
		cacheMutex.Unlock()
		return
	}
	cacheMutex.Unlock()

	req, _ := http.NewRequest("GET", stationURL, nil)
	req.Header.Set("api_key", apiKey)

	resp, err := client.Do(req)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)

	cacheMutex.Lock()
	memCache[stationURL] = &body
	cacheMutex.Unlock()

	w.WriteHeader(resp.StatusCode)
	w.Write(body)

	defer resp.Body.Close()
}

func lines(w http.ResponseWriter, r *http.Request) {

	// first check if this is cached
	cacheMutex.Lock()
	if cached, ok := memCache[linesURL]; ok {
		w.WriteHeader(200)
		w.Write(*cached)
		cacheMutex.Unlock()
		return
	}
	cacheMutex.Unlock()

	req, _ := http.NewRequest("GET", linesURL, nil)
	req.Header.Set("api_key", apiKey)

	resp, err := client.Do(req)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)

	cacheMutex.Lock()
	memCache[linesURL] = &body
	cacheMutex.Unlock()

	w.WriteHeader(resp.StatusCode)
	w.Write(body)

	defer resp.Body.Close()
}
