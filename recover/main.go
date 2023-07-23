package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/panic/", panicDemo)
	mux.HandleFunc("/panic-after/", panicAfterDemo)
	mux.HandleFunc("/", hello)
	log.Fatal(http.ListenAndServe(":3000", recoverMiddleware(mux, true)))
}

func recoverMiddleware(next http.Handler, isDev bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				stack := debug.Stack()
				log.Println(string(stack))

				w.WriteHeader(http.StatusInternalServerError)
				if isDev {
					fmt.Fprintf(w, "<h1>panic: %v</h1><pre>%s</pre>", err, string(stack))
					return
				}
				http.Error(w, "Something went wrong", http.StatusInternalServerError)

			}
		}()

		newResponseWriter := &responseWriter{ResponseWriter: w}
		next.ServeHTTP(newResponseWriter, r)
		newResponseWriter.flush()
	}
}

type responseWriter struct {
	http.ResponseWriter
	writes [][]byte
	status int
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.writes = append(rw.writes, b)
	return len(b), nil
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
}

func (rw *responseWriter) flush() error {
	if rw.status != 0 {
		fmt.Println("sfsdfsdfsfsfsfsdfsd")
		fmt.Println(rw.status)
		rw.ResponseWriter.WriteHeader(rw.status)
	}
	for _, b := range rw.writes {
		_, err := rw.ResponseWriter.Write(b)
		if err != nil {
			return err
		}
	}
	return nil
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
