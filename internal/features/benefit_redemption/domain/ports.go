package domain

// RedemptionRepository defines the secondary port for benefit_redemption data persistence.
type RedemptionRepository interface {
	Create(r *BenefitRedemption) error
	FindByCode(code string) (*BenefitRedemption, error)
	FindByUsuarioID(usuarioID uint) ([]BenefitRedemption, error)
	FindByEmpresaID(empresaID uint) ([]BenefitRedemption, error)
	FindAll() ([]BenefitRedemption, error)

	// MarkAsUsed atomically updates estado='usado', fecha_uso=now(), validado_por=negocioUserID
	// ONLY if the current estado='generado'. Returns rows affected (0 means race condition or already used).
	MarkAsUsed(redemptionID uint, validadoPor uint) (int64, error)

	// MarkAsExpired atomically updates estado='vencido' ONLY if current estado='generado'.
	// Returns rows affected for the same concurrency-safety reason as MarkAsUsed.
	MarkAsExpired(redemptionID uint) (int64, error)

	FindBenefitByID(id uint) (*BenefitForRedemption, error)
	GetUserEmpresaID(userID uint) (uint, error)
}

// RedemptionService defines the primary port for benefit redemption business operations.
type RedemptionService interface {
	RedeemBenefit(benefitID uint, userID uint) (*RedemptionResponse, error)
	ValidateCode(code string, negocioUserID uint) (*ValidateCodeResponse, error)
	GetMisRedenciones(userID uint) ([]*RedemptionResponse, error)
	GetRedemptionsByEmpresa(negocioUserID uint) ([]*RedemptionResponse, error)
	GetAllRedemptions() ([]*RedemptionResponse, error)
}
