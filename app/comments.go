package app

import (
	"net/http"
	"spacecafe/db"
	"time"
)

var Comments = []handler{
	handler{
		Action: "Like Comment",
		Method: "POST",
		URI:    "/comments/like",
		Auth:   true,
		Params: []string{"comment_id"},

		// we don't have to assure membership because it is not a big damage
		// if the user is not a member it would be added just one time. so we are
		// trading one query with the possiblity of adding one like to the item, good deal
		Do: func(w http.ResponseWriter, r *http.Request) {
			user := who(r)

			// save like
			err := db.Query(`insert into comments_likes(user_id, comment_id, created)values($1, $2, $3) on conflict(comment_id, user_id) do nothing`, user.Id, r.FormValue("comment_id"), time.Now()).Run()
			if err != nil {
				panic(err)
			}

			// increase comment likes and ignore errors
			go db.Query(`update comments set likes = likes + 1 where id = $1`, r.FormValue("comment_id")).Run()
		}},
	handler{
		Action: "Unlike Comment",
		Method: "POST",
		URI:    "/comments/unlike",
		Auth:   true,
		Params: []string{"comment_id"},

		// we don't have to check if the user already liked the comment
		// because we already ignoring the errors, and checking that won't save us anything
		Do: func(w http.ResponseWriter, r *http.Request) {
			user := who(r)

			// delete like and ignore errors
			db.Query(`delete from comments_likes where comment_id = $1 and user_id = $2`, r.FormValue("comment_id"), user.Id).Run()

			// decrease comment likes and ignore errors
			go db.Query(`update comments set likes = likes - 1 where id = $1`, r.FormValue("comment_id")).Run()

		}}}

// const Posts = []handler{
// 	handler{
// 		Action: "New Post",
// 		Method: "POST",
// 		URI:    "/posts",
// 		Auth:   true,
// 		Params: []string{"body:max=160,eq=real,anonymous", "table", "identity"},
// 		Do: func(w http.ResponseWriter, r *http.Request) {
// 			user := who(r)

// 			// Get member role
// 			var role string
// 			err = data.Query(`select role from memberships where table_name = $1 and user_id = $2`, r.FormValue("table"), user.Id).Rows(&role)
// 			if err != nil {
// 				panic(err)
// 			}

// 			if role != "admin" && role != "writer" && role != "manager" {
// 				http.Error(w, "you do not have enough privileges to post on this table", http.StatusUnauthorized)
// 				return
// 			}

// 			post := &Post{
// 				body:     r.FormValue("body"),
// 				UserName: user.NameByIdentity(r.FormValue("identity")),
// 				UserId: user.Id,
// 			}
// 			err = post.Save()
// 			if err != nil {
// 				http.Error(w, err, http.StatusBadRequest)
// 				return
// 			}

// 			JSON(w, post)
// 		}
// 	}
// }

// const Tables = []handler{
// 	handler{
// 		Action: "Show HomePage",
// 		Method: "GET",
// 		URI:    "/",
// 		Auth:   true,
// 		Do: func(w http.ResponseWriter, r *http.Request) {
// 			user := who(r)

// 			// get user tables
// 			var tables []Table
// 			err := db.Query(`select name, type, description, role, members, posts, active_rate
// 			  from tables t join memberships m on t.name = m.table_name where m.user_id = $1;`, user.Id).Rows(&tables)
// 			if err != nil {
// 				panic(err)
// 			}

// 			// get user name and anonumous image
// 			err = db.Query(`select name, nimg from users where id = $1`, user.Id).Rows(&user)
// 			if err != nil {
// 				panic(err)
// 			}

// 			data := struct {
// 				Tables []Table
// 				User   *User
// 			}{
// 				Tables: tables,
// 				User:   &user,
// 			}
// 			Temp(w, "home", &data)
// 		}
// 	}
// }

// const Posts = []handler.Handler{
// 	handler.Handler{
// 		Action: "Create New Post",
// 		Method: "POST",
// 		URI:    "/posts",
// 		Auth:   true,
// 		Params: []string{"body:max=140", "table", "identity"},
// 		Do: func(w http.ResponseWriter, r *http.Request) {
// 			member := members.Member{UserId: r.Userid, Table: r.FormValue("table")}
// 			if !member.CanPost() {
// 				return errors.New("you do not have enough privileges to post on this table")
// 			}
// 			post := posts.Post{
// 				body:     r.FormValue("body"),
// 				UserName: user.Name(r.Userid, r.FormValue("identity")),
// 			}
// 			err = post.Save()
// 			if err != nil {
// 				return err
// 			}
// 			return json(&p)
// 		}},
// 	handler.Handler{
// 		Action: "Show Post",
// 		Desc:   "Return "
// 		Method: "GET",
// 		URI:    "/posts/:id",
// 		Do: func(r *http.Request) http.Response {
// 			// TODO: check table privacy
// 			post, err := posts.Get(r.FormValue("post_id"))
// 			if err != nil {
// 				return err
// 			}
// 			postJson, err := util.ToJSON(post)
// 			if err != nil {
// 				return err
// 			}
// 			data := struct {
// 				Table    string
// 				PostJson *string
// 			}{
// 				Table:    p.TableName,
// 				PostJson: &postJson,
// 			}
// 			return template("post", &data)
// 		}},
// 	handler.Handler{
// 		Action: "Unlike Post",
// 		Method: "POST",
// 		URI:    "/posts/unlike",
// 		Params: "post_id",
// 		Do: func(r *http.Request) http.Response {
// 			err := likes.Delete("where user_id = $1 and post_id = $2", r.Userid, r.FormValue("post_id"))
// 			if err != nil {
// 				return err
// 			}
// 			post := posts.Post{r.FormValue("post_id")}
// 			go post.DecreaseLikes()
// 		}}}

// posts.Get("*", "table = $1 and ___", ...)
// posts.Delete("created_at < ___", ...)
// posts.Save([]post)

// post.Get("*")
// post.Save()
// post.Update()
// post.Delete()
// post.IsVerified()
