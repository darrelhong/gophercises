package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/quick"
	"github.com/alecthomas/chroma/styles"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/", debugHandler)
	mux.HandleFunc("/panic/", panicDemo)
	mux.HandleFunc("/panic-after/", panicAfterDemo)
	mux.HandleFunc("/", hello)
	log.Fatal(http.ListenAndServe(":3000", devMw(mux)))
}

func devMw(app http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
				stack := debug.Stack()
				log.Println(string(stack))
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "<h1>panic: %v</h1><pre>%s</pre>", err, makeLinks(string(stack)))
			}
		}()
		app.ServeHTTP(w, r)
	}
}

func debugHandler(w http.ResponseWriter, r *http.Request) {
	filePath := r.FormValue("path")
	line := r.FormValue("line")
	lineNum, err := strconv.Atoi(line)
	if err != nil {
		lineNum = -1
	}
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	b := bytes.NewBuffer(nil)
	_, err = io.Copy(b, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var highlighLines [][2]int
	if lineNum > 0 {
		highlighLines = append(highlighLines, [2]int{lineNum, lineNum})
	}
	formatter := html.New(html.WithLineNumbers(true), html.HighlightLines(highlighLines))
	lexers := lexers.Get("go")
	iterator, err := lexers.Tokenise(nil, b.String())
	styles := styles.Get("nord")
	w.Header().Set("Content-Type", "text/html")
	formatter.Format(w, styles, iterator)

	quick.Highlight(w, b.String(), "go", "html", "nord")
}

func panicDemo(w http.ResponseWriter, r *http.Request) {
	funcThatPanics()
}

func panicAfterDemo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Hello!</h1>")
	funcThatPanics()
}

func funcThatPanics() {
	panic("Oh no!")
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>Hello!</h1>")
}

func makeLinks(stack string) string {
	lines := strings.Split(stack, "\n")
	for idx, line := range lines {
		if len(line) != 0 && line[0] != '\t' {
			continue
		}
		lineArr := strings.Split(line, ":")
		file := strings.TrimPrefix(lineArr[0], "\t")
		if len(file) == 0 {
			continue
		}
		lineNum := strings.Split(lineArr[1], " ")[0]
		query := url.Values{}
		query.Set("path", file)
		query.Set("line", lineNum)
		lines[idx] = "\t<a href=\"/debug/?" + query.Encode() + "\">" + file + ":" + lineNum + "</a>" + line[len(file)+2+len(lineNum):]
	}
	return strings.Join(lines, "\n")
}
