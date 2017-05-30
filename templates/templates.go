package templates

import (
	"html/template"
	"net/http"
	"strings"
	"time"
)

var tmpls *template.Template

func add(x, y int) int {
	return x + y
}

func date(date time.Time) string {
	return date.Format("02/Jan/06")
}

func floatToInt(f float64) int {
	return int(f)
}

func noescape(html string) template.HTML {
	return template.HTML(html)
}

func nospaces(text string) string {
	return strings.Replace(text, " ", "", -1)
}

func spaced(text string) string {
	return strings.Replace(text, "_", " ", -1)
}

func Execute(w http.ResponseWriter, name string, data interface{}) {
	tmpls.ExecuteTemplate(w, name+".html", data)
}

func Parse() {
	funcMap := template.FuncMap{"add": add, "floatToInt": floatToInt, "spaced": spaced, "date": date, "noescape": noescape, "nospaces": nospaces}
	tmpls = template.Must(template.New("tmpls").Funcs(funcMap).ParseGlob("./templates/*.html"))
}
