package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"time"
	ws "golang.org/x/net/websocket"
)

func socketHandler(player *Player) func(*ws.Conn) {
	return func(conn *ws.Conn) {
		ch := make(chan string, 16)
		listenHandle := player.Listen(ch)
		defer close(ch)
		defer player.Unlisten(listenHandle)

		conn.SetDeadline(time.Now().Add(time.Hour * 42 * 42 * 42))

		for {
			_, err := conn.Write([]uint8(<-ch))
			if err != nil {
				break
			}
		}
	}
}

func htDataAttach(r *mux.Router, player *Player) {
	r.Path("/player/state").Methods("POST").HandlerFunc(htPlayerSetState(player))
	r.Path("/player/next").Methods("POST").HandlerFunc(htPlayerNext(player))
	r.Path("/track/current").Methods("GET").HandlerFunc(htPlayerCurrentTrack(player))
	r.Path("/track/playlist").Methods("GET").HandlerFunc(htPlayerPlaylist(player))
	r.Path("/track/browse{path:.*}").Methods("GET").HandlerFunc(htPlayerTracks(player))
	r.Path("/listen").Handler(ws.Handler(socketHandler(player)))
}

func htPlayerNext(player *Player) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if err := player.Next(); err != nil {
			panic(err)
		}

		if _, err := res.Write([]byte("{}")); err != nil {
			panic(err)
		}
	}
}

func htPlayerSetState(player *Player) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		var data struct {
			State string `json:"state"`
		}

		defer req.Body.Close()
		if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
			panic(err)
		}

		if err := player.SetState(data.State); err != nil {
			panic(err)
		}

		if _, err := res.Write([]byte("{}")); err != nil {
			panic(err)
		}
	}
}

func htPlayerCurrentTrack(player *Player) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if track, progress, err := player.CurrentTrack(); err != nil {
			panic(err)
		} else if state, err := player.State(); err != nil {
			panic(err)
		} else {
			err := json.NewEncoder(res).Encode(map[string]interface{}{
				"progress": progress,
				"track":    track,
				"state":    state,
			})
			if err != nil {
				panic(err)
			}
		}
	}
}

func htPlayerPlaylist(player *Player) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		tracks, err := player.Playlist()
		if err != nil {
			panic(err)
		}

		err = json.NewEncoder(res).Encode(map[string]interface{}{
			"tracks": tracks,
		})
		if err != nil {
			panic(err)
		}
	}
}

func htPlayerTracks(player *Player) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		tracks, err := player.ListTracks(mux.Vars(req)["path"])
		if err != nil {
			panic(err)
		}

		err = json.NewEncoder(res).Encode(map[string]interface{}{
			"tracks": tracks,
		})
		if err != nil {
			panic(err)
		}
	}
}