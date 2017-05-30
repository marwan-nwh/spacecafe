package handlers

import (
	"github.com/russross/blackfriday"
	"html"
	"log"
	"net/http"
	"spacecafe/data"
	"spacecafe/pkg/gravity"
	"spacecafe/templates"
	"strconv"
	"strings"
	"time"
)

type membership struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

func Explore(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		templates.Execute(w, "explore", nil)
		return
	}

	userid, ok := auth(w, r)
	if !ok {
		return
	}

	var err error
	err = assure(r, "sort", "page", "request_id")
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var tables []table
	page, err := strconv.Atoi(r.FormValue("page"))
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	sort := strings.ToLower(r.FormValue("sort"))

	switch sort {
	case "top":
		err = data.Query(`
			select t.name, t.description, t.members, t.created,
			exists(select 1 from memberships m where m.table_name = t.name and m.user_id = $1) as is_member
			from tables t limit 5 offset $2
			`, userid, 10*page).Rows(&tables)
	case "latest":
		err = data.Query(`
			select t.name, t.description, t.members, t.created,
			exists(select 1 from memberships m where m.table_name = t.name and m.user_id = $1) as is_member
			from tables t order by t.created desc limit 5 offset $2
			`, userid, 10*page).Rows(&tables)
	}
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := struct {
		RequestId string  `json:"request_id"`
		Tables    []table `json:"tables"`
	}{
		RequestId: r.FormValue("request_id"),
		Tables:    tables,
	}
	json(w, &response)
}

func ToggleMembership(w http.ResponseWriter, r *http.Request) {
	userid, ok := softAuth(w, r)
	if !ok {
		return
	}
	err := assure(r, "table")
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	// TODO: if admins downgraded user to reader, the user shouldn't be able to just leave the table and join again
	// roles could be moved to another table, or just make a table with downgraded users and don't delete rows on leaving
	// and check it before joining
	var role string
	err = data.Query(`select role from memberships where table_name = $1 and user_id = $2`, r.FormValue("table"), userid).Rows(&role)
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if role == "manager" {
		jsonError(w, "You are the Manager of this table, transfer your role to another member before leaving.", 406)
		return
	}

	if role != "" {
		err = data.Query(`delete from memberships where table_name = $1 and user_id = $2`, r.FormValue("table"), userid).Run()
		if err != nil {
			log.Println(err)
			return
		}
	}

	if role == "" {
		err = data.Query(`insert into memberships(table_name, user_id, role)values($1, $2, $3) on conflict (table_name,user_id) do nothing`, r.FormValue("table"), userid, "writer").Run()
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func GetAdmins(w http.ResponseWriter, r *http.Request) {
	_, ok := softAuth(w, r)
	if !ok {
		return
	}

	err := assure(r, "table")
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var admins []membership
	err = data.Query(`select id, name, role from admins_view where table_name = $1`, r.FormValue("table")).Rows(&admins)
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json(w, &admins)
}

func AddAdmin(w http.ResponseWriter, r *http.Request) {
	userid, ok := auth(w, r)
	if !ok {
		return
	}

	var role string
	err := data.Query(`select role from memberships where table_name = $1 and user_id = $2`, r.FormValue("table"), userid).Rows(&role)
	if err != nil {
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// only managers can add admins, or transfer thier role
	if role != "manager" {
		jsonError(w, "You are not authorized to do this action.", http.StatusNotAcceptable)
		return
	}

	var usr user
	idInt, err := strconv.Atoi(r.FormValue("id"))
	if err == nil {
		usr.Id = idInt
		err = data.Query(`select name from users where id = $1`, idInt).Rows(&usr)
		if err != nil {
			log.Println(err)
			jsonError(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		err = data.Query(`select id, name from users where email = $1`, r.FormValue("email")).Rows(&usr)
		if err != nil {
			log.Println(err)
			jsonError(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	if userid == strconv.Itoa(usr.Id) {
		jsonError(w, "You can not do this action to "+usr.Name, http.StatusNotAcceptable)
		return
	}

	var rl string
	err = data.Query(`select role from memberships where table_name = $1 and user_id = $2`, r.FormValue("table"), usr.Id).Rows(&rl)
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rl == "admin" && r.FormValue("type") == "admin" {
		jsonError(w, usr.Name+" is already an admin.", http.StatusNotAcceptable)
		return
	}

	// admin
	// manager
}

type homeUser struct {
	Id   string `json:"-"`
	Name string `json:"-"`
	Nimg string `json:"-"`
}

func Home(w http.ResponseWriter, r *http.Request) {
	userid, ok := auth(w, r)
	if !ok {
		return
	}

	tables := []table{}
	usr := homeUser{Id: userid}
	err := data.Query(`select name, type, description, role, members, posts, active_rate
	  from tables t join memberships m on t.name = m.table_name where m.user_id = $1;`, userid).Rows(&tables)
	if err != nil {
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = data.Query(`select name, nimg from users where id = $1`, userid).Rows(&usr)
	if err != nil {
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Tables []table
		User   *homeUser
	}{
		Tables: tables,
		User:   &usr,
	}
	templates.Execute(w, "home", &data)
}

type table struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	About       string    `json:"about"`
	AboutMd     string    `json:"about_md"`
	Type        string    `json:"type"`
	OwnerId     int       `json:"-"`
	Members     int       `json:"members"`
	Posts       int       `json:"posts"`
	Comments    int       `json:"comments"`
	Role        string    `json:"role"`
	Likes       int       `json:"likes"`
	ActiveRate  float64   `json:"-"`
	Feeds       int       `json:"feeds"`
	IsMember    bool      `json:"is_member"`
	Create      time.Time `json:"created"`
}

// TODO: add all possible words
var reserved = []string{"overview", "new", "my table", "mytable", "your table", "yourtable", "stage", "stages", "new tables", "tables", "friends", "followers", "admin", "admins", "favorites", "favorite users"}

func isReserved(name string) bool {
	for _, r := range reserved {
		if r == name {
			return true
		}
	}
	return false
}

func NewTable(w http.ResponseWriter, r *http.Request) {
	userid, ok := auth(w, r)
	if !ok {
		return
	}
	if r.Method == "GET" {
		templates.Execute(w, "table_new", nil)
		return
	}

	err := assure(r, "name:max=20,min=3", "description:max=100", "about:max=2000,empty")
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// if r.FormValue("type") != "public" && r.FormValue("type") != "anonymous" {
	// 	jsonError(w, "Wrong Type", http.StatusBadRequest)
	// 	return
	// }

	name := strings.ToLower(r.FormValue("name"))

	if isReserved(name) {
		jsonError(w, "sorry this name is reserved by spacecafe", http.StatusBadRequest)
		return
	}
	alphapet := "abcdefghijklmnopqrstuvwxyz"
	numbers := "0123456789"
	nameLetters := strings.Split(name, "")
	for _, letter := range nameLetters {
		if letter != " " && !strings.Contains(alphapet, letter) && !strings.Contains(numbers, letter) {
			jsonError(w, "table name: only use lowercase letters, numbers and spaces", http.StatusBadRequest)
			return
		}
	}

	description := html.EscapeString(r.FormValue("description"))

	// description := strictSanitizer.Sanitize(r.FormValue("description"))

	aboutSanitizer := strictSanitizer
	aboutSanitizer.AllowStandardAttributes()
	aboutSanitizer.RequireParseableURLs(true)
	aboutSanitizer.AllowAttrs("href", "target", "rel").OnElements("a")
	aboutSanitizer.AddTargetBlankToFullyQualifiedLinks(true)
	aboutSanitizer.AllowStandardURLs()
	aboutSanitizer.AllowRelativeURLs(false)
	aboutSanitizer.AllowElements("p", "strong", "h1", "a", "ul", "li", "br")
	unsafeHTML := blackfriday.MarkdownBasic([]byte(r.FormValue("about")))
	// log.Println(string(unsafeHTML))
	about := aboutSanitizer.Sanitize(string(unsafeHTML))

	// table := &table{
	// 	Type:  r.FormValue("type"),
	// 	Name:  name,
	// 	Desc:  description,
	// 	About: about,
	// }
	// json(w, &table)

	// save in database

	name = strings.TrimSpace(name)

	// allow one space between words
	segments := strings.Split(name, " ")
	nms := []string{}
	for _, segment := range segments {
		if segment != " " {
			nms = append(nms, segment)
		}
	}
	name = strings.Join(nms, "_")

	tx, err := data.Begin()
	if err != nil {
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	created := time.Now()
	err = tx.Query(`insert into tables(name, description, about, about_md, type, owner_id, created)values($1, $2, $3, $4, $5, $6, $7)`, name, description, about, r.FormValue("about"), "public", userid, created).Run()
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			jsonError(w, "Table <a href='/t/"+name+"'>"+name+"</a> already exist", http.StatusInternalServerError)
			return
		}
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tx.Query(`insert into memberships(table_name, user_id, role)values($1, $2, $3) on conflict (table_name,user_id) do nothing`, name, userid, "manager").Run()
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	jsonError(w, name, http.StatusOK) // not error
	return
}

func UpdateTable(w http.ResponseWriter, r *http.Request) {
	userid, ok := auth(w, r)
	if !ok {
		return
	}

	err := assure(r, "table", "description:max=100", "about:max=2000,empty")
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: check admins not just the creator
	var ownerId int
	err = data.Query(`select owner_id from tables where name = $1`, r.FormValue("table")).Rows(&ownerId)
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if strconv.Itoa(ownerId) != userid {
		jsonError(w, "", http.StatusUnauthorized)
		return
	}

	description := html.EscapeString(r.FormValue("description"))

	aboutSanitizer := strictSanitizer
	aboutSanitizer.AllowStandardAttributes()
	aboutSanitizer.RequireParseableURLs(true)
	aboutSanitizer.AllowAttrs("href", "target", "rel").OnElements("a")
	aboutSanitizer.AddTargetBlankToFullyQualifiedLinks(true)
	aboutSanitizer.AllowStandardURLs()
	aboutSanitizer.AllowRelativeURLs(false)
	aboutSanitizer.AllowElements("p", "strong", "h1", "a", "ul", "li", "br")
	unsafeHTML := blackfriday.MarkdownBasic([]byte(r.FormValue("about")))
	about := aboutSanitizer.Sanitize(string(unsafeHTML))

	err = data.Query(`update tables set description = $1, about = $2, about_md = $3 where name = $4`,
		description, about, r.FormValue("about"), r.FormValue("table")).Run()
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json(w, &table{Description: description, About: about})
}

func ShowTable(w http.ResponseWriter, r *http.Request) {
	// extract table name from the url
	name := strings.Split(r.URL.Path, "/")[2]

	// do softAuth to make tables visible to visitors
	userid, _ := softAuth(w, r)

	tbl := table{}

	// get table data
	err := data.Query(`select name, type, description, about, about_md, members, created from tables where name = $1`, name).Rows(&tbl)
	if err != nil {
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// get user role
	role := "notmember"
	if userid != "" {
		err := data.Query(`select role from memberships where table_name = $1 and user_id = $2`, name, userid).Rows(&role)
		if err != nil {
			jsonError(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		tbl.Role = role
	} else {
		tbl.Role = "visitor"
	}

	templates.Execute(w, "table", &tbl)
}

type feed struct {
	Id        string `json:"id"`
	Title     string `json:"title"`
	Url       string `json:"url"`
	LastTitle string `json:"-"`
}

// TODO: add custom name and image
func AddFeed(w http.ResponseWriter, r *http.Request) {
	err := assure(r, "url", "table")
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	userid, ok := auth(w, r)
	if !ok {
		return
	}

	// TODO: check admins not just the creator
	var ownerId int
	err = data.Query(`select owner_id from tables where name = $1`, r.FormValue("table")).Rows(&ownerId)
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if strconv.Itoa(ownerId) != userid {
		jsonError(w, "", http.StatusUnauthorized)
		return
	}

	fd, err := gravity.Parse(r.FormValue("url"))
	if err != nil {
		jsonError(w, "Something wrong with the URl or the source server, please try agin", http.StatusBadRequest)
		return
	}

	if fd.Title == "" {
		fd.Title = "News Source"
	}

	id := gravity.MD5(r.FormValue("url"))

	var lastTitle string
	err = data.Query(`select last_title from feeds where id = $1`, id).Rows(&lastTitle)
	if err != nil && len(fd.Items) > 0 {
		lastTitle = fd.Items[0].Title
	}

	tx, err := data.Begin()
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tx.Query(`insert into feeds(id, title, last_title, url)values($1, $2, $3, $4) on conflict (id) do nothing`, id, fd.Title, lastTitle, r.FormValue("url")).Run()
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tx.Query(`insert into listening(feed_id, table_name, created)values($1, $2, $3) on conflict (feed_id, table_name) do nothing`, id, r.FormValue("table"), time.Now()).Run()
	if err != nil {
		log.Println(err)
		tx.Rollback()
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		tx.Rollback()
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	gravity.Pull(id, r.FormValue("url"), lastTitle)

	go data.Query(`update feeds set listeners = listeners + 1 where id = $1`, id).Run()
	go data.Query(`update tables set feeds = feeds + 1 where name = $1`, r.FormValue("table")).Run()

	json(w, &feed{
		Id:    id,
		Title: fd.Title,
		Url:   r.FormValue("url"),
	})
}

func GetFeeds(w http.ResponseWriter, r *http.Request) {
	err := assure(r, "table")
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var feeds []feed
	err = data.Query(`select id, title, url from feeds f join listening l on f.id = l.feed_id and l.table_name = $1`, r.FormValue("table")).Rows(&feeds)
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// log.Println(feeds[0])
	json(w, &feeds)
}

func DeleteFeed(w http.ResponseWriter, r *http.Request) {
	err := assure(r, "table", "feed_id")
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	userid, ok := auth(w, r)
	if !ok {
		return
	}

	// TODO: check admins not just the creator
	var ownerId int
	err = data.Query(`select owner_id from tables where name = $1`, r.FormValue("table")).Rows(&ownerId)
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if strconv.Itoa(ownerId) != userid {
		jsonError(w, "", http.StatusUnauthorized)
		return
	}

	err = data.Query(`delete from listening where feed_id = $1 and table_name = $2`, r.FormValue("feed_id"), r.FormValue("table")).Run()
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	go data.Query(`update tables set feeds = feeds - 1 where name = $1`, r.FormValue("table")).Run()

	var listeners int
	err = data.Query(`update feeds set listeners = listeners - 1 where id = $1 returning listeners`, r.FormValue("feed_id")).Rows(&listeners)
	if err == nil && listeners == 0 {
		// remove feed if no listeners left
		data.Query(`delete from feeds where feed_id = $1`, r.FormValue("feed_id")).Run()
		gravity.Kill(r.FormValue("feed_id"))
	}

	jsonError(w, "done", http.StatusOK)
}

func LoadFeeds() error {
	var feeds []feed
	err := data.Query(`select id, url, last_title from feeds`).Rows(&feeds)
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		gravity.Pull(feed.Id, feed.Url, feed.LastTitle)
	}
	return nil
}

// func LeaveTable(w http.ResponseWriter, r *http.Request) {
// 	userid, ok := auth(w, r)
// 	if !ok {
// 		return
// 	}

// 	err := assure(r, "table")
// 	if err != nil {
// 		jsonError(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	err = data.Query(`delete from membership where user_id = $1 and table_name = $2`, userid, r.FormValue("table")).Run()

// 	// err = data.Query(`insert into memberships(table_name, user_id, role)values($1, $2, $3) on conflict (table_name,user_id) do nothing`, name, userid, "admin").Run()
// 	if err != nil {
// 		log.Println(err)
// 		jsonError(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}

// 	jsonError(w, "done", http.StatusOK)
// }
