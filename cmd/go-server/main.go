package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/manishtomar-cpi/go-server/internal/config"
	student "github.com/manishtomar-cpi/go-server/internal/http/handllers/students"
	"github.com/manishtomar-cpi/go-server/internal/storage/sqlite"
)

func main() {
	// loads config from YAML
	cfg := config.MustLoad()

	//db setup
	storage, dbErr := sqlite.New(cfg)

	if dbErr != nil {
		log.Fatal(dbErr)
	}

	slog.Info("storage init", slog.String("env", cfg.Env))
	//setup router
	//http.NewServeMux() is like express.Router()
	//HandleFunc("GET /", handler) is like app.get('/', handler)
	router := http.NewServeMux()
	router.HandleFunc("POST /api/students", student.New(storage))
	router.HandleFunc("GET /api/ready", student.Ready())
	//setup server -> This is similar to: app.listen(8082, () => console.log('Server started'));
	server := http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}
	fmt.Println("server started")

	//shut down server gracefully -> mean if server shut down in production so the ongoing requests will not intruppted first those requests will complete then the server will shut down
	done := make(chan os.Signal, 1)                                    //make buffered channel that will listen all interptions and send the response to done
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM) // means if something these happen notify to done chan

	go func() { // so over server is running in seprate go routine
		err := server.ListenAndServe()

		if err != nil {
			log.Fatal("failed to start server")
		}
	}()
	<-done // we will block here untill we dont get any intruptions ->  signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	slog.Info("shutting down the server...")

	//Try to gracefully shut down the server, but if it takes longer than 5 seconds, force quit.
	ctx, cancle := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle()
	err := server.Shutdown(ctx) // shutdown the server graceffully but somethime its take time somethime it may hang here so that we used the timer if server not shutdown in this time report us
	if err != nil {
		slog.Error("failed to shut down server", slog.String("error:", err.Error()))
	}
	slog.Info("Server shutdoen successfully")
}
