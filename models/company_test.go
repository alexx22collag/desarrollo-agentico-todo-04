package models

import "strings"

func (ms *ModelSuite) Test_Company_Create_Valid() {
	company := &Company{Name: "Initech", Website: "https://initech.example.com", Sector: "Software"}

	verrs, err := ms.DB.ValidateAndCreate(company)
	ms.NoError(err)
	ms.False(verrs.HasAny())
	ms.NotZero(company.ID)
}

func (ms *ModelSuite) Test_Company_Name_Required() {
	for _, name := range []string{"", "  \t "} {
		verrs, err := ms.DB.ValidateAndCreate(&Company{Name: name})
		ms.NoError(err)
		ms.Contains(verrs.Get("name"), "El nombre es obligatorio.")
	}
}

func (ms *ModelSuite) Test_Company_MaxLengths() {
	verrs, err := ms.DB.ValidateAndCreate(&Company{Name: strings.Repeat("a", CompanyNameMaxLength+1)})
	ms.NoError(err)
	ms.Contains(verrs.Get("name"), "El nombre no puede superar los 150 caracteres.")

	verrs, err = ms.DB.ValidateAndCreate(&Company{Name: "X", Website: strings.Repeat("w", CompanyWebsiteMaxLength+1)})
	ms.NoError(err)
	ms.Contains(verrs.Get("website"), "La web no puede superar los 200 caracteres.")

	verrs, err = ms.DB.ValidateAndCreate(&Company{Name: "X", Sector: strings.Repeat("s", CompanySectorMaxLength+1)})
	ms.NoError(err)
	ms.Contains(verrs.Get("sector"), "El sector no puede superar los 100 caracteres.")
}

func (ms *ModelSuite) Test_Company_Name_UniqueIgnoringCase() {
	ms.NoError(ms.DB.Create(&Company{Name: "Acme"}))

	verrs, err := ms.DB.ValidateAndCreate(&Company{Name: "  ACME "})
	ms.NoError(err)
	ms.Contains(verrs.Get("name"), "Ya existe una empresa con ese nombre.")

	count, err := ms.DB.Count(&Company{})
	ms.NoError(err)
	ms.Equal(1, count)
}

func (ms *ModelSuite) Test_Company_Update_KeepsOwnName() {
	company := &Company{Name: "Acme"}
	ms.NoError(ms.DB.Create(company))

	company.Sector = "Industria"
	verrs, err := ms.DB.ValidateAndUpdate(company)
	ms.NoError(err)
	ms.False(verrs.HasAny())
}

func (ms *ModelSuite) Test_Company_Delete_LeavesContactsWithoutCompany() {
	ms.LoadFixture("empresas con contactos")

	company := &Company{}
	ms.NoError(ms.DB.Where("name = ?", "Acme S.L.").First(company))
	ms.NoError(ms.DB.Destroy(company))

	count, err := ms.DB.Count(&Contact{})
	ms.NoError(err)
	ms.Equal(3, count)

	contact := &Contact{}
	ms.NoError(ms.DB.Where("name = ?", "Ana Acme").First(contact))
	ms.False(contact.CompanyID.Valid)
}
