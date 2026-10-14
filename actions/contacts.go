package actions

import (
	"net/http"

	"todo/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
)

// loadContacts devuelve los contactos ordenados por nombre, sin distinguir mayúsculas.
func loadContacts(tx *pop.Connection) (models.Contacts, error) {
	contacts := models.Contacts{}
	if err := tx.Order("lower(name) asc, id asc").All(&contacts); err != nil {
		return nil, err
	}
	return contacts, nil
}

// ContactsIndex muestra la lista de contactos y el formulario de alta.
func ContactsIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	contacts, err := loadContacts(tx)
	if err != nil {
		return err
	}

	c.Set("contacts", contacts)
	c.Set("contact", &models.Contact{})
	c.Set("errors", validate.NewErrors())
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
		c.Set("contact", contact)
		c.Set("errors", verrs)
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

	c.Set("contact", contact)
	c.Set("errors", validate.NewErrors())
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
		Name     string `form:"name" json:"name"`
		Email    string `form:"email" json:"email"`
		Phone    string `form:"phone" json:"phone"`
		JobTitle string `form:"job_title" json:"job_title"`
	}{}
	if err := c.Bind(&input); err != nil {
		return err
	}
	contact.Name = input.Name
	contact.Email = input.Email
	contact.Phone = input.Phone
	contact.JobTitle = input.JobTitle
	contact.Normalize()

	verrs, err := tx.ValidateAndUpdate(contact)
	if err != nil {
		return err
	}
	if verrs.HasAny() {
		c.Set("contact", contact)
		c.Set("errors", verrs)
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
