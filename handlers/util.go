package handlers

import (
	"crypto/rand"
	// "crypto/sha256"
	// "encoding/base64"
	ejson "encoding/json"
	"errors"
	"github.com/microcosm-cc/bluemonday"
	// "fmt"
	"log"
	"spacecafe/data"
	"spacecafe/templates"
	// "time"
	// "github.com/gorilla/securecookie"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const (
	UnauthorizedRedirectURL = "http://localhost:8000/login"
)

var (
	strictSanitizer = bluemonday.StrictPolicy()
)

func spaced(text string) string {
	return strings.Replace(text, "_", " ", -1)
}

// func Assure(r *http.Request, params ...string) error {
// 	return assure(r, params[0])
// }

// assure presence of params and checks other validations
// numeric: should be numerical value
// max=x: can't be more than x
func assure(r *http.Request, params ...string) error {
	var err error
	err = r.ParseForm()
	if err != nil {
		return err
	}

	for _, param := range params {
		p := strings.Split(param, ":")

		if len(p) < 2 { // param has no validations, just assure existance
			if r.FormValue(p[0]) == "" {
				return errors.New(p[0] + " can't be empty")
			}
			continue
		}

		if r.FormValue(p[0]) == "" && !strings.Contains(p[1], "empty") {
			return errors.New(p[0] + " can't be empty")
		}

		s := strings.Split(p[1], ",")

		for _, c := range s {
			if strings.Contains(c, "=") {
				vs := strings.Split(c, "=")
				switch vs[0] {
				case "max":
					max, _ := strconv.Atoi(vs[1])
					prm := strings.TrimSpace(r.FormValue(p[0]))
					if len(prm) > max {
						return errors.New(p[0] + " can't be more than " + vs[1] + " characters")
					}
				case "min":
					min, _ := strconv.Atoi(vs[1])
					prm := strings.TrimSpace(r.FormValue(p[0]))
					if len(prm) < min {
						return errors.New(p[0] + " can't be less than " + vs[1] + " characters")
					}
				}
			} else {
				switch c {
				case "numeric":
					_, err := strconv.Atoi(r.FormValue(p[0]))
					if err != nil {
						return errors.New(p[0] + " should be an integer string")
					}
				}
			}
		}

	}
	return nil
}

// func auth(r *http.Request) (*user, bool) {
// 	var user user
// 	err := db.Query(`select * from users where email = $1`, "mrwnmonm@gmail.com").Rows(&user)
// 	if err != nil {
// 		return nil, false
// 	}
// 	user.IdStr = strconv.Itoa(user.Id)
// 	return &user, true
// }

func auth(w http.ResponseWriter, r *http.Request) (string, bool) {
	id, err := data.GetSession(r)
	if err != nil {
		templates.Execute(w, "welcome", nil)
		return "", false
	}
	return id, true
}

func softAuth(w http.ResponseWriter, r *http.Request) (string, bool) {
	id, err := data.GetSession(r)
	if err != nil {
		log.Println(err)
		return "", false
	}
	return id, true
}

// Authenticate reads session id from "X-TOKEN" header, and fetch it from redis
// if not found it returns 401
// TODO: server can't expire certain session
// it only can expire all sessions per user
// func auth(r *http.Request) (int, bool) {
// 	secret := r.Header.Get("Secret")
// 	if secret == "" {
// 		return 0, false
// 	}
// 	// remove the leading and the trailing double parentheses from the header value
// 	secret = strings.TrimPrefix(secret, "\"")
// 	secret = strings.TrimSuffix(secret, "\"")

// 	// the secret should be on the form userId:unixtime:code
// 	s := strings.Split(secret, ":")

// 	if len(s) != 3 {
// 		return 0, false
// 	}

// 	id, err := strconv.Atoi(s[0])
// 	if err != nil {
// 		return 0, false
// 	}

// 	random, err := redis.GetUserSecret(id)
// 	if err != nil {
// 		return 0, false
// 	}

// 	if s[2] != util.SHA(s[0]+":"+s[1]+":"+random) {
// 		return 0, false
// 	}

// 	return id, true
// }

func random(strenght int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, strenght)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

// func SHA(x string) string {
// 	h := sha256.New()
// 	h.Write([]byte(x))
// 	return base64.URLEncoding.EncodeToString(h.Sum(nil))
// }

// Using gorilla secure cookie for authentication
// more info http://www.gorillatoolkit.org/pkg/securecookie
// It is recommended to use a key with 32 or 64 bytes.
// var hashKey = []byte{0xcc, 0xc, 0x95, 0xca, 0x4a, 0xa0, 0x4f, 0x96, 0x77, 0x9, 0xab, 0xee, 0xf, 0xbe, 0xd0, 0x42, 0xf9, 0xf4, 0xb8, 0xa9, 0x4a, 0xfe, 0x9f, 0x6b, 0xcb, 0xa, 0x58, 0x1c, 0xf1, 0xbb, 0x31, 0x4c, 0x5c, 0x47, 0x8, 0x0, 0xbd, 0xd0, 0x35, 0x85, 0xae, 0xba, 0x9e, 0xe0, 0x2a, 0x81, 0x77, 0x96, 0x2e, 0x95, 0xc3, 0x2a, 0xe7, 0x6a, 0x39, 0x48, 0x68, 0xd0, 0xf0, 0x5c, 0x88, 0xc3, 0x8f, 0xb9}

//The blockKey is optional, used to encrypt the cookie value -- set it to nil to not use encryption.
//If set, the length must correspond to the block size of the encryption algorithm.
//For AES, used by default, valid lengths are 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.
// var blockKey = []byte{0xd0, 0xe, 0x4, 0x7c, 0x6e, 0x0, 0xd3, 0x82, 0x63, 0xfe, 0x3c, 0xcb, 0x22, 0xf9, 0xd9, 0xa6, 0x85, 0x54, 0xa, 0x37, 0x38, 0x1c, 0x2d, 0xef, 0xc, 0x39, 0xa1, 0xcd, 0xdc, 0xc9, 0x23, 0x68}
// var SecureCookie = securecookie.New(hashKey, blockKey)

// func main() {
// 	e, _ := SecureCookie.Encode("xtoken", "password")
// 	fmt.Printf("%#v", e)
// }

// var s string

// var cookieMap map[string]string

// func Encode() (string, error) {
// 	random := randStr(7)
// 	hash := map[string]string{
// 		"rand": random,
// 		"ip": "asdfasd",
// 		"name": "",
// 		"id":   "222",
// 	}
// 	encoded, err := SecureCookie.Encode("secret", hash)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// save random to redis
// 	return encoded, nil
// }

// func Decode(encoded string) (*map[string]string, error) {
// 	hash := make(map[string]string)
// 	err := SecureCookie.Decode("secret", encoded, &hash)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// get random from redis and check it
// 	return &hash, nil
// }

func json(w http.ResponseWriter, data interface{}, statuses ...int) {
	jsonObj, err := ejson.Marshal(data)
	// TODO: see if you need to use htmlescape here or not
	// json.HTMLEscape(dst, src)
	if err != nil {
		// log.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte{})
		return
	}

	status := 200
	if len(statuses) > 0 {
		status = statuses[0]
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// log.Println(jsonObj)
	w.Write(jsonObj)
	return
}

func toJson(data interface{}) (string, error) {
	jsonObj, err := ejson.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonObj), nil
}

type responseErr struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	json(w, &responseErr{status, msg}, status)
}

var (
	urlExp *regexp.Regexp
)

func init() {
	urlExp = regexp.MustCompile(`((https?|ftp)://|www\.)[^\s/$.?#].[^\s]*`)
}

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

func linkfy(body string) string {
	// TODO: see what else has to do to the body string
	// body = template.HTMLEscapeString(body)

	// hyperlink web links
	// body = urlExp.ReplaceAllString(body, `<a href="$0" target="_blank" rel="nofollow">$0</a>`)

	body = urlExp.ReplaceAllStringFunc(body, replaceurls)

	body = strings.Replace(body, "\n", "", -1)

	// reduce multilines to just one, and make right paragraphs and line breaks
	// body = "<p>" + body + "</p>"
	// x := strings.Split(body, "\n")
	// last := 0 // 0 means (<p> or <br>)   1 means </p>
	// for i := range x {
	// 	x[i] = strings.Trim(x[i], " ")
	// 	if i == 0 {
	// 		continue
	// 	}
	// 	if x[i] == "" {
	// 		last = 1
	// 	} else {
	// 		switch last {
	// 		case 0:
	// 			x[i] = "<br>" + x[i]
	// 			last = 0
	// 		case 1:
	// 			x[i] = "</p><p>" + x[i]
	// 			last = 0
	// 		}
	// 	}
	// }

	// final := strings.Join(x, "")

	// if final == "<p></p>" {
	// 	final = ""
	// }

	return body
}
