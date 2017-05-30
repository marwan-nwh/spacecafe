package data

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"flag"
	"fmt"

	"github.com/eaigner/jet"
	"github.com/garyburd/redigo/redis"
	_ "github.com/lib/pq"
	// "github.com/microcosm-cc/bluemonday"
	"html"
	"log"
	"net/http"
	"os"
	"spacecafe/pkg/gravity"
	"strconv"
	"strings"
	"time"
)

// old credentials
const (
	pass = "3vSTX7uXaQXnQszdX33PYrmHcYsunyeKnYAK7C8Wt4EvQdERwRKEzs9E92UbtrEHF5PMvzcKW3KtPnULJMGwL9HkYAZyCWPFgTqcXtNgspDfgps6tBSV9yPyspfeaGxm"
)

// TODO: enable sslmode
var driver = "user=spacecafe host=spacecafe-test.c1woibcsxv8q.us-west-2.rds.amazonaws.com password='" + pass + "' dbname=spacecafe port=5432"

var db *jet.Db

func Init() {
	openDB()
	openRedis()
}

func openDB() {
	// TODO: set max idel and open conn
	var err error
	if os.Getenv("SC_DB_SOURCE") != "" {
		db, err = jet.Open("postgres", os.Getenv("SC_DB_SOURCE"))
	} else {
		db, err = jet.Open("postgres", driver)
	}
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
}

func Query(query string, args ...interface{}) jet.Runnable {
	return db.Query(query, args...)
}

func QueryRow(query string, args ...interface{}) *sql.Row {
	return db.QueryRow(query, args...)
}

type tx struct {
	jtx *jet.Tx
}

func Begin() (tx, error) {
	jtx, err := db.Begin()
	return tx{jtx}, err
}

func (tx tx) Query(query string, args ...interface{}) jet.Runnable {
	return tx.jtx.Query(query, args...)
}

func (tx tx) Commit() error {
	return tx.jtx.Commit()
}

func (tx tx) Rollback() error {
	return tx.jtx.Rollback()
}

// var (
// 	strictSanitizer = bluemonday.StrictPolicy()
// )

func replaceurls(url string) string {
	text := strings.TrimPrefix(url, "https:")
	text = strings.TrimPrefix(text, "http:")
	text = strings.TrimPrefix(text, "//")
	text = strings.TrimPrefix(text, "www.")
	if len(text) > 27 {
		text = text[:27] + "..."

	}
	return "<a href='" + url + "' target='_blank' rel='nofollow noopener'>" + text + "</a>"
}

func Listen(news gravity.News) {
	go func() {
		for {
			select {
			case feed := <-news:
				var tables []string
				// bulk insert
				// http://stackoverflow.com/questions/758945/whats-the-fastest-way-to-do-a-bulk-insert-into-postgres
				// http://stackoverflow.com/questions/21108084/golang-mysql-insert-multiple-data-at-once

				// feed.Link is feed.Id, gravity do this because gofeed.Feed doesn't have and id
				err := Query(`select table_name from listening where feed_id = $1`, feed.Link).Rows(&tables)
				if err == nil {
					// feed.Title = strictSanitizer.Sanitize(feed.Title)
					feed.Title = html.EscapeString(feed.Title)
					for _, table := range tables {
						for _, item := range feed.Items {
							item.Title = html.EscapeString(item.Title)
							// TODO: add limits to body and username
							if len(item.Title) > 150 {
								item.Title = item.Title[:147] + "..."
							}
							body := fmt.Sprintf("%s %s", item.Title, replaceurls(item.Link))
							Query(`insert into posts(body, table_name, user_id, user_name, created, type)values($1, $2, $3, $4, $5, "f")`, body, table, 0, feed.Title, item.PublishedParsed).Run()
						}
					}
				}
				Query(`update feeds set last_title = $1 where id = $2`, feed.Items[0].Title, feed.Link).Run()
			}
		}
	}()
}

var (
	redispool *redis.Pool
)

func newRedisPool(server, password string) *redis.Pool {
	// TODO: set max, pass
	maxIdle := 5
	maxActive := 10
	if password == "" {
		return &redis.Pool{
			MaxIdle:     maxIdle,
			MaxActive:   maxActive,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", server)
				if err != nil {
					return nil, err
				}
				// TODO
				// if _, err := c.Do("AUTH", password); err != nil {
				// 	c.Close()
				// 	return nil, err
				// }
				return c, err
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		}
	}
	return &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			// TODO
			if _, err := c.Do("AUTH", password); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func openRedis() {
	redisServer := flag.String("redisServer", ":6379", "")
	redisPassword := flag.String("redisPassword", "", "")
	redispool = newRedisPool(*redisServer, *redisPassword)
	redisconn := redispool.Get()
	defer redisconn.Close()
	_, err := redisconn.Do("PING")
	if err != nil {
		panic(err)
	}
}

func NewSession(id string) (string, error) {
	now := time.Now().Unix()
	rnd := random(64)
	conn := redispool.Get()
	defer conn.Close()
	key := fmt.Sprintf("s:%s:%d", id, now)
	_, err := conn.Do("set", key, rnd, "ex", "2592000") // 2592000 = 30*24*60*60 = 30 days
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%d-%s", id, now, rnd), nil
}

func GetSession(r *http.Request) (string, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return "", err
	}
	values := strings.Split(cookie.Value, "-")
	if len(values) != 3 {
		return "", errors.New("token should be 3 segments")
	}
	conn := redispool.Get()
	defer conn.Close()
	random, err := redis.String(conn.Do("get", fmt.Sprintf("s:%s:%s", values[0], values[1])))
	if err != nil {
		return "", err
	}
	if random != values[2] {
		// TODO: send email to yourself
		return "", errors.New("possible hacking") // user trying to act as another user
	}
	return values[0], nil
}

func EndSession(r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return
	}
	values := strings.Split(cookie.Value, "-")
	if len(values) != 3 {
		return
	}
	conn := redispool.Get()
	defer conn.Close()
	conn.Do("del", fmt.Sprintf("s:%s:%s", values[0], values[1]))
}

func random(strenght int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, strenght)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func InitNotifications(id string) {
	conn := redispool.Get()
	defer conn.Close()
	key := fmt.Sprintf("n%s", id)
	_, err := conn.Do("set", key, 0)
	if err != nil {
		log.Println(err)
	}
}

func CheckNotificationsUnread(w http.ResponseWriter, id string) int {
	conn := redispool.Get()
	defer conn.Close()
	unread, err := redis.Int(conn.Do("get", fmt.Sprintf("n%s", id)))
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Notifications", strconv.Itoa(unread))
	return unread
	// 	conn.Do("set", fmt.Sprintf("s:%s:%s", values[0], values[1]), random, "ex", "2592000")
}

func Notify(userid int, title, body, url, typ, table string) {
	now := time.Now()
	err := Query(`insert into notifications (user_id,title, body, url, type, table_name, created) values($1, $2, $3, $4, $5, $6, $7)`, userid, title, body, url, typ, table, now).Run()
	if err != nil {
		log.Println(err)
		return
	}
	conn := redispool.Get()
	defer conn.Close()
	key := fmt.Sprintf("n%d", userid)
	log.Println(userid)
	_, err = conn.Do("incr", key)
	if err != nil {
		log.Println(err)
	}
}

// func CreatePostLikes() {
// 	conn := redispool.Get()
// 	defer conn.Close()
// 	_, err := conn.Do("sadd", "p"+postId, userId)
// 	if err != nil {

// 	}
// }

// func LikePost(postId, userId string) {
// 	conn := redispool.Get()
// 	defer conn.Close()
// 	exists, err := redis.Bool(conn.Do("Exists", "p"+postId))
// 	if exists {
// 		conn.Do("sadd", "p"+postId, userId)
// 	}

// 	// count, err := redis.Int(conn.Do("SCARD", "p"+postId))
// 	// if err != nil {
// 	// 	return 0, err
// 	// }
// 	// if count == 1 {
// 	// 	conn.Do("expire", "p"+postId, 15*24*60*60)
// 	// }
// 	// return count, nil
// }

// func UnlikePost(postId, userId string) {
// 	conn := redispool.Get()
// 	defer conn.Close()
// 	conn.Do("srem", "p"+postId, userId)
// }
