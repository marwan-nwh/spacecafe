package handlers

// import (
// 	"fmt"
// 	"net/http"
// 	"spacecafe/data"
// 	"spacecafe/pkg/gravity"
// 	"spacecafe/templates"
// )

// func NewStage(w http.ResponseWriter, r *http.Request) {
// 	if r.Method == "GET" {
// 		templates.Execute(w, "stage_new", nil)
// 		return
// 	}

// 	err := assure(r, "url")
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	feed, err := gravity.Parse(r.FormValue("url"))
// 	if err != nil {
// 		http.Error(w, "Feeds URL is not valid", http.StatusBadRequest)
// 		return
// 	}

// 	id := gravity.MD5(r.FormValue("url"))
// 	err = data.Query(`insert into stages(id, title, description, url) values($1, $2, $3, $4) on conflict do nothing`, id, feed.Title, feed.Description, r.FormValue("url")).Run()
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	gravity.Pull(r.FormValue("url"))

// 	fmt.Fprintf(w, "Stage created successfuly. Id: %s, Title: %s", id, feed.Title)
// }

// TODO: use transactions
// func New(url string) (string, error) {
// 	feed, err := Fetch(url)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	err = db.Query(`insert into feeds(id, name, url)
// 		values ($1, $2, $3) on conflict do nothing`, id, feed.Title, url).Run()
// 	if err != nil {
// 		return "", err
// 	}
// 	return id, nil
// }

// func Show(w http.ResponseWriter, r *http.Request) {

// }

// stage -> listeners
// user -> feed(500)

//
// redis lists, heavey work on redis
// postgres array,
// middle table, deleteing last records?

// create stage
// check if feed is unique
// fetch links
// save links

// when to update
// what to do after update - add to users feed - update users unread counts
