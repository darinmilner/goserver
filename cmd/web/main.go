package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/darinmilner/goserver/internal/config"
	"github.com/darinmilner/goserver/internal/driver"
	"github.com/darinmilner/goserver/internal/handlers"
	"github.com/darinmilner/goserver/internal/helpers"
	"github.com/darinmilner/goserver/internal/models"
	"github.com/darinmilner/goserver/internal/render"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

//main function
func main() {

	db, err := run()

	if err != nil {
		log.Fatal(err)
	}

	defer db.SQL.Close()

	defer close(app.MailChan)

	log.Println("Starting Email listener...")
	//helpers.HashPassword("password123")
	listenForMail()

	//Email from Go Standard Library
	// from := "me@here.com"
	// auth := smtp.PlainAuth("", from, "", "localhost")

	// err = smtp.SendMail("localhost:1025", auth, from, []string{"you@there.com"}, []byte("Hello, Test"))

	if err != nil {
		log.Print("Email Error")
	}
	fmt.Println(fmt.Sprintf("Starting app on port %s", portNumber))

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}
	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {

	//put into the session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})
	//Change to true when in production

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	app.InProduction = false

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour

	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction //True in Production

	app.Session = session

	//connect to db
	log.Println("Connecting to database")
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=bookings user=darinmilner password=")
	if err != nil {
		log.Fatal("Can not connect to DB", err)
		return nil, err
	}

	log.Println("Connected to DB")

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Can not create template cache", err)
		return nil, err
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app, db)

	handlers.NewHandlers(repo)

	render.NewRenderer(&app)

	helpers.NewHelpers(&app)

	return db, nil
}
