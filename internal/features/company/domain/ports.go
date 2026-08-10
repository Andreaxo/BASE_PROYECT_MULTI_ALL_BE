package domain

type CompanyRepository interface {
	Create(company *Company, req *CreateCompanyRequest) error
	FindByID(id uint) (*Company, error)
	FindAll() ([]Company, error)
	Update(company *Company) error
	Delete(id uint) error
}

type CompanyService interface {
	CreateCompany(req *CreateCompanyRequest, createdBy *uint) (*Company, error)
	GetCompanyByID(id uint) (*Company, error)
	GetAllCompanies() ([]Company, error)
	UpdateCompany(id uint, req *UpdateCompanyRequest, updatedBy *uint) (*Company, error)
	DeleteCompany(id uint) error
}
