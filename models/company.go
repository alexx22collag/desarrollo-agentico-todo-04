package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// Longitudes máximas de los campos de una empresa.
const (
	CompanyNameMaxLength    = 150
	CompanyWebsiteMaxLength = 200
	CompanySectorMaxLength  = 100
)

// Company representa una empresa del CRM.
type Company struct {
	ID        uuid.UUID `json:"id" db:"id" form:"-"`
	Name      string    `json:"name" db:"name" form:"name"`
	Website   string    `json:"website" db:"website" form:"website"`
	Sector    string    `json:"sector" db:"sector" form:"sector"`
	CreatedAt time.Time `json:"created_at" db:"created_at" form:"-"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" form:"-"`
}

// Companies es una colección de Company.
type Companies []Company

// Normalize recorta los espacios de todos los campos de texto.
func (c *Company) Normalize() {
	c.Name = strings.TrimSpace(c.Name)
	c.Website = strings.TrimSpace(c.Website)
	c.Sector = strings.TrimSpace(c.Sector)
}

// Validate se ejecuta en cada llamada a los métodos "pop.Validate*".
func (c *Company) Validate(tx *pop.Connection) (*validate.Errors, error) {
	verrs := validate.NewErrors()

	name := strings.TrimSpace(c.Name)
	switch {
	case name == "":
		verrs.Add("name", "El nombre es obligatorio.")
	case len([]rune(name)) > CompanyNameMaxLength:
		verrs.Add("name", fmt.Sprintf("El nombre no puede superar los %d caracteres.", CompanyNameMaxLength))
	default:
		exists, err := tx.Where("lower(name) = lower(?) AND id <> ?", name, c.ID).Exists(&Company{})
		if err != nil {
			return verrs, err
		}
		if exists {
			verrs.Add("name", "Ya existe una empresa con ese nombre.")
		}
	}

	if len([]rune(strings.TrimSpace(c.Website))) > CompanyWebsiteMaxLength {
		verrs.Add("website", fmt.Sprintf("La web no puede superar los %d caracteres.", CompanyWebsiteMaxLength))
	}

	if len([]rune(strings.TrimSpace(c.Sector))) > CompanySectorMaxLength {
		verrs.Add("sector", fmt.Sprintf("El sector no puede superar los %d caracteres.", CompanySectorMaxLength))
	}

	return verrs, nil
}
