package main

import (
	"github.com/baharoam/reservation/pkg/config"
	"github.com/baharoam/reservation/pkg/handlers"
	"net/http"

	//"github.com/bmizerany/pat"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

//mux is http handler
func routes(app *config.AppConfig ) http.Handler{
//	mux := pat.New()
//	mux.Get("/", http.HandlerFunc(handlers.Repo.Home))
//	mux.Get("/about", http.HandlerFunc(handlers.Repo.About))
	mux := chi.NewRouter()
	mux.Use(SessionLoad)
//	mux.Use(WriteToConsole)
	mux.Use(NoSurf)
	mux.Use(middleware.GetHead)
	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	return mux
}