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

type comment struct {
	Id         int       `json:"id"`
	Body       string    `json:"body"`
	BodySample string    `json:"-"`
	TableName  string    `json:"table_name"`
	ParentId   int       `json:"parent_id"`
	PostId     int       `json:"post_id"`
	UserId     int       `json:"user_id"`
	UserName   string    `json:"user_name"`
	Replies    int       `json:"replies"`
	Depth      int8      `json:"depth"`
	Likes      int       `json:"likes"`
	Liked      bool      `json:"liked"`
	Created    time.Time `json:"created"`
}

func ShowThread(w http.ResponseWriter, r *http.Request) {
	// userid, ok := auth(w, r)
	// if !ok {
	// 	return
	// }

	// err := assure(r, "thread_id")
	// if err != nil {
	// 	jsonError(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }
	// id := strings.Split(r.URL.Path, "/")[2]

	templates.Execute(w, "thread", nil)
}

func NewComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	userid, ok := auth(w, r)
	if !ok {
		return
	}

	// TODO: limit body
	err := assure(r, "body", "parent_id")
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var table string
	var postId string
	var parentId int
	var parentUserId int
	var c comment
	if strings.Contains(r.FormValue("parent_id"), "p:") {
		postId = strings.Split(r.FormValue("parent_id"), ":")[1]
		err = data.Query(`select table_name, user_id, body_sample from posts where id = $1`, postId).Rows(&c)
		table = c.TableName
		parentUserId = c.UserId
	} else {
		err = data.Query(`select table_name, post_id, user_id, body_sample from comments where id = $1`, r.FormValue("parent_id")).Rows(&c)
		postId = strconv.Itoa(c.PostId)
		table = c.TableName
		parentUserId = c.UserId
		parentId, _ = strconv.Atoi(r.FormValue("parent_id"))
	}
	if table == "" {
		jsonError(w, "Table doesn't exist", http.StatusInternalServerError)
		return
	}
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var role string
	err = data.Query(`select role from memberships where table_name = $1 and user_id = $2`, table, userid).Rows(&role)
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if role == "reader" {
		jsonError(w, "you do not have enogh privileges to comment on this table", http.StatusUnauthorized)
		return
	}

	var userName string
	err = data.Query(`select name from users where id = $1`, userid).Rows(&userName)
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	body := r.FormValue("body")
	body = linkfy(body)
	// allow line breaks, reduce to one
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
	var id int
	if strings.Contains(r.FormValue("parent_id"), "p:") {
		err = data.Query(`insert into comments(post_id, body, body_sample, table_name, user_id, user_name, created)values($1, $2, $3, $4, $5, $6, $7) returning id`, postId, body, bodySample, table, userid, userName, created).Rows(&id)
		go data.Query(`update posts set comments = comments + 1 where id = $1`, postId).Run()
	} else {
		err = data.Query(`insert into comments(parent_id, post_id, body, body_sample, table_name, user_id, user_name, created)values($1, $2, $3, $4, $5, $6, $7, $8) returning id`, r.FormValue("parent_id"), postId, body, bodySample, table, userid, userName, created).Rows(&id)
		go data.Query(`update comments set replies = replies + 1 where id = $1`, parentId).Run()
	}
	if err != nil {
		log.Println(err)
		// todo: check if error, parent doesn't exist
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	notificationType := "r"
	if strings.Contains(r.FormValue("parent_id"), "p:") {
		notificationType = "c"
	}

	notificationBody := "<strong>" + userName + "</strong>: " + r.FormValue("body")
	if len(notificationBody) > 78 {
		notificationBody = notificationBody[:78] + "..."
	}

	// TODO: notify users in same level, and don't notify if same user
	go data.Notify(parentUserId, c.BodySample, notificationBody, "", notificationType, table)

	userId, _ := strconv.Atoi(userid)
	postIdInt, _ := strconv.Atoi(postId)
	json(w, &comment{Id: id, Body: body, UserId: userId, UserName: userName, TableName: table, ParentId: parentId, PostId: postIdInt, Created: created})
}

func GetComments(w http.ResponseWriter, r *http.Request) {
	err := assure(r, "post_id")
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	comments := []comment{}
	if r.FormValue("thread_id") != "" {
		// ensure numeric
		_, err = strconv.Atoi(r.FormValue("thread_id"))
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = data.Query(`WITH RECURSIVE cte (id, body, user_id, user_name, table_name, parent_id, replies, likes, depth, created, path)  AS (
		    SELECT  id,
		        body,
		        user_id,
		        user_name,
		        table_name,
		        parent_id,
		        replies,
		        likes,
		        1 AS depth,
		        created,
		        array[id] AS path
		    FROM    comments
		    WHERE   id = $1

		    UNION ALL

		    SELECT  comments.id,
		        comments.body,
		        comments.user_id,
		        comments.user_name,
		        comments.table_name,
		        comments.parent_id,
		        comments.replies,
		        comments.likes,
		        cte.depth + 1 AS depth,
		        comments.created,
		        cte.path || comments.id
		    FROM    comments
		    join cte on comments.parent_id = cte.id
		    )
		    SELECT id, body, user_id, user_name, table_name, parent_id, replies, likes, depth, created FROM cte
		ORDER BY path;`, r.FormValue("thread_id")).Rows(&comments)
	} else {
		err = data.Query(`
		WITH RECURSIVE cte (id, body, user_id, user_name, table_name, parent_id, replies, likes, depth, created, path)  AS (
		    SELECT  id,
		        body,
		        user_id,
		        user_name,
		        table_name,
		        parent_id,
		        replies,
		        likes,
		        1 AS depth,
		        created,
		        array[id] AS path
		    FROM    comments
		    WHERE   parent_id is null AND post_id = $1

		    UNION ALL

		    SELECT  comments.id,
		        comments.body,
		        comments.user_id,
		        comments.user_name,
		        comments.table_name,
		        comments.parent_id,
		        comments.replies,
		        comments.likes,
		        cte.depth + 1 AS depth,
		        comments.created,
		        cte.path || comments.id
		    FROM    comments
		    JOIN cte ON comments.parent_id = cte.id AND post_id = $2
		    )
		    SELECT id, body, user_id, user_name, table_name, parent_id, replies, likes, depth, created FROM cte
		ORDER BY path;`, r.FormValue("post_id"), r.FormValue("post_id")).Rows(&comments)
	}

	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json(w, &comments)
}

func GetCommentsFlat(w http.ResponseWriter, r *http.Request) {
	err := assure(r, "post_id")
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	comments := []comment{}
	if r.FormValue("thread_id") != "" {
		err = data.Query(`SELECT id, body, user_id, user_name, table_name, parent_id, replies, likes, created FROM comments where id = $1`, r.FormValue("thread_id")).Rows(&comments)
	} else {
		err = data.Query(`SELECT id, body, user_id, user_name, table_name, parent_id, replies, likes, created FROM comments where post_id = $1 and parent_id is null order by created asc`, r.FormValue("post_id")).Rows(&comments)
	}

	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json(w, &comments)
}

func GetReplies(w http.ResponseWriter, r *http.Request) {
	err := assure(r, "parent_id")
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	comments := []comment{}
	err = data.Query(`SELECT id, body, user_id, user_name, table_name, parent_id, replies, likes, created FROM comments where parent_id = $1 order by created asc`, r.FormValue("parent_id")).Rows(&comments)
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json(w, &comments)
}
