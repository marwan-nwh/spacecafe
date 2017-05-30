package handlers

import (
	"log"
	"net/http"
	"spacecafe/data"
	"spacecafe/templates"
	"strconv"
	"strings"
	"time"
)

type post struct {
	Id           int       `json:"id"`
	Body         string    `json:"body"`
	TableName    string    `json:"table_name"`
	UserId       int       `json:"user_id"`
	SecretUserId int       `json:"-"`
	UserName     string    `json:"user_name"`
	Nimg         string    `json:"nimg"`
	Comments     int       `json:"comments"`
	Likes        int       `json:"likes"`
	Liked        bool      `json:"liked"`
	Rank         int       `json:"rank"`
	Created      time.Time `json:"created"`
}

// action
// data
// response

func NewPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	userid, ok := auth(w, r)
	if !ok {
		return
	}

	err := assure(r, "body", "table", "identity")
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var role string
	err = data.Query(`select role from memberships where table_name = $1 and user_id = $2`, r.FormValue("table"), userid).Rows(&role)
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if role != "admin" && role != "writer" && role != "manager" {
		jsonError(w, "you do not have enogh privileges to post on this table", http.StatusUnauthorized)
		return
	}

	if r.FormValue("identity") != "real" && r.FormValue("identity") != "anonymous" {
		jsonError(w, "", http.StatusBadRequest)
		return
	}

	usr := user{}
	if r.FormValue("identity") == "real" {
		err = data.Query(`select name from users where id = $1`, userid).Rows(&usr)
	} else {
		err = data.Query(`select nname as name, nimg from users where id = $1`, userid).Rows(&usr)
	}

	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	body := r.FormValue("body")
	body = linkfy(body)

	// test length after linkfing to avoid calculating it with regexp as in the client side
	if len(body) > 160 {
		jsonError(w, "Post can't be more than 160 char.", http.StatusBadRequest)
		return
	}
	bodySanitizer := strictSanitizer
	bodySanitizer.AllowStandardAttributes()
	bodySanitizer.RequireParseableURLs(true)
	bodySanitizer.AllowAttrs("href", "target", "rel").OnElements("a")
	bodySanitizer.AddTargetBlankToFullyQualifiedLinks(true)
	bodySanitizer.AllowStandardURLs()
	bodySanitizer.AllowRelativeURLs(false)
	bodySanitizer.AllowElements("p", "a")
	body = bodySanitizer.Sanitize(body)

	var bodySample string
	if len(r.FormValue("body")) > 69 {
		bodySample = r.FormValue("body")[:69] + "..."
	} else {
		bodySample = r.FormValue("body")
	}

	created := time.Now()
	secretUserId := "0"
	nimg := ""
	if r.FormValue("identity") == "anonymous" {
		secretUserId = userid
		userid = "0"
		nimg = usr.Nimg
	}

	var id int
	err = data.Query(`insert into posts(body, body_sample, table_name, user_id, secret_user_id, user_name, nimg, created, type, rank)values($1, $2, $3, $4, $5, $6, $7, $8, 'p', ceiling(extract(epoch from age($9, timestamp '2016-11-01'))) ) returning id`, body, bodySample, r.FormValue("table"), userid, secretUserId, usr.Name, nimg, created, created).Rows(&id)
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	useridInt, _ := strconv.Atoi(userid)
	json(w, &post{Id: id, Body: body, UserId: useridInt, UserName: usr.Name, Nimg: nimg, Likes: 0, Comments: 0, TableName: r.FormValue("table"), Created: created})
}

// TODO: add request_id
func GetPosts(w http.ResponseWriter, r *http.Request) {
	// UploadImg()
	// if r.FormValue("table") == "stage" {
	// 	userid, ok := auth(w, r)
	// 	if !ok {
	// 		return
	// 	}
	// 	posts := []post{}
	// 	err := data.Query(`select *, exists(select 1 FROM posts_likes where post_id = p.id and user_id = $1) as "liked"
	// 	from posts p order by rank desc`, userid).Rows(&posts)
	// 	if err != nil {
	// 		log.Println(err)
	// 		jsonError(w, "Internal server error", http.StatusInternalServerError)
	// 		return
	// 	}

	// 	json(w, &posts)
	// 	return
	// }
	var err error

	err = assure(r, "table", "sorting", "last_item_id", "page")
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	userid, ok := softAuth(w, r)
	if ok {
		data.CheckNotificationsUnread(w, userid)
	}

	_, err = strconv.Atoi(r.FormValue("last_item_id"))
	if err != nil {
		if r.FormValue("last_item_id") == "Infinity" {
			// TODO
		} else {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	posts := []post{}
	if userid != "" {
		if r.FormValue("sorting") == "latest" {
			err = data.Query(`select *, exists(select 1 FROM posts_likes where post_id = p.id and user_id = $1) as "liked"
	 		from posts p where table_name = $2 order by id desc limit 5 offset 5 * $3`, userid, r.FormValue("table"), r.FormValue("page")).Rows(&posts)
		} else {
			err = data.Query(`select *, exists(select 1 FROM posts_likes where post_id = p.id and user_id = $1) as "liked"
	 		from posts p where table_name = $2 order by rank desc limit 5 offset 5 * $3`, userid, r.FormValue("table"), r.FormValue("page")).Rows(&posts)
		}
	} else {
		// TODO: return only the right data
		if r.FormValue("sorting") == "latest" {
			err = data.Query(`select * from posts p where table_name = $1 order by id desc limit 5`, r.FormValue("table")).Rows(&posts)
		} else {
			err = data.Query(`select * from posts p where table_name = $1 order by rank desc limit 5`, r.FormValue("table")).Rows(&posts)
		}
	}
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// response := struct {
	// 	Posts []post `json:"posts"`
	// }{Posts: posts}
	json(w, &posts)
}

func GetPost(w http.ResponseWriter, r *http.Request) {
	// TODO: check table privacy
	postId := strings.Split(r.URL.Path, "/")[2]

	var p post
	err := data.Query(`select * from posts p where id = $1`, postId).Rows(&p)
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	postJson, err := toJson(p)
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	data := struct {
		Table    string
		PostJson *string
	}{
		Table:    p.TableName,
		PostJson: &postJson,
	}
	templates.Execute(w, "post", &data)
}

// we don't have to assure membership because it is not a big damage
// if some userid that doesn't belong to the table added to the set
// and it won't be dublicated
func LikePost(w http.ResponseWriter, r *http.Request) {
	userid, ok := auth(w, r)
	if !ok {
		return
	}

	err := assure(r, "post_id")
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	created := time.Now()

	err = data.Query(`insert into posts_likes(user_id, post_id, created)values($1, $2, $3) on conflict(post_id, user_id) do nothing`, userid, r.FormValue("post_id"), created).Run()
	if err != nil {
		log.Println(err)
	}
	// TODO: increase post likes, table likes, active rate
}

func UnlikePost(w http.ResponseWriter, r *http.Request) {
	userid, ok := auth(w, r)
	if !ok {
		return
	}

	err := assure(r, "post_id")
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = data.Query(`delete from posts_likes where user_id = $1 and post_id = $2`, userid, r.FormValue("post_id")).Run()
	// log.Println(err)
	if err != nil {
	}
	// TODO: decrease post likes, table likes
}

// update posts set rank = ceiling(extract(epoch from age(created, timestamp '2016-11-01')))  + (120 * comments) + (60 * likes) ;UPDATE 315
// select rank, ceiling(extract(epoch from age(created, timestamp '2016-11-01'))) as init_rank,likes,comments,id, ceiling(((2 *comments)+likes)/(ceiling(extract(epoch from age(now(), created)))/3600)^1.5) as HN,ceiling(extract(epoch from age('2016-11-28 20:00:00+02', created)))/3600  ,left(body, 35) as body from posts where table_name = 'spacecafe' order by rank desc;
