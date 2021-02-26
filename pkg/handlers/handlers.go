package handlers

import (
	"net/http"

	"github.com/darinmilner/goserver/pkg/config"
	"github.com/darinmilner/goserver/pkg/models"
	"github.com/darinmilner/goserver/pkg/render"
)

//Repo is the repository used by the handlers
var Repo *Repository

//Repository is the repository type struct
type Repository struct {
	App *config.AppConfig
}

//NewRepo creates a new repository
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

//NewHandlers sets the repository for handlers
func NewHandlers(r *Repository) {
	Repo = r
}

//Home page function
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.RemoteAddr
	m.App.Session.Put(r.Context(), "remoteIP", remoteIP)

	render.RenderTemplate(w, "home.page.html", &models.TemplateData{})
}

//About page function
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	//perform some logic
	stringMap := make(map[string]string)
	stringMap["test"] = "Hello Again"

	remoteIP := m.App.Session.GetString(r.Context(), "remoteIP")

	stringMap["remoteIP"] = remoteIP
	//send the data

	render.RenderTemplate(w, "about.page.html", &models.TemplateData{
		StringMap: stringMap,
	})

}
