package models

import (
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// Longitudes máximas de los campos de un contacto.
const (
	ContactNameMaxLength     = 120
	ContactEmailMaxLength    = 200
	ContactPhoneMaxLength    = 40
	ContactJobTitleMaxLength = 120
)

// Contact representa una persona de contacto del CRM.
type Contact struct {
	ID       uuid.UUID `json:"id" db:"id" form:"-"`
	Name     string    `json:"name" db:"name" form:"name"`
	Email    string    `json:"email" db:"email" form:"email"`
	Phone    string    `json:"phone" db:"phone" form:"phone"`
	JobTitle string    `json:"job_title" db:"job_title" form:"job_title"`
	// CompanyID es la empresa del contacto; no válido significa «sin empresa».
	CompanyID uuid.NullUUID `json:"company_id" db:"company_id" form:"-"`
	// CompanyName se rellena al listar contactos; no se guarda.
	CompanyName string    `json:"-" db:"-" form:"-"`
	CreatedAt   time.Time `json:"created_at" db:"created_at" form:"-"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at" form:"-"`
}

// Contacts es una colección de Contact.
type Contacts []Contact

// Normalize recorta los espacios de todos los campos de texto.
func (c *Contact) Normalize() {
	c.Name = strings.TrimSpace(c.Name)
	c.Email = strings.TrimSpace(c.Email)
	c.Phone = strings.TrimSpace(c.Phone)
	c.JobTitle = strings.TrimSpace(c.JobTitle)
}

// Validate se ejecuta en cada llamada a los métodos "pop.Validate*".
func (c *Contact) Validate(tx *pop.Connection) (*validate.Errors, error) {
	verrs := validate.NewErrors()

	name := strings.TrimSpace(c.Name)
	if name == "" {
		verrs.Add("name", "El nombre es obligatorio.")
	} else if len([]rune(name)) > ContactNameMaxLength {
		verrs.Add("name", fmt.Sprintf("El nombre no puede superar los %d caracteres.", ContactNameMaxLength))
	}

	email := strings.TrimSpace(c.Email)
	if email != "" {
		if len([]rune(email)) > ContactEmailMaxLength {
			verrs.Add("email", fmt.Sprintf("El email no puede superar los %d caracteres.", ContactEmailMaxLength))
		} else if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email {
			verrs.Add("email", "El email no tiene un formato válido.")
		}
	}

	if len([]rune(strings.TrimSpace(c.Phone))) > ContactPhoneMaxLength {
		verrs.Add("phone", fmt.Sprintf("El teléfono no puede superar los %d caracteres.", ContactPhoneMaxLength))
	}

	if len([]rune(strings.TrimSpace(c.JobTitle))) > ContactJobTitleMaxLength {
		verrs.Add("job_title", fmt.Sprintf("El cargo no puede superar los %d caracteres.", ContactJobTitleMaxLength))
	}

	return verrs, nil
}
