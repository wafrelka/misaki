package misaki

import (
	"encoding/json"
	"net/http"
	"path"
	"github.com/markbates/pkger"
)

func from_pkger_file(file_path string) http.Handler {
	h := func(w http.ResponseWriter, req *http.Request) {
		file, err := pkger.Open(path.Join("/pkg/assets", file_path))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer file.Close()
		stat, err := file.Stat()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.ServeContent(w, req, stat.Name(), stat.ModTime(), file)
	}
	return http.HandlerFunc(h)
}

func make_exact(url_path string, h http.Handler) http.Handler {
	g := func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != url_path {
			http.NotFound(w, req)
			return
		}
		h.ServeHTTP(w, req)
	}
	return http.HandlerFunc(g)
}

func match_origin(req *http.Request) bool {
	host := req.Host
	origin := req.Header.Get("Origin")
	http_host := "http://" + host
	https_host := "https://" + host
	return origin == "" || http_host == origin || https_host == origin
}

type CommandPicker func(*http.Request) string
type CommandHandler func(string) (string, int)
type RequestHandler func(http.ResponseWriter, *http.Request)

func synthesize_request_handler(command_handler CommandHandler, command_picker CommandPicker) RequestHandler {

	fn := func(w http.ResponseWriter, req *http.Request) {

		if req.Method != "POST" || req.Body == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !match_origin(req) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		cmd_name := command_picker(req)
		resp, code := command_handler(cmd_name)
		w.WriteHeader(code)
		w.Write([]byte(resp))
	}

	return fn
}

func NewMisakiHandler(command_handler CommandHandler, cmds []Command) http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc(
		"/request/",
		synthesize_request_handler(
			command_handler,
			func(req *http.Request) string {
				return req.URL.Path[len("/request/"):]
			},
		),
	)

	mux.HandleFunc(
		"/request",
		synthesize_request_handler(
			command_handler,
			func(req *http.Request) string {
				return req.PostFormValue("command")
			},
		),
	)

	mux.HandleFunc("/commands", func(w http.ResponseWriter, req *http.Request) {
		resp, _ := json.Marshal(cmds)
		w.Write(resp)
	})

	mux.Handle("/app.css", from_pkger_file("app.css"))
	mux.Handle("/app.js", from_pkger_file("app.js"))
	mux.Handle("/", make_exact("/", from_pkger_file("app.html")))

	return mux
}
