package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/baharoam/reservation/internal/config"
	"github.com/baharoam/reservation/internal/driver"
	"github.com/baharoam/reservation/internal/handlers"
	"github.com/baharoam/reservation/internal/helpers"
	"github.com/baharoam/reservation/internal/models"
	"github.com/baharoam/reservation/internal/render"
)

const portNumber = ":8080"
var app config.AppConfig
var session *scs.SessionManager
// main is the main function
func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()
	defer close(app.MailChan)
	listenForMail()

	fmt.Println(fmt.Sprintf("Staring application on port %s", portNumber))

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func run() (*driver.DB, error) {

		// what am I going to put in the session
		gob.Register(models.Reservation{})
		gob.Register(models.User{})
		gob.Register(models.Room{})
		gob.Register(models.Restriction{})

		// read flags

		inProduction := flag.Bool("production", true, "Application is in production")
		useCache := flag.Bool("cache", true, "Use template cache")
		dbHost := flag.String("dbhost", "localhost", "Database host")
		dbName := flag.String("dbname", "reservations", "Database name")
		dbUser := flag.String("dbuser", "postgres", "Database user")
		dbPass := flag.String("dbPass", "", "Database password")
		dbPort := flag.String("dbPort", "5432", "Database port")
		dbSSL := flag.String("dbssl", "disable", "Database ssl settings (disable, prefer, require)")

		flag.Parse()

		if *dbName == "" || *dbUser =="" {
			fmt.Println("Missing required flags")
			os.Exit(1)
		}

		mailChan := make(chan models.MailData)
		app.MailChan = mailChan

		// change this to true when in production
		app.InProduction = *inProduction
		app.UseCache = *useCache

	
		// set up the session
		session = scs.New()
		session.Lifetime = 24 * time.Hour
		session.Cookie.Persist = true
		session.Cookie.SameSite = http.SameSiteLaxMode
		session.Cookie.Secure = app.InProduction
	
		app.Session = session
	// connect to database ("host=localhost port=5432 dbname=reservations user=postgres password=")
	log.Println("Connecting to database...")
	connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", *dbHost, *dbPort, *dbName, *dbUser, *&dbPass, *dbSSL)
	db, err := driver.ConnectSQL(connectionString)
	if err != nil {
		log.Fatal("Cannot connect to database! Dying...")
	}

	log.Println("Connected to database!")

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
		return nil, err
	}

	app.TemplateCache = tc

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)



	return db, nil

}


