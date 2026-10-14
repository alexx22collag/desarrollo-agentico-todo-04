package actions

import (
	"net/url"
	"strings"

	"todo/models"

	"github.com/gofrs/uuid"
)

const acmeID = "11111111-1111-4111-8111-111111111111"

func (as *ActionSuite) Test_CompaniesIndex_EmptyState() {
	res := as.HTML("/empresas").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "No hay empresas todavía")
}

func (as *ActionSuite) Test_CompaniesIndex_ListsSortedIgnoringCase() {
	as.LoadFixture("empresas con contactos")

	body := as.HTML("/empresas").Get().Body.String()
	acme := strings.Index(body, "Acme S.L.")
	globex := strings.Index(body, "globex")
	as.True(acme >= 0 && acme < globex, "orden esperado: Acme, globex")
}

func (as *ActionSuite) Test_Layout_HighlightsCompaniesSection() {
	body := as.HTML("/empresas").Get().Body.String()
	as.Contains(body, `href="/empresas" class="nav-link" aria-current="true"`)
	as.NotContains(body, `href="/contactos" class="nav-link" aria-current="true"`)
	as.NotContains(body, `href="/" class="nav-link" aria-current="true"`)
}

func (as *ActionSuite) Test_CompaniesCreate_Valid() {
	res := as.HTML("/empresas").Post(&models.Company{Name: "  Initech  ", Website: "https://initech.example.com", Sector: "Software"})
	as.Equal(303, res.Code)
	as.Equal("/empresas", res.Header().Get("Location"))

	company := &models.Company{}
	as.NoError(as.DB.First(company))
	as.Equal("Initech", company.Name)

	as.Contains(as.HTML("/empresas").Get().Body.String(), "Empresa creada.")
}

func (as *ActionSuite) Test_CompaniesCreate_EmptyName_Invalid() {
	res := as.HTML("/empresas").Post(&models.Company{Name: " ", Sector: "Software"})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El nombre es obligatorio.")
	as.Contains(res.Body.String(), "Software")
}

func (as *ActionSuite) Test_CompaniesCreate_DuplicateName_Invalid() {
	as.LoadFixture("empresas con contactos")

	res := as.HTML("/empresas").Post(&models.Company{Name: "ACME s.l."})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "Ya existe una empresa con ese nombre.")

	count, err := as.DB.Count(&models.Company{})
	as.NoError(err)
	as.Equal(2, count)
}

func (as *ActionSuite) Test_CompaniesShow_ListsItsContacts() {
	as.LoadFixture("empresas con contactos")

	res := as.HTML("/empresas/%s", acmeID).Get()
	as.Equal(200, res.Code)
	body := res.Body.String()
	as.Contains(body, "Acme S.L.")
	as.Contains(body, "Ana Acme")
	as.Contains(body, "Beto Acme")
	as.NotContains(body, "Cris Libre")
}

func (as *ActionSuite) Test_CompaniesShow_WithoutContacts() {
	as.LoadFixture("empresas con contactos")

	res := as.HTML("/empresas/%s", "22222222-2222-4222-8222-222222222222").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "Esta empresa no tiene contactos todavía.")
}

func (as *ActionSuite) Test_CompaniesShow_NotFound() {
	as.Equal(404, as.HTML("/empresas/%s", uuid.Must(uuid.NewV4())).Get().Code)
}

func (as *ActionSuite) Test_CompaniesUpdate_Valid() {
	as.LoadFixture("empresas con contactos")

	res := as.HTML("/empresas/%s", acmeID).Put(url.Values{"name": {"Acme Corp"}, "website": {""}, "sector": {"Química"}})
	as.Equal(303, res.Code)

	company := &models.Company{}
	as.NoError(as.DB.Find(company, acmeID))
	as.Equal("Acme Corp", company.Name)
	as.Equal("Química", company.Sector)
}

func (as *ActionSuite) Test_CompaniesUpdate_KeepingOwnNameIsValid() {
	as.LoadFixture("empresas con contactos")

	res := as.HTML("/empresas/%s", acmeID).Put(url.Values{"name": {"acme s.l."}, "sector": {"Otro"}})
	as.Equal(303, res.Code)
}

func (as *ActionSuite) Test_CompaniesUpdate_DuplicateName_Invalid() {
	as.LoadFixture("empresas con contactos")

	res := as.HTML("/empresas/%s", acmeID).Put(url.Values{"name": {"Globex"}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "Ya existe una empresa con ese nombre.")
}

func (as *ActionSuite) Test_CompaniesEdit_NotFound() {
	as.Equal(404, as.HTML("/empresas/%s/edit", uuid.Must(uuid.NewV4())).Get().Code)
}

func (as *ActionSuite) Test_CompaniesDestroy_LeavesContactsWithoutCompany() {
	as.LoadFixture("empresas con contactos")

	res := as.HTML("/empresas/%s", acmeID).Delete()
	as.Equal(303, res.Code)
	as.Equal("/empresas", res.Header().Get("Location"))

	companies, err := as.DB.Count(&models.Company{})
	as.NoError(err)
	as.Equal(1, companies)

	contacts, err := as.DB.Count(&models.Contact{})
	as.NoError(err)
	as.Equal(3, contacts)

	contact := &models.Contact{}
	as.NoError(as.DB.Where("name = ?", "Ana Acme").First(contact))
	as.False(contact.CompanyID.Valid)
}

func (as *ActionSuite) Test_CompaniesDestroy_NotFound() {
	as.Equal(404, as.HTML("/empresas/%s", uuid.Must(uuid.NewV4())).Delete().Code)
}

func (as *ActionSuite) Test_Contacts_ShowTheirCompanyInTheList() {
	as.LoadFixture("empresas con contactos")

	body := as.HTML("/contactos").Get().Body.String()
	as.Contains(body, "Acme S.L.")
	as.Contains(body, `<option value="">Sin empresa</option>`)
}

func (as *ActionSuite) Test_ContactsCreate_WithCompany() {
	as.LoadFixture("empresas con contactos")

	res := as.HTML("/contactos").Post(url.Values{"name": {"Dani Nuevo"}, "company_id": {acmeID}})
	as.Equal(303, res.Code)

	contact := &models.Contact{}
	as.NoError(as.DB.Where("name = ?", "Dani Nuevo").First(contact))
	as.True(contact.CompanyID.Valid)
	as.Equal(acmeID, contact.CompanyID.UUID.String())
}

func (as *ActionSuite) Test_ContactsCreate_UnknownCompanyBecomesNoCompany() {
	as.LoadFixture("empresas con contactos")

	res := as.HTML("/contactos").Post(url.Values{"name": {"Eva"}, "company_id": {uuid.Must(uuid.NewV4()).String()}})
	as.Equal(303, res.Code)

	contact := &models.Contact{}
	as.NoError(as.DB.Where("name = ?", "Eva").First(contact))
	as.False(contact.CompanyID.Valid)
}

func (as *ActionSuite) Test_ContactsUpdate_ChangeAndRemoveCompany() {
	as.LoadFixture("empresas con contactos")
	contact := &models.Contact{}
	as.NoError(as.DB.Where("name = ?", "Cris Libre").First(contact))

	res := as.HTML("/contactos/%s", contact.ID).Put(url.Values{"name": {"Cris Libre"}, "company_id": {"22222222-2222-4222-8222-222222222222"}})
	as.Equal(303, res.Code)
	as.NoError(as.DB.Reload(contact))
	as.True(contact.CompanyID.Valid)

	res = as.HTML("/contactos/%s", contact.ID).Put(url.Values{"name": {"Cris Libre"}, "company_id": {""}})
	as.Equal(303, res.Code)
	as.NoError(as.DB.Reload(contact))
	as.False(contact.CompanyID.Valid)
}

func (as *ActionSuite) Test_ContactsEdit_PreselectsCompany() {
	as.LoadFixture("empresas con contactos")
	contact := &models.Contact{}
	as.NoError(as.DB.Where("name = ?", "Ana Acme").First(contact))

	body := as.HTML("/contactos/%s/edit", contact.ID).Get().Body.String()
	as.Contains(body, `<option value="`+acmeID+`" selected>Acme S.L.</option>`)
}
