package actions

import (
	"net/http"

	"todo/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
)

// loadCompanies devuelve las empresas ordenadas por nombre, sin distinguir mayúsculas.
func loadCompanies(tx *pop.Connection) (models.Companies, error) {
	companies := models.Companies{}
	if err := tx.Order("lower(name) asc, id asc").All(&companies); err != nil {
		return nil, err
	}
	return companies, nil
}

// CompaniesIndex muestra la lista de empresas y el formulario de alta.
func CompaniesIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	companies, err := loadCompanies(tx)
	if err != nil {
		return err
	}

	c.Set("companies", companies)
	c.Set("company", &models.Company{})
	c.Set("errors", validate.NewErrors())
	return c.Render(http.StatusOK, r.HTML("companies/index.plush.html"))
}

// CompaniesCreate crea una empresa nueva a partir del formulario.
func CompaniesCreate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	company := &models.Company{}
	if err := c.Bind(company); err != nil {
		return err
	}
	company.Normalize()

	verrs, err := tx.ValidateAndCreate(company)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		companies, err := loadCompanies(tx)
		if err != nil {
			return err
		}

		c.Set("companies", companies)
		c.Set("company", company)
		c.Set("errors", verrs)
		return c.Render(http.StatusUnprocessableEntity, r.HTML("companies/index.plush.html"))
	}

	c.Flash().Add("success", "Empresa creada.")
	return c.Redirect(http.StatusSeeOther, "/empresas")
}

// CompaniesShow muestra la ficha de una empresa con sus contactos.
func CompaniesShow(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	company := &models.Company{}
	if err := tx.Find(company, c.Param("company_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	contacts := models.Contacts{}
	if err := tx.Where("company_id = ?", company.ID).Order("lower(name) asc, id asc").All(&contacts); err != nil {
		return err
	}

	c.Set("company", company)
	c.Set("contacts", contacts)
	return c.Render(http.StatusOK, r.HTML("companies/show.plush.html"))
}

// CompaniesEdit muestra el formulario de edición de una empresa existente.
func CompaniesEdit(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	company := &models.Company{}
	if err := tx.Find(company, c.Param("company_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	c.Set("company", company)
	c.Set("errors", validate.NewErrors())
	return c.Render(http.StatusOK, r.HTML("companies/edit.plush.html"))
}

// CompaniesUpdate actualiza los datos de una empresa.
func CompaniesUpdate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	company := &models.Company{}
	if err := tx.Find(company, c.Param("company_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	input := struct {
		Name    string `form:"name" json:"name"`
		Website string `form:"website" json:"website"`
		Sector  string `form:"sector" json:"sector"`
	}{}
	if err := c.Bind(&input); err != nil {
		return err
	}
	company.Name = input.Name
	company.Website = input.Website
	company.Sector = input.Sector
	company.Normalize()

	verrs, err := tx.ValidateAndUpdate(company)
	if err != nil {
		return err
	}
	if verrs.HasAny() {
		c.Set("company", company)
		c.Set("errors", verrs)
		return c.Render(http.StatusUnprocessableEntity, r.HTML("companies/edit.plush.html"))
	}

	c.Flash().Add("success", "Empresa actualizada.")
	return c.Redirect(http.StatusSeeOther, "/empresas")
}

// CompaniesDestroy elimina una empresa; sus contactos se quedan sin empresa.
func CompaniesDestroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	company := &models.Company{}
	if err := tx.Find(company, c.Param("company_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(company); err != nil {
		return err
	}

	c.Flash().Add("success", "Empresa eliminada.")
	return c.Redirect(http.StatusSeeOther, "/empresas")
}
