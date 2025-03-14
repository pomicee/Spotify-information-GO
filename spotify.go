package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

type TrackResponse struct {
	Success bool      `json:"success"`
	Track   TrackInfo `json:"track"`
}

type TrackInfo struct {
	Name       string `json:"name"`
	FullTitle  string `json:"fullTitle"`
	ID         string `json:"id"`
	URL        string `json:"url"`
	PreviewURL string `json:"preview_url"`
	Duration   string `json:"duration"`
	DurationMs int    `json:"duration_ms"`
	Explicit   bool   `json:"explicit"`
	Popularity int    `json:"popularity"`
}

type ArtistShortResponse struct {
	Success bool       `json:"success"`
	Artist  ArtistInfo `json:"artist"`
}

type ArtistInfo struct {
    Name         string   `json:"name"`
    ID           string   `json:"id"`
    URL          string   `json:"url"`
    Image        string   `json:"image"`
    Genres       []string `json:"genres"`
    Followers    int      `json:"followers"`
    Popularity   int      `json:"popularity"`
    // MonthlyListeners is not available through Spotify's public API
    MonthlyListeners int      `json:"monthlyListeners,omitempty"`
    Albums           int      `json:"albums"`
    Singles         int      `json:"singles"`
    Compilations    int      `json:"compilations"`
}

type ArtistFullResponse struct {
	Success bool             `json:"success"`
	Artist  ArtistFullInfo  `json:"artist"`
}

type ArtistFullInfo struct {
	Name      string           `json:"name"`
	TopTracks []TopTrackInfo  `json:"topTracks"`
	Albums    []AlbumBasicInfo `json:"albums"`
	AlbumStats AlbumStats      `json:"albumStats"`
}

type TopTrackInfo struct {
	Name       string `json:"name"`
	Popularity int    `json:"popularity"`
}

type AlbumBasicInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type AlbumStats struct {
	Album        int `json:"album"`
	Single       int `json:"single"`
	Compilation  int `json:"compilation"`
}

type AlbumResponse struct {
	Success bool       `json:"success"`
	Album   AlbumInfo `json:"album"`
}

type AlbumInfo struct {
	Name        string        `json:"name"`
	Artists     []ArtistBasic `json:"artists"`
	ReleaseDate string        `json:"releaseDate"`
	Genres      []string      `json:"genres"`
	TotalTracks int           `json:"totalTracks"`
	Popularity  int           `json:"popularity"`
	Type        string        `json:"type"`
	URL         string        `json:"url"`
	Images      []ImageInfo   `json:"images"`
	Tracks      []TrackBasic  `json:"tracks"`
}

type ArtistBasic struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	URL  string `json:"url"`
}

type ImageInfo struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type TrackBasic struct {
	Name        string `json:"name"`
	Duration    int    `json:"duration"`
	TrackNumber int    `json:"trackNumber"`
	URL         string `json:"url"`
}

type SpotifyClient struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	TokenType    string
	ExpiresAt    time.Time
	HTTPClient   *http.Client
}

func NewSpotifyClient(clientID, clientSecret string) *SpotifyClient {
	return &SpotifyClient{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		HTTPClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *SpotifyClient) authenticate() error {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(c.ClientID + ":" + c.ClientSecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	c.AccessToken = tokenResp.AccessToken
	c.TokenType = tokenResp.TokenType
	c.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

func (c *SpotifyClient) ensureValidToken() error {
	if c.AccessToken == "" || time.Now().After(c.ExpiresAt) {
		return c.authenticate()
	}
	return nil
}

func (c *SpotifyClient) makeRequest(method, endpoint string) ([]byte, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, "https://api.spotify.com/v1"+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func handleSpotifySongs(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing query parameter 'q'", http.StatusBadRequest)
		return
	}

	client := NewSpotifyClient(clientID, clientSecret)
	
	// Search for tracks
	data, err := client.makeRequest("GET", "/search?q="+url.QueryEscape(query)+"&type=track&limit=1")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var searchResult map[string]interface{}
	if err := json.Unmarshal(data, &searchResult); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tracks := searchResult["tracks"].(map[string]interface{})
	items := tracks["items"].([]interface{})
	if len(items) == 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "No tracks found",
		})
		return
	}

	track := items[0].(map[string]interface{})
	response := TrackResponse{
		Success: true,
		Track: TrackInfo{
			Name:       track["name"].(string),
			ID:         track["id"].(string),
			URL:        track["external_urls"].(map[string]interface{})["spotify"].(string),
			Duration:   formatDuration(int(track["duration_ms"].(float64))),
			DurationMs: int(track["duration_ms"].(float64)),
			Popularity: int(track["popularity"].(float64)),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleArtistShort(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing query parameter 'q'", http.StatusBadRequest)
		return
	}

	client := NewSpotifyClient(clientID, clientSecret)
	
	// Search for artist
	data, err := client.makeRequest("GET", "/search?q="+url.QueryEscape(query)+"&type=artist&limit=1")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var searchResult map[string]interface{}
	if err := json.Unmarshal(data, &searchResult); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	artists := searchResult["artists"].(map[string]interface{})
	items := artists["items"].([]interface{})
	if len(items) == 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "No artist found",
		})
		return
	}

	artist := items[0].(map[string]interface{})
	
	albumsData, err := client.makeRequest("GET", "/artists/"+artist["id"].(string)+"/albums")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var albumsResult map[string]interface{}
	if err := json.Unmarshal(albumsData, &albumsResult); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var albums, singles, compilations int
	for _, item := range albumsResult["items"].([]interface{}) {
		albumType := item.(map[string]interface{})["album_type"].(string)
		switch albumType {
		case "album":
			albums++
		case "single":
			singles++
		case "compilation":
			compilations++
		}
	}

	response := ArtistShortResponse{
		Success: true,
		Artist: ArtistInfo{
			Name:       artist["name"].(string),
			ID:         artist["id"].(string),
			URL:        artist["external_urls"].(map[string]interface{})["spotify"].(string),
			Image:      getArtistImage(artist),
			Genres:     getStringSlice(artist["genres"].([]interface{})),
			Followers:  int(artist["followers"].(map[string]interface{})["total"].(float64)),
			Popularity: int(artist["popularity"].(float64)),
			Albums:     albums,
			Singles:    singles,
			Compilations: compilations,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleArtistFull(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing query parameter 'q'", http.StatusBadRequest)
		return
	}

	client := NewSpotifyClient(clientID, clientSecret)
	
	data, err := client.makeRequest("GET", "/search?q="+url.QueryEscape(query)+"&type=artist&limit=1")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var searchResult map[string]interface{}
	if err := json.Unmarshal(data, &searchResult); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	artists := searchResult["artists"].(map[string]interface{})
	items := artists["items"].([]interface{})
	if len(items) == 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "No artist found",
		})
		return
	}

	artist := items[0].(map[string]interface{})
	artistID := artist["id"].(string)

	tracksData, err := client.makeRequest("GET", "/artists/"+artistID+"/top-tracks?market=US")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var tracksResult map[string]interface{}
	if err := json.Unmarshal(tracksData, &tracksResult); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	albumsData, err := client.makeRequest("GET", "/artists/"+artistID+"/albums")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var albumsResult map[string]interface{}
	if err := json.Unmarshal(albumsData, &albumsResult); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := ArtistFullResponse{
		Success: true,
		Artist: ArtistFullInfo{
			Name:      artist["name"].(string),
			TopTracks: getTopTracks(tracksResult["tracks"].([]interface{})),
			Albums:    getAlbums(albumsResult["items"].([]interface{})),
			AlbumStats: getAlbumStats(albumsResult["items"].([]interface{})),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleAlbum(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing query parameter 'q'", http.StatusBadRequest)
		return
	}

	client := NewSpotifyClient(clientID, clientSecret)
	
	data, err := client.makeRequest("GET", "/search?q="+url.QueryEscape(query)+"&type=album&limit=1")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var searchResult map[string]interface{}
	if err := json.Unmarshal(data, &searchResult); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	albums := searchResult["albums"].(map[string]interface{})
	items := albums["items"].([]interface{})
	if len(items) == 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "No album found",
		})
		return
	}

	album := items[0].(map[string]interface{})
	albumID := album["id"].(string)

	albumData, err := client.makeRequest("GET", "/albums/"+albumID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var albumResult map[string]interface{}
	if err := json.Unmarshal(albumData, &albumResult); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := AlbumResponse{
		Success: true,
		Album: AlbumInfo{
			Name:        albumResult["name"].(string),
			Artists:     getArtists(albumResult["artists"].([]interface{})),
			ReleaseDate: albumResult["release_date"].(string),
			TotalTracks: int(albumResult["total_tracks"].(float64)),
			Popularity:  int(albumResult["popularity"].(float64)),
			Type:        albumResult["album_type"].(string),
			URL:         albumResult["external_urls"].(map[string]interface{})["spotify"].(string),
			Images:      getImages(albumResult["images"].([]interface{})),
			Tracks:      getTracks(albumResult["tracks"].(map[string]interface{})["items"].([]interface{})),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func formatDuration(ms int) string {
	seconds := ms / 1000
	minutes := seconds / 60
	seconds = seconds % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

func getArtistImage(artist map[string]interface{}) string {
	images := artist["images"].([]interface{})
	if len(images) > 0 {
		return images[0].(map[string]interface{})["url"].(string)
	}
	return ""
}

func getStringSlice(items []interface{}) []string {
	result := make([]string, len(items))
	for i, item := range items {
		result[i] = item.(string)
	}
	return result
}

func getTopTracks(tracks []interface{}) []TopTrackInfo {
	result := make([]TopTrackInfo, len(tracks))
	for i, track := range tracks {
		t := track.(map[string]interface{})
		result[i] = TopTrackInfo{
			Name:       t["name"].(string),
			Popularity: int(t["popularity"].(float64)),
		}
	}
	return result
}

func getAlbums(albums []interface{}) []AlbumBasicInfo {
	result := make([]AlbumBasicInfo, len(albums))
	for i, album := range albums {
		a := album.(map[string]interface{})
		result[i] = AlbumBasicInfo{
			Name: a["name"].(string),
			Type: a["album_type"].(string),
		}
	}
	return result
}

func getAlbumStats(albums []interface{}) AlbumStats {
	var stats AlbumStats
	for _, album := range albums {
		a := album.(map[string]interface{})
		switch a["album_type"].(string) {
		case "album":
			stats.Album++
		case "single":
			stats.Single++
		case "compilation":
			stats.Compilation++
		}
	}
	return stats
}

func getArtists(artists []interface{}) []ArtistBasic {
	result := make([]ArtistBasic, len(artists))
	for i, artist := range artists {
		a := artist.(map[string]interface{})
		result[i] = ArtistBasic{
			Name: a["name"].(string),
			ID:   a["id"].(string),
			URL:  a["external_urls"].(map[string]interface{})["spotify"].(string),
		}
	}
	return result
}

func getImages(images []interface{}) []ImageInfo {
	result := make([]ImageInfo, len(images))
	for i, image := range images {
		img := image.(map[string]interface{})
		result[i] = ImageInfo{
			URL:    img["url"].(string),
			Height: int(img["height"].(float64)),
			Width:  int(img["width"].(float64)),
		}
	}
	return result
}

func getTracks(tracks []interface{}) []TrackBasic {
	result := make([]TrackBasic, len(tracks))
	for i, track := range tracks {
		t := track.(map[string]interface{})
		result[i] = TrackBasic{
			Name:        t["name"].(string),
			Duration:    int(t["duration_ms"].(float64)),
			TrackNumber: int(t["track_number"].(float64)),
			URL:         t["external_urls"].(map[string]interface{})["spotify"].(string),
		}
	}
	return result
}

var (
	clientID     = ""
	clientSecret = ""
)

func main() {
	http.HandleFunc("/spotify/songs", handleSpotifySongs)
	http.HandleFunc("/spotify/artist/short", handleArtistShort)
	http.HandleFunc("/spotify/artist/full", handleArtistFull)
	http.HandleFunc("/spotify/album", handleAlbum)

	fmt.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
