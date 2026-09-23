package infrastructure

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"multicliente-backend/internal/features/user/domain"
	"multicliente-backend/internal/platform/i18n"
)

// UserHandler handles HTTP requests for user CRUD operations.
type UserHandler struct {
	service domain.UserService
}

// NewUserHandler creates a new UserHandler with the given service.
func NewUserHandler(service domain.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// Create handles POST /api/users
func (h *UserHandler) Create(c *gin.Context) {
	var req domain.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	roleVal, _ := c.Get("role")
	roleCode, _ := roleVal.(string)
	if roleCode != "superadmin" {
		if req.RoleID != nil && *req.RoleID == 1 {
			c.JSON(http.StatusForbidden, gin.H{"error": "No tiene permisos para crear usuarios con el rol de Super Administrador"})
			return
		}
	}

	createdBy := getUserIDFromContext(c)

	response, err := h.service.CreateUser(&req, createdBy)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetAll handles GET /api/users
func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.service.GetAllUsers()
	if err != nil {
		i18n.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetByID handles GET /api/users/:id
func (h *UserHandler) GetByID(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid user ID")
		return
	}
	id := uint(idVal)

	user, err := h.service.GetUserByID(id)
	if err != nil {
		i18n.Error(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// Update handles PUT /api/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid user ID")
		return
	}
	id := uint(idVal)

	roleVal, _ := c.Get("role")
	roleCode, _ := roleVal.(string)

	if roleCode != "superadmin" {
		targetUser, err := h.service.GetUserByID(id)
		if err == nil && targetUser != nil && targetUser.RoleID != nil && *targetUser.RoleID == 1 {
			c.JSON(http.StatusForbidden, gin.H{"error": "No tiene permisos para modificar a un Super Administrador"})
			return
		}
	}

	var req domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	if roleCode != "superadmin" {
		if req.RoleID != nil && *req.RoleID == 1 {
			c.JSON(http.StatusForbidden, gin.H{"error": "No tiene permisos para asignar el rol de Super Administrador"})
			return
		}
	}

	updatedBy := getUserIDFromContext(c)

	response, err := h.service.UpdateUser(id, &req, updatedBy)
	if err != nil {
		i18n.Error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// Delete handles DELETE /api/users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		i18n.ErrorString(c, http.StatusBadRequest, "invalid user ID")
		return
	}
	id := uint(idVal)

	if err := h.service.DeleteUser(id); err != nil {
		i18n.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

// getUserIDFromContext extracts the authenticated user's uint ID from the Gin context.
func getUserIDFromContext(c *gin.Context) *uint {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		return nil
	}

	// It could be float64 when decoded from JWT claims map
	if f, ok := userIDVal.(float64); ok {
		u := uint(f)
		return &u
	}

	// Or string
	if s, ok := userIDVal.(string); ok {
		idVal, err := strconv.ParseUint(s, 10, 32)
		if err == nil {
			u := uint(idVal)
			return &u
		}
	}

	return nil
}
