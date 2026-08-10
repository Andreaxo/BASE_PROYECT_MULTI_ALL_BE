package application

import (
	"errors"
	"multicliente-backend/internal/features/benefit/domain"
)

type benefitService struct {
	repo domain.BenefitRepository
}

func NewBenefitService(repo domain.BenefitRepository) domain.BenefitService {
	return &benefitService{repo: repo}
}

func (s *benefitService) CreateBenefit(req *domain.CreateBenefitRequest) (*domain.Benefit, error) {
	if req.Name == "" {
		return nil, errors.New("Name is required")
	}
	if req.Description == "" {
		return nil, errors.New("Description is required")
	}
	if req.CompanyBenefits == nil {
		return nil, errors.New("company benefits is required")
	}
	benefit := &domain.Benefit{
		Name:            req.Name,
		Description:     req.Description,
		CompanyBenefits: req.CompanyBenefits,
	}

	if err := s.repo.Create(benefit); err != nil {
		return nil, err
	}

	return benefit, nil
}

func (s *benefitService) GetBenefitByID(id uint) (*domain.Benefit, error) {
	benefit, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("benefit not found")
	}

	return benefit, nil
}

func (s *benefitService) GetAllBenefits() ([]domain.Benefit, error) {
	return s.repo.FindAll()
}

func (s *benefitService) GetBenefitsByCompany(companyID uint) ([]domain.Benefit, error) {
	return s.repo.FindAllByCompany(companyID)
}

func (s *benefitService) UpdateBenefit(id uint, req *domain.UpdateBenefitRequest) (*domain.Benefit, error) {
	benefit, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("benefit not found")
	}

	if req.Name != nil {
		benefit.Name = *req.Name
	}

	if req.Description != nil {
		benefit.Description = *req.Description
	}

	if req.CompanyBenefits != nil && *req.CompanyBenefits > 0 {
		benefit.CompanyBenefits = req.CompanyBenefits
	}

	if err := s.repo.Update(benefit); err != nil {
		return nil, err
	}

	return benefit, nil
}

func (s *benefitService) DeleteBenefit(id uint) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("benefit not found")
	}

	return s.repo.Delete(id)
}
