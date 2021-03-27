package render

import (
	"log"
	"net/http"
	"testing"

	"github.com/darinmilner/goserver/internal/models"
)

func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	session.Put(r.Context(), "flash", "123")

	result := AddDefaultData(&td, r)

	if result.Flash != "123" {
		t.Error("Flash value 123 not found in session")
	}
}

func TestRenderTemplate(t *testing.T) {

	pathToTemplates = "./../../templates"
	tc, err := CreateTemplateCache()

	log.Println(pathToTemplates)

	if err != nil {
		t.Error(err)
	}

	app.TemplateCache = tc

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	var ww myWriter

	err = RenderTemplate(&ww, r, "home.page.html", &models.TemplateData{})
	if err != nil {
		t.Error("Error writing template to browser", err)
	}

	err = RenderTemplate(&ww, r, "DoesntExist.page.html", &models.TemplateData{})
	if err == nil {
		t.Error("Rendered nonexistant template to browser")
	}

}

func getSession() (*http.Request, error) {
	r, err := http.NewRequest("GET", "/someurl", nil)
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	ctx, _ = session.Load(ctx, r.Header.Get("X-Session"))
	r = r.WithContext(ctx)

	return r, nil
}

func TestNewTemplates(t *testing.T) {
	NewTemplates(app)
}

func TestCreateTemplateCache(t *testing.T) {
	pathToTemplates = "./../../templates"

	_, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}
}
