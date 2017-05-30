package main

import (
	"log"
	"net/http"
	"os"
	"spacecafe/app"
	"spacecafe/data"
	"spacecafe/db"
	"spacecafe/handlers"

	gorilla "github.com/gorilla/handlers"
	// "spacecafe/pkg/gravity"
	"spacecafe/pkg/sentry"
	"spacecafe/templates"
)

func main() {
	data.Init()
	db.Init()
	templates.Parse()
	// news := gravity.Init()
	// data.Listen(news)
	// loadFeeds()

	mux := http.NewServeMux()
	app.RegisterHandlers(mux)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/signup", handlers.Signup)
	mux.HandleFunc("/login", handlers.Login)
	mux.HandleFunc("/logout", handlers.Logout)
	mux.HandleFunc("/profile/complete", handlers.CompleteProfile)
	mux.HandleFunc("/users/images", handlers.CompleteProfile)
	mux.HandleFunc("/notifications/new", handlers.GetNewNotifications)
	mux.HandleFunc("/notifications", handlers.GetNotifications)
	mux.HandleFunc("/notifications/unread", handlers.GetNotificationsUnread)
	mux.HandleFunc("/notifications/read", handlers.MarkNotificationAsRead)
	// mux.HandleFunc("/stages/new", handlers.NewStage)
	mux.HandleFunc("/t/new", handlers.NewTable)
	mux.HandleFunc("/tables/feeds", handlers.GetFeeds)
	mux.HandleFunc("/tables/feeds/delete", handlers.DeleteFeed)
	mux.HandleFunc("/tables/feeds/new", handlers.AddFeed)
	mux.HandleFunc("/tables/admins", handlers.GetAdmins)
	mux.HandleFunc("/tables/admins/new", handlers.AddAdmin)
	mux.HandleFunc("/tables/settings", handlers.UpdateTable)
	mux.HandleFunc("/t/", handlers.ShowTable)
	mux.HandleFunc("/memberships/toggle", handlers.ToggleMembership)

	mux.HandleFunc("/explore", handlers.Explore)

	mux.HandleFunc("/posts/new", handlers.NewPost)
	mux.HandleFunc("/post/", handlers.GetPost)
	mux.HandleFunc("/posts", handlers.GetPosts)

	mux.HandleFunc("/posts/like", handlers.LikePost)
	mux.HandleFunc("/posts/unlike", handlers.UnlikePost)

	mux.HandleFunc("/comments/new", handlers.NewComment)
	mux.HandleFunc("/comments", handlers.GetComments)
	mux.HandleFunc("/comments/flat", handlers.GetCommentsFlat)
	mux.HandleFunc("/comments/replies", handlers.GetReplies)
	mux.HandleFunc("/thread/", handlers.ShowThread)
	mux.HandleFunc("/", handlers.Home)

	var err error
	if os.Getenv("SC_ENV") != "production" {
		err = http.ListenAndServe(":8000", mux)
	} else {
		err = http.ListenAndServe(":8000", gorilla.CompressHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mux.ServeHTTP(w, r)
		})))
	}
	if err != nil {
		panic(err)
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
}

func loadFeeds() {
	err := handlers.LoadFeeds()
	if err != nil {
		panic(err)
	}
}

// TODO: 404 pages
// TODO: handle panic on server
// TODO: add a layer to track user ip and how many requests they make per min, to prevent dos attacks
// TODO: HTTP compression
// TODO: SEO
// can't create more than 3 tables a month

// 25 tables because there will be alot of tables for just reading with not engagment
// not all of the users post on their tables, alot of them just read
