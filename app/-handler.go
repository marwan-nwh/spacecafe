package app

import (
	"log"
	"net/http"
	// "os"
	// "spacecafe/pkg/sentry"
	// "text/tabwriter"
)

type handler struct {
	Action string
	Desc   string
	Method string
	URI    string
	Auth   bool
	Params []string
	Before http.Handler
	After  http.Handler
	Do     func(w http.ResponseWriter, r *http.Request)
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// check request method
	if h.Method != "" && h.Method != r.Method {
		http.NotFound(w, r)
		return
	}

	if h.Auth {
		_, ok := auth(r)
		if !ok {
			http.Error(w, "", http.StatusUnauthorized)
			return
		}
	}

	if len(h.Params) > 0 {
		err := assure(r, h.Params)
		if err != nil {
			log.Println(err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}
	}

	h.Do(w, r)
}

func RegisterHandlers(mux *http.ServeMux) {
	// file, err := os.Create("app/-API")
	// if err != nil {
	// 	panic(err)
	// }
	// const padding = 3
	// w := tabwriter.NewWriter(file, 0, 0, padding, ' ', tabwriter.TabIndent)
	// method := "ANY"
	// auth := "Auth"
	// fmt.Fprintln(w, "Comments")
	for _, h := range Comments {
		// if h.Method != "" {
		// 	method = h.Method
		// }
		// if !h.Auth {
		// 	auth = ""
		// }
		// fmt.Fprintln(w, method+"\t"+h.URI+"\t"+auth+"\t//"+h.Action+"\t")

		// h.Do = raven.RecoveryHandler(h.Do)

		// TODO: make all handlers use the new handler to recover from thier panics
		mux.Handle(h.URI, h)
	}
	// // GenerateResource("tests")
	// // w.Flush()
	// file.Close()
}

// func GenerateResource(name string) {
// 	file, err := os.Create("app/" + name + ".go")
// 	if err != nil {
// 		panic(err)
// 	}
// 	file.WriteString(fmt.Sprintf(`package app

// import (
// 	"net/http"
// )

// var %s = []handler{
// 	handler{
// 		Action: "Create",
// 		Method: "POST",
// 		URI:    "/%s",
// 		// Auth:   <true, false>,
// 		// Params: []string{...},
// 		Do: func(w http.ResponseWriter, r *http.Request) {

// 		}},
// 	handler{
// 		Action: "Show",
// 		Method: "GET",
// 		URI:    "/%s/:id",
// 		// Auth:   <true, false>,
// 		// Params: []string{...},
// 		Do: func(w http.ResponseWriter, r *http.Request) {

// 		}},
// 	handler{
// 		Action: "Update",
// 		Method: "PUT",
// 		URI:    "/%s/:id",
// 		// Auth:   <true, false>,
// 		// Params: []string{...},
// 		Do: func(w http.ResponseWriter, r *http.Request) {

// 		}},
// 	handler{
// 		Action: "Delete",
// 		Method: "DELETE",
// 		URI:    "/%s/:id",
// 		// Auth:   <true, false>,
// 		// Params: []string{...},
// 		Do: func(w http.ResponseWriter, r *http.Request) {

// 		}}}
// 		`, strings.Title(name), name, name, name, name))
// 	file.Close()
// }
