// This example demonstrates how to authenticate with Spotify.
// In order to run this example yourself, you'll need to:
//
//  1. Register an application at: https://developer.spotify.com/my-applications/
//       - Use "http://localhost:8080/callback" as the redirect URI
//  2. Set the SPOTIFY_ID environment variable to the client ID you got in step 1.
//  3. Set the SPOTIFY_SECRET environment variable to the client secret from step 1.
package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/zmb3/spotify"
	"unicode/utf8"
)

// redirectURI is the OAuth redirect URI for the application.
// You must register an application at Spotify's developer portal
// and enter this value.
const redirectURI = "http://localhost:8080/callback"

var html = `
<style>
table, th, td {
    border: 1px solid black;
}
</style>
<br/>
<form>
<input type="text" name="input"><br/>
<input type=submit formmethod="get" formaction="/player/search"><br/>
</form>
<a href="/player/play">Play</a><br/>
<a href="/player/pause">Pause</a><br/>
<a href="/player/next">Next track</a><br/>
<a href="/player/previous">Previous Track</a><br/>
<a href="/player/shuffle">Shuffle</a><br/>
`

var (
	auth  = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserReadPlaybackState, spotify.ScopeUserModifyPlaybackState, spotify.ScopePlaylistModifyPrivate,spotify.ScopePlaylistModifyPublic)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

func main() {
	// We'll want these variables sooner rather than later
	var client *spotify.Client
	var playerState *spotify.PlayerState

	http.HandleFunc("/callback", completeAuth)

	http.HandleFunc("/player/", func(w http.ResponseWriter, r *http.Request) {
		action := strings.TrimPrefix(r.URL.Path, "/player/")
		s := action
		j := len(s)
		for i := 0; i < 36 && j > 0; i++ {
			_, size := utf8.DecodeLastRuneInString(s[:j])
			j -= size
		}
		lastByRune := s[j:]
		fmt.Println(lastByRune)
		if len(lastByRune)==36{
			action=strings.TrimSuffix(action,"/"+lastByRune)
		}
		fmt.Println("Got request for:", action)
		var err error
		var results *spotify.SearchResult
		switch action {
		case "search":
			querySearch:=strings.TrimLeft(r.URL.RawQuery, "?input")
			fmt.Println(querySearch)
			results,err=client.Search(querySearch,spotify.SearchTypeAlbum)
			if results.Albums != nil {
				fmt.Println("Album:")
				html = `
<style>
table, th, td {
    border: 1px solid black;
}
</style>
<br/>
<form>
<input type="text" name="input"><br/>
<input type=submit formmethod="get" formaction="/player/search"><br/>
</form>
<a href="/player/play">Play</a><br/>
<a href="/player/pause">Pause</a><br/>
<a href="/player/next">Next track</a><br/>
<a href="/player/previous">Previous Track</a><br/>
<a href="/player/shuffle">Shuffle</a><br/>
<table>
`
				for _, item := range results.Albums.Albums {
					html+="<tr><td><a href=/player/playsong/"+string(item.URI)+">"+string(item.Name)+"</a><br/></td></tr>"
					fmt.Println("   ", item.URI)
				}
				html+=`</table>`
			}
		case "playsong":
			user, e:=client.CurrentUser()
			if e != nil {
				log.Println(user)
				log.Print(e)
			}
			var opt spotify.PlayOptions
			var context spotify.PlaybackContext
			context.URI=spotify.URI(lastByRune)
			opt.PlaybackContext=&context.URI

			//code for individual tracks
			//Uris := []spotify.URI{"spotify:track:1GWIyLJLyza41v0TEdvUOG"}
			//opt.URIs=Uris

			client.PlayOpt(&opt)

			//code for adding tracks gto playlists
			//huh,err:=client.AddTracksToPlaylist(user.ID,spotify.ID("6FjNBaaDr9v7yxMonTsUvm"),spotify.ID(lastByRune))
			//log.Println(huh)
			//if err != nil {
			//	log.Println(playerState.Item.ID)
			//	log.Print(err)
			//}

		case "play":
			err = client.Play()
		case "pause":
			err = client.Pause()
		case "next":
			err = client.Next()
		case "previous":
			err = client.Previous()
		case "shuffle":
			playerState.ShuffleState = !playerState.ShuffleState
			err = client.Shuffle(playerState.ShuffleState)
		}
		if err != nil {
			log.Print(err)
		}



		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, html)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})

	go func() {
		//clientId and secretId used to Authenticate
		clientId:=""
		secretId:=""
		auth.SetAuthInfo(clientId,secretId)
		url := auth.AuthURL(state)
		fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

		// wait for auth to complete
		client = <-ch

		// use the client to make calls that require authorization
		user, err := client.CurrentUser()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("You are logged in as:", user.ID)

		playerState, err = client.PlayerState()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Found your %s (%s)\n", playerState.Device.Type, playerState.Device.Name)
	}()

	http.ListenAndServe(":8080", nil)

}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}
	// use the token to get an authenticated client
	client := auth.NewClient(tok)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "Login Completed!"+html)
	ch <- &client
}