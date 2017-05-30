package handlers

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nfnt/resize"
	"github.com/satori/go.uuid"
	"github.com/vincent-petithory/dataurl"
	"golang.org/x/crypto/bcrypt"
	"html"
	"image/png"
	"log"
	"net/http"
	"os"
	"spacecafe/data"
	"spacecafe/templates"
	"strings"
	"time"
)

type user struct {
	Id int `json:"id"`
	// IdStr    string `json:"-"`
	Name     string `json:"name"`
	Email    string `json:"-"`
	Nimg     string `json:"-"`
	Password []byte `json:"-"`
	// Feeds    []int  `json:"-"`
}

const (
	AmznAccessKeyID     = "AKIAI4WWTVYNOLO54SYQ"
	AmznSecretAccessKey = "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"
)

type notification struct {
	Id        int       `json:"id"`
	UserId    int       `json:"user_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Type      string    `json:"type"`
	TableName string    `json:"table_name"`
	Url       string    `json:"url"`
	Read      bool      `json:"read"`
	Created   time.Time `json:"created"`
}

type unread struct {
	Unread        int            `json:"unread"`
	Notifications []notification `json:"notifications"`
}

func GetNotificationsUnread(w http.ResponseWriter, r *http.Request) {
	userid, ok := auth(w, r)
	if !ok {
		return
	}
	var u unread
	u.Unread = data.CheckNotificationsUnread(w, userid)
	// u.Notifications = []notification{}
	// data.Query(`select * from notifications where user_id = $1 order by created desc limit 5`, userid).Rows(&u.Notifications)
	json(w, &u)
}

func GetNewNotifications(w http.ResponseWriter, r *http.Request) {
	userid, ok := auth(w, r)
	if !ok {
		return
	}
	err := assure(r, "top_id")
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	var notifications []notification
	err = data.Query(`select * from notifications where user_id = $1 and id > $2 order by created desc limit 5`, userid, r.FormValue("top_id")).Rows(&notifications)
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	json(w, notifications)
}

func GetNotifications(w http.ResponseWriter, r *http.Request) {
	userid, ok := auth(w, r)
	if !ok {
		return
	}
	err := assure(r, "last_id")
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	var notifications []notification
	err = data.Query(`select * from notifications where user_id = $1 and id < $2 order by created desc limit 5`, userid, r.FormValue("last_id")).Rows(&notifications)
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	json(w, notifications)
}

func MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	_, ok := auth(w, r)
	if !ok {
		return
	}

	err := assure(r, "notification_id")
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	go data.Query(`update notifications set read = true where id = $1`, r.FormValue("notification_id")).Run()
}

func CompleteProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		templates.Execute(w, "profile_complete", nil)
		return
	}
	userid, ok := auth(w, r)
	if !ok {
		return
	}

	var MaxFileSize int64 = 1024 * 1025 * 50
	if r.ContentLength > MaxFileSize {
		log.Println(r.ContentLength)
		log.Println("Images are very large, please select smaller images.")
		jsonError(w, "Images are very large, please select smaller images.", 417)
		return
	}

	err := r.ParseMultipartForm(MaxFileSize)
	if err != nil {
		log.Println(err)
		jsonError(w, "Sorry, something worng happened, please try again", 417)
		return
	}

	err = assure(r, "anonymous_name:max=25")
	if err != nil {
		log.Println(err)
		jsonError(w, err.Error(), 417)
		return
	}

	images := []string{}

	// parse real identity image from data url
	realDataURL, err := dataurl.Decode(strings.NewReader(r.FormValue("real_img")))
	if err != nil {
		log.Println(err)
		jsonError(w, "Sorry, something worng happened, please try again", 417)
		return
	}

	// save real identity image 190
	err = resizeSavePNG(bytes.NewReader(realDataURL.Data), userid+".png", 190)
	if err != nil {
		log.Println(err)
		jsonError(w, "Sorry, something worng happened, please try again", 417)
		return
	}
	images = append(images, userid+".png")

	// save real identity image 44
	err = resizeSavePNG(bytes.NewReader(realDataURL.Data), userid+"_44.png", 44)
	if err != nil {
		log.Println(err)
		jsonError(w, "Sorry, something worng happened, please try again", 417)
		return
	}
	images = append(images, userid+"_44.png")

	// parse anonymous identity image from data url
	anonymousDataURL, err := dataurl.Decode(strings.NewReader(r.FormValue("anonymous_img")))
	if err != nil {
		log.Println(err)
		jsonError(w, "Sorry, something worng happened, please try again", 417)
		return
	}

	// save anonymous identity image 44 with uuid name
	uid := uuid.NewV4()
	err = resizeSavePNG(bytes.NewReader(anonymousDataURL.Data), uid.String()+".png", 44)
	if err != nil {
		log.Println(err)
		jsonError(w, "Sorry, something worng happened, please try again", 417)
		return
	}
	images = append(images, uid.String()+".png")

	// upload images to amazon s3, them remove them
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(AmznAccessKeyID, AmznSecretAccessKey, ""),
	})
	svc := s3.New(sess)
	bucket := "spacecafe-users"
	for _, img := range images {
		file, err := os.Open("./static/img/" + img)
		if err != nil {
			log.Println(err)
			jsonError(w, "Sorry, something worng happened, please try again", 417)
			return
		}
		_, err = svc.PutObject(&s3.PutObjectInput{
			Body:        file,
			ContentType: aws.String("image/png"),
			Bucket:      &bucket,
			Key:         &img,
		})
		if err != nil {
			log.Println(err)
			jsonError(w, "Sorry, something worng happened, please try again", 417)
			return
		}
		file.Close()
		go os.Remove("./static/img/" + img)
	}

	nname := r.FormValue("anonymous_name")
	nname = html.EscapeString(nname)
	// set anonymous name and anonymous img name
	err = data.Query(`update users set nimg = $1, nname = $2 where id = $3`, uid.String(), nname, userid).Run()
	if err != nil {
		jsonError(w, "Sorry, something worng happened, please try again", 417)
		return
	}
	// log.Println("Successfully uploaded data")
}

func resizeSavePNG(br *bytes.Reader, name string, size int) error {
	img, err := png.Decode(br)
	if err != nil {
		return err
	}
	m := resize.Resize(uint(size), 0, img, resize.Lanczos3)
	out, err := os.Create("./static/img/" + name)
	if err != nil {
		return err
	}
	defer out.Close()
	err = png.Encode(out, m)
	if err != nil {
		return err
	}
	return nil
}

func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		templates.Execute(w, "signup", nil)
		return
	}
	err := assure(r, "name:max=25", "email:max=150", "password:max=150")
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	e := r.FormValue("email")
	if e != "eslam.a.elftoh@gmail.com" && e != "mrwnmonm@gmail.com" {
		jsonError(w, "Fuck you", http.StatusBadRequest)
		return
	}

	// check if account already exists
	exists := false
	err = data.Query(`select exists(select 1 from users where lower(email)=$1) AS "exists"`, r.FormValue("email")).Rows(&exists)
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if exists {
		jsonError(w, "Email already exists", http.StatusNotAcceptable)
		return
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(r.FormValue("password")), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// name := strictSanitizer.Sanitize(r.FormValue("name"))
	name := html.EscapeString(r.FormValue("name"))

	secret := random(15)

	// insert new user
	created := time.Now()
	var id int
	err = data.Query(`INSERT INTO users(
			         id, name, nname, nimg, email, password, secret, created)
			 		VALUES (nextval('users_id_seq'), $1, 'Anonymous', 'anonymous', lower($2), $3, $4, $5) RETURNING id`, name, r.FormValue("email"), hashedPassword, secret, created).Rows(&id)
	if err != nil {
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	idStr := fmt.Sprint(id)
	token, err := data.NewSession(idStr)
	if err != nil {
		// TODO: serious error, should be known as soon as possible
		log.Println(err)
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	go data.InitNotifications(idStr)

	json(w, &struct {
		UserId   string `json:"user_id"`
		UserName string `json:"user_name"`
		Token    string `json:"token"`
	}{Token: token,
		UserId:   idStr,
		UserName: name})
	// TODO: send confirmation email
}

func Login(w http.ResponseWriter, r *http.Request) {
	// if r.Method == "GET" {
	// 	templates.Execute(w, "login", nil)
	// 	return
	// }
	// UploadImg()
	err := assure(r, "email", "password")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user user
	err = data.Query(`select id, name, password from users where lower(email) = $1`, r.FormValue("email")).Rows(&user)
	if err == sql.ErrNoRows {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// CompareHashAndPassword compares a bcrypt hashed password with its possible plaintext equivalent. Returns nil on success, or an error on failure.
	err = bcrypt.CompareHashAndPassword(user.Password, []byte(r.FormValue("password")))
	if err != nil { //login failed
		http.Error(w, "", http.StatusUnauthorized)
		return
	}
	token, err := data.NewSession(fmt.Sprint(user.Id))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		// TODO: serious error, should be known as soon as possible
		log.Println(err)
		return
	}

	json(w, &struct {
		UserId   string `json:"user_id"`
		UserName string `json:"user_name"`
		Token    string `json:"token"`
	}{Token: token,
		UserId:   fmt.Sprint(user.Id),
		UserName: user.Name})

	// c := &http.Cookie{
	// 	Name:  "Token",
	// 	Value: token,
	// }

	// r.AddCookie(c)
	// http.SetCookie(w, c)
	// http.Redirect(w, r, "http://localhost:8000/home", http.StatusMovedPermanently)

	// fmt.Fprintf(w, "You are logged in :) - User%d: %s", user.Id, user.Name)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	go data.EndSession(r)
}
