package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/manishtomar-cpi/go-server/internal/config"
)

func main() {
	// loads config from YAML
	cfg := config.MustLoad()

	//db setup

	//setup router
	//http.NewServeMux() is like express.Router()
	//HandleFunc("GET /", handler) is like app.get('/', handler)
	router := http.NewServeMux()
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) { // w is response , r is request
		w.Write([]byte("welcome to go server"))
	})
	//setup server -> This is similar to: app.listen(8082, () => console.log('Server started'));
	server := http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}
	fmt.Println("server started")

	err := server.ListenAndServe()

	if err != nil {
		log.Fatal("failed to start server")
	}
}
