package actions

import (
	"net/http"
	"strings"

	"todo/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// loadContacts devuelve los contactos ordenados por nombre, sin distinguir mayúsculas,
// con el nombre de su empresa rellenado.
func loadContacts(tx *pop.Connection) (models.Contacts, error) {
	contacts := models.Contacts{}
	if err := tx.Order("lower(name) asc, id asc").All(&contacts); err != nil {
		return nil, err
	}

	companies, err := loadCompanies(tx)
	if err != nil {
		return nil, err
	}
	names := map[uuid.UUID]string{}
	for _, company := range companies {
		names[company.ID] = company.Name
	}
	for i := range contacts {
		if contacts[i].CompanyID.Valid {
			contacts[i].CompanyName = names[contacts[i].CompanyID.UUID]
		}
	}
	return contacts, nil
}

// companyFromForm lee el selector de empresa del formulario. Un valor vacío
// significa «sin empresa»; un id que no existe se descarta como «sin empresa».
func companyFromForm(tx *pop.Connection, raw string) (uuid.NullUUID, error) {
	id, err := uuid.FromString(strings.TrimSpace(raw))
	if err != nil {
		return uuid.NullUUID{}, nil
	}
	exists, err := tx.Where("id = ?", id).Exists(&models.Company{})
	if err != nil || !exists {
		return uuid.NullUUID{}, err
	}
	return uuid.NullUUID{UUID: id, Valid: true}, nil
}

// setContactForm deja en el contexto lo que necesita el formulario de contacto.
func setContactForm(c buffalo.Context, tx *pop.Connection, contact *models.Contact, errors *validate.Errors) error {
	companies, err := loadCompanies(tx)
	if err != nil {
		return err
	}
	c.Set("companies", companies)
	c.Set("contact", contact)
	c.Set("errors", errors)
	return nil
}

// ContactsIndex muestra la lista de contactos y el formulario de alta.
func ContactsIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	contacts, err := loadContacts(tx)
	if err != nil {
		return err
	}

	c.Set("contacts", contacts)
	if err := setContactForm(c, tx, &models.Contact{}, validate.NewErrors()); err != nil {
		return err
	}
	return c.Render(http.StatusOK, r.HTML("contacts/index.plush.html"))
}

// ContactsCreate crea un contacto nuevo a partir del formulario.
func ContactsCreate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	contact := &models.Contact{}
	if err := c.Bind(contact); err != nil {
		return err
	}
	contact.Normalize()
	company, err := companyFromForm(tx, c.Param("company_id"))
	if err != nil {
		return err
	}
	contact.CompanyID = company

	verrs, err := tx.ValidateAndCreate(contact)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		contacts, err := loadContacts(tx)
		if err != nil {
			return err
		}

		c.Set("contacts", contacts)
		if err := setContactForm(c, tx, contact, verrs); err != nil {
			return err
		}
		return c.Render(http.StatusUnprocessableEntity, r.HTML("contacts/index.plush.html"))
	}

	c.Flash().Add("success", "Contacto creado.")
	return c.Redirect(http.StatusSeeOther, "/contactos")
}

// ContactsEdit muestra el formulario de edición de un contacto existente.
func ContactsEdit(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	contact := &models.Contact{}
	if err := tx.Find(contact, c.Param("contact_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := setContactForm(c, tx, contact, validate.NewErrors()); err != nil {
		return err
	}
	return c.Render(http.StatusOK, r.HTML("contacts/edit.plush.html"))
}

// ContactsUpdate actualiza los datos de un contacto.
func ContactsUpdate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	contact := &models.Contact{}
	if err := tx.Find(contact, c.Param("contact_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	input := struct {
		Name      string `form:"name" json:"name"`
		Email     string `form:"email" json:"email"`
		Phone     string `form:"phone" json:"phone"`
		JobTitle  string `form:"job_title" json:"job_title"`
		CompanyID string `form:"company_id" json:"company_id"`
	}{}
	if err := c.Bind(&input); err != nil {
		return err
	}
	contact.Name = input.Name
	contact.Email = input.Email
	contact.Phone = input.Phone
	contact.JobTitle = input.JobTitle
	contact.Normalize()
	company, err := companyFromForm(tx, input.CompanyID)
	if err != nil {
		return err
	}
	contact.CompanyID = company

	verrs, err := tx.ValidateAndUpdate(contact)
	if err != nil {
		return err
	}
	if verrs.HasAny() {
		if err := setContactForm(c, tx, contact, verrs); err != nil {
			return err
		}
		return c.Render(http.StatusUnprocessableEntity, r.HTML("contacts/edit.plush.html"))
	}

	c.Flash().Add("success", "Contacto actualizado.")
	return c.Redirect(http.StatusSeeOther, "/contactos")
}

// ContactsDestroy elimina un contacto.
func ContactsDestroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	contact := &models.Contact{}
	if err := tx.Find(contact, c.Param("contact_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(contact); err != nil {
		return err
	}

	c.Flash().Add("success", "Contacto eliminado.")
	return c.Redirect(http.StatusSeeOther, "/contactos")
}
