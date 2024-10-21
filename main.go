package main
import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)
type Artist struct {
	ID            int                 `json: "id"`
	Image         string              `json:"image"`
	Name          string              `json:"name"`
	Members       []string            `json:"members"`
	CreationDate  int                 `json:"creationDate"`
	FirstAlbum    string              `json:"firstAlbum"`
	DateLocations map[string][]string `json:"datesLocations"`
}
type Relation struct {
	DateLocations map[string][]string `json:"datesLocations"`
}
func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		artists, err := GetAllArtistsFromAPI("https://groupietrackers.herokuapp.com/api/artists")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Printf("Error getting artists from API: %v", err)
			return
		}
		RenderArtistsPage(w, artists)
	})
	http.HandleFunc("/artist/", artistHandler)
	log.Println("Server listening on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
func RenderArtistsPage(w http.ResponseWriter, artists []Artist) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Template parsing error: %v", err)
		return
	}
	pageData := struct {
		Title   string
		Artists []Artist
	}{
		Title:   "Artist List",
		Artists: artists,
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, pageData); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Template rendering error: %v", err)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}
func GetAllArtistsFromAPI(url string) ([]Artist, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var artists []Artist
	err = json.Unmarshal(body, &artists)
	if err != nil {
		return nil, err
	}
	return artists, nil
}
func artistHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	idInt, _ := strconv.Atoi(id)
	if idInt < 1 || idInt > 52 {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("ID , %s not Found", id)
		return
	}
	response, err := http.Get(fmt.Sprintf("https://groupietrackers.herokuapp.com/api/artists/%s", id))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error getting artist from API: %v", err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		http.Error(w, "Not Found", http.StatusNotFound)
		log.Printf("Artist with ID %s not found", id)
		return
	}
	var artist Artist
	err = json.NewDecoder(response.Body).Decode(&artist)
	if err != nil {
		http.Error(w, "Unexpected Error Occurred", http.StatusInternalServerError)
		log.Printf("Error decoding artist: %v", err)
		return
	}
	res, err := http.Get(fmt.Sprintf("https://groupietrackers.herokuapp.com/api/relation/%s", id))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error getting relations from API: %v", err)
		return
	}
	defer res.Body.Close()
	if response.StatusCode != http.StatusOK {
		http.Error(w, "Not Found", http.StatusNotFound)
		log.Printf("Artist with ID %s not found", id)
		return
	}
	var relations Relation
	err = json.NewDecoder(res.Body).Decode(&relations)
	if err != nil {
		http.Error(w, "Unexpected Error Occurred", http.StatusInternalServerError)
		log.Printf("Error decoding relations: %v", err)
		return
	}
	artist.DateLocations = relations.DateLocations
	RenderArtistPage(w, artist)
}
func RenderArtistPage(w http.ResponseWriter, artist Artist) {
	tmpl, err := template.ParseFiles("templates/artists.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Template parsing error: %v", err)
		return
	}
	pageData := struct {
		Title  string
		Artist Artist
	}{
		Title:  "Artist Info",
		Artist: artist,
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, pageData); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Template rendering error: %v", err)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}