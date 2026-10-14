package actions

import (
	"strings"

	"todo/public"
	"todo/templates"

	"github.com/gobuffalo/buffalo/render"
)

var r *render.Engine

func init() {
	r = render.New(render.Options{
		// HTML layout to be used for all HTML requests:
		HTMLLayout: "application.plush.html",

		// fs.FS containing templates
		TemplatesFS: templates.FS(),

		// fs.FS containing assets
		AssetsFS: public.FS(),

		// Add template helpers here:
		Helpers: render.Helpers{
			// navSection devuelve la sección del menú a la que pertenece una ruta.
			"navSection": navSection,
			// for non-bootstrap form helpers uncomment the lines
			// below and import "github.com/gobuffalo/helpers/forms"
			// forms.FormKey:     forms.Form,
			// forms.FormForKey:  forms.FormFor,
		},
	})
}

// navSection indica la sección del menú lateral que corresponde a una ruta.
func navSection(path string) string {
	if path == "/contactos" || strings.HasPrefix(path, "/contactos/") {
		return "contacts"
	}
	return "tasks"
}
