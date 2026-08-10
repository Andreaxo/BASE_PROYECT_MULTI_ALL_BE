package infrastructure

import (
	"gorm.io/gorm"

	"multicliente-backend/internal/features/user/domain"
	"multicliente-backend/internal/platform/database"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new GORM-based UserRepository.
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) populateAudits(users []*domain.User) {
	var userIDs []uint
	for _, u := range users {
		if u.CreateBy != nil {
			userIDs = append(userIDs, *u.CreateBy)
		}
		if u.UpdateBy != nil {
			userIDs = append(userIDs, *u.UpdateBy)
		}
	}
	namesMap, err := database.GetUserNamesMap(r.db, userIDs)
	if err != nil {
		return
	}
	for _, u := range users {
		if u.CreateBy != nil {
			u.CreateByName = namesMap[*u.CreateBy]
		}
		if u.UpdateBy != nil {
			u.UpdateByName = namesMap[*u.UpdateBy]
		}
	}
}

func (r *userRepository) Create(user *domain.User) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *userRepository) FindByID(id uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.Preload("Role").Preload("Companies").First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	r.populateAudits([]*domain.User{&user})
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Preload("Role").Preload("Companies").First(&user, "email = ?", email).Error; err != nil {
		return nil, err
	}
	r.populateAudits([]*domain.User{&user})
	return &user, nil
}

func (r *userRepository) FindByCodeRefer(code string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Preload("Role").Preload("Companies").First(&user, "code_refer = ?", code).Error; err != nil {
		return nil, err
	}
	r.populateAudits([]*domain.User{&user})
	return &user, nil
}

func (r *userRepository) FindAll() ([]domain.User, error) {
	var users []domain.User
	if err := r.db.Preload("Role").Preload("Companies").Order("create_at DESC").Find(&users).Error; err != nil {
		return nil, err
	}
	usersPtrs := make([]*domain.User, len(users))
	for i := range users {
		usersPtrs[i] = &users[i]
	}
	r.populateAudits(usersPtrs)
	return users, nil
}

func (r *userRepository) Update(user *domain.User) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Save basic user details
		if err := tx.Save(user).Error; err != nil {
			return err
		}

		// Sync many-to-many Companies association
		// By doing this within transaction, we ensure the join table matches user.Companies
		if err := tx.Model(user).Association("Companies").Replace(user.Companies); err != nil {
			return err
		}

		return nil
	})
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var user domain.User
		if err := tx.First(&user, "id = ?", id).Error; err != nil {
			return err
		}

		// Clear company associations first
		if err := tx.Model(&user).Association("Companies").Clear(); err != nil {
			return err
		}

		// Delete the user
		if err := tx.Delete(&user).Error; err != nil {
			return err
		}
		return nil
	})
}
