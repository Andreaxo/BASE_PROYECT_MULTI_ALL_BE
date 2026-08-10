package domain

type BenefitRepository interface {
	Create(benefit *Benefit) error
	FindByID(id uint) (*Benefit, error)
	FindAll() ([]Benefit, error)
	FindAllByCompany(companyID uint) ([]Benefit, error)
	Update(benefit *Benefit) error
	Delete(id uint) error
}

type BenefitService interface {
	CreateBenefit(req *CreateBenefitRequest) (*Benefit, error)
	GetBenefitByID(id uint) (*Benefit, error)
	GetAllBenefits() ([]Benefit, error)
	GetBenefitsByCompany(companyID uint) ([]Benefit, error)
	UpdateBenefit(id uint, req *UpdateBenefitRequest) (*Benefit, error)
	DeleteBenefit(id uint) error
}
