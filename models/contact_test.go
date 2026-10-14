package models

import "strings"

func (ms *ModelSuite) Test_Contact_Create_Valid() {
	contact := &Contact{Name: "Laura Gómez", Email: "laura@example.com", Phone: "600 000 000", JobTitle: "CEO"}

	verrs, err := ms.DB.ValidateAndCreate(contact)
	ms.NoError(err)
	ms.False(verrs.HasAny())
	ms.NotZero(contact.ID)

	count, err := ms.DB.Count(&Contact{})
	ms.NoError(err)
	ms.Equal(1, count)
}

func (ms *ModelSuite) Test_Contact_OnlyNameIsRequired() {
	verrs, err := ms.DB.ValidateAndCreate(&Contact{Name: "Solo nombre"})
	ms.NoError(err)
	ms.False(verrs.HasAny())
}

func (ms *ModelSuite) Test_Contact_Name_Required() {
	for _, name := range []string{"", "   \t "} {
		verrs, err := ms.DB.ValidateAndCreate(&Contact{Name: name})
		ms.NoError(err)
		ms.True(verrs.HasAny())
		ms.Contains(verrs.Get("name"), "El nombre es obligatorio.")
	}

	count, err := ms.DB.Count(&Contact{})
	ms.NoError(err)
	ms.Equal(0, count)
}

func (ms *ModelSuite) Test_Contact_MaxLengths() {
	verrs, err := ms.DB.ValidateAndCreate(&Contact{Name: strings.Repeat("a", ContactNameMaxLength+1)})
	ms.NoError(err)
	ms.Contains(verrs.Get("name"), "El nombre no puede superar los 120 caracteres.")

	verrs, err = ms.DB.ValidateAndCreate(&Contact{Name: "X", Email: strings.Repeat("a", ContactEmailMaxLength) + "@example.com"})
	ms.NoError(err)
	ms.Contains(verrs.Get("email"), "El email no puede superar los 200 caracteres.")

	verrs, err = ms.DB.ValidateAndCreate(&Contact{Name: "X", Phone: strings.Repeat("1", ContactPhoneMaxLength+1)})
	ms.NoError(err)
	ms.Contains(verrs.Get("phone"), "El teléfono no puede superar los 40 caracteres.")

	verrs, err = ms.DB.ValidateAndCreate(&Contact{Name: "X", JobTitle: strings.Repeat("c", ContactJobTitleMaxLength+1)})
	ms.NoError(err)
	ms.Contains(verrs.Get("job_title"), "El cargo no puede superar los 120 caracteres.")
}

func (ms *ModelSuite) Test_Contact_Email_Format() {
	for _, email := range []string{"sin-arroba", "a@", "@example.com", "Ana <ana@example.com>", "dos palabras@example.com"} {
		verrs, err := ms.DB.ValidateAndCreate(&Contact{Name: "X", Email: email})
		ms.NoError(err)
		ms.Contains(verrs.Get("email"), "El email no tiene un formato válido.", email)
	}

	for _, email := range []string{"ana@example.com", "a.b+c@sub.example.org"} {
		verrs, err := ms.DB.ValidateAndCreate(&Contact{Name: "X", Email: email})
		ms.NoError(err)
		ms.False(verrs.HasAny(), email)
	}
}

func (ms *ModelSuite) Test_Contact_Normalize_TrimsFields() {
	contact := &Contact{Name: "  Ana  ", Email: " ana@example.com ", Phone: " 1 ", JobTitle: "\tJefa "}
	contact.Normalize()

	ms.Equal("Ana", contact.Name)
	ms.Equal("ana@example.com", contact.Email)
	ms.Equal("1", contact.Phone)
	ms.Equal("Jefa", contact.JobTitle)
}
