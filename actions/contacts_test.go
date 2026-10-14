package actions

import (
	"net/url"
	"strings"

	"todo/models"

	"github.com/gofrs/uuid"
)

func (as *ActionSuite) Test_ContactsIndex_EmptyState() {
	res := as.HTML("/contactos").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "No hay contactos todavía")
}

func (as *ActionSuite) Test_ContactsIndex_ListsContactsSortedByNameIgnoringCase() {
	as.LoadFixture("tres contactos")

	res := as.HTML("/contactos").Get()
	as.Equal(200, res.Code)

	body := res.Body.String()
	as.Contains(body, "ana@example.com")
	as.Contains(body, "Directora comercial")

	ana := strings.Index(body, "ana Torres")
	bruno := strings.Index(body, "Bruno Díaz")
	carla := strings.Index(body, "Carla Ruiz")
	as.True(ana >= 0 && ana < bruno && bruno < carla, "orden esperado: ana, Bruno, Carla")
}

func (as *ActionSuite) Test_Layout_HighlightsActiveSection() {
	contacts := as.HTML("/contactos").Get().Body.String()
	as.Contains(contacts, `href="/contactos" class="nav-link" aria-current="true"`)
	as.NotContains(contacts, `href="/" class="nav-link" aria-current="true"`)

	tasks := as.HTML("/").Get().Body.String()
	as.Contains(tasks, `href="/" class="nav-link" aria-current="true"`)
	as.NotContains(tasks, `href="/contactos" class="nav-link" aria-current="true"`)
}

func (as *ActionSuite) Test_ContactsCreate_Valid() {
	res := as.HTML("/contactos").Post(&models.Contact{
		Name:     "  Nuevo Contacto  ",
		Email:    " nuevo@example.com ",
		Phone:    "600 123 456",
		JobTitle: "Gerente",
	})
	as.Equal(303, res.Code)
	as.Equal("/contactos", res.Header().Get("Location"))

	contact := &models.Contact{}
	as.NoError(as.DB.First(contact))
	as.Equal("Nuevo Contacto", contact.Name)
	as.Equal("nuevo@example.com", contact.Email)
	as.Equal("600 123 456", contact.Phone)
	as.Equal("Gerente", contact.JobTitle)

	list := as.HTML("/contactos").Get()
	as.Contains(list.Body.String(), "Contacto creado.")
	as.Contains(list.Body.String(), "Nuevo Contacto")
}

func (as *ActionSuite) Test_ContactsCreate_EmptyName_Invalid() {
	res := as.HTML("/contactos").Post(&models.Contact{Name: "   ", Email: "x@example.com"})
	as.Equal(422, res.Code)
	body := res.Body.String()
	as.Contains(body, "El nombre es obligatorio.")
	as.Contains(body, "x@example.com") // conserva lo escrito

	count, err := as.DB.Count(&models.Contact{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_ContactsCreate_InvalidEmail() {
	res := as.HTML("/contactos").Post(&models.Contact{Name: "Ana", Email: "no-es-un-email"})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El email no tiene un formato válido.")

	count, err := as.DB.Count(&models.Contact{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_ContactsEdit_ShowsForm() {
	as.LoadFixture("tres contactos")
	contact := &models.Contact{}
	as.NoError(as.DB.Where("name = ?", "Carla Ruiz").First(contact))

	res := as.HTML("/contactos/%s/edit", contact.ID).Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "Carla Ruiz")
	as.Contains(res.Body.String(), "carla@example.com")
}

func (as *ActionSuite) Test_ContactsEdit_NotFound() {
	res := as.HTML("/contactos/%s/edit", uuid.Must(uuid.NewV4())).Get()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_ContactsUpdate_Valid() {
	as.LoadFixture("tres contactos")
	contact := &models.Contact{}
	as.NoError(as.DB.Where("name = ?", "Bruno Díaz").First(contact))

	res := as.HTML("/contactos/%s", contact.ID).Put(url.Values{
		"name":      {" Bruno D. "},
		"email":     {"bruno@example.com"},
		"phone":     {"611 222 333"},
		"job_title": {"Logística"},
	})
	as.Equal(303, res.Code)
	as.Equal("/contactos", res.Header().Get("Location"))

	as.NoError(as.DB.Reload(contact))
	as.Equal("Bruno D.", contact.Name)
	as.Equal("bruno@example.com", contact.Email)
	as.Equal("611 222 333", contact.Phone)
	as.Equal("Logística", contact.JobTitle)
}

func (as *ActionSuite) Test_ContactsUpdate_Invalid_KeepsStoredData() {
	as.LoadFixture("tres contactos")
	contact := &models.Contact{}
	as.NoError(as.DB.Where("name = ?", "Carla Ruiz").First(contact))

	res := as.HTML("/contactos/%s", contact.ID).Put(url.Values{"name": {""}, "email": {"carla@example.com"}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El nombre es obligatorio.")

	as.NoError(as.DB.Reload(contact))
	as.Equal("Carla Ruiz", contact.Name)
}

func (as *ActionSuite) Test_ContactsUpdate_NotFound() {
	res := as.HTML("/contactos/%s", uuid.Must(uuid.NewV4())).Put(url.Values{"name": {"X"}})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_ContactsDestroy() {
	as.LoadFixture("tres contactos")
	contact := &models.Contact{}
	as.NoError(as.DB.Where("name = ?", "Bruno Díaz").First(contact))

	res := as.HTML("/contactos/%s", contact.ID).Delete()
	as.Equal(303, res.Code)
	as.Equal("/contactos", res.Header().Get("Location"))

	count, err := as.DB.Count(&models.Contact{})
	as.NoError(err)
	as.Equal(2, count)

	list := as.HTML("/contactos").Get()
	as.Contains(list.Body.String(), "Contacto eliminado.")
}

func (as *ActionSuite) Test_ContactsDestroy_NotFound() {
	res := as.HTML("/contactos/%s", uuid.Must(uuid.NewV4())).Delete()
	as.Equal(404, res.Code)
}
