package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

var translations = map[string]map[string]string{
	"es": {
		"email_required":             "Por favor ingresa tu correo electrónico.",
		"invalid_email":              "Ingresa un formato de correo electrónico válido.",
		"password_required":          "Por favor ingresa tu contraseña.",
		"invalid_credentials":        "El correo electrónico o la contraseña son incorrectos.",
		"invalid_user_id":            "ID de usuario inválido",
		"user_not_found":             "Usuario no encontrado",
		"invalid_company_id":         "ID de empresa inválido",
		"company_not_found":          "Empresa no encontrada",
		"invalid_role_id":            "ID de rol inválido",
		"role_not_found":             "Rol no encontrado",
		"invalid_menu_id":            "ID de menú inválido",
		"menu_not_found":             "Menú no encontrado",
		"invalid_category_id":        "ID de categoría inválido",
		"category_not_found":         "Categoría no encontrada",
		"invalid_article_id":         "ID de artículo inválido",
		"article_not_found":          "Artículo no encontrado",
		"access_denied":              "Acceso denegado",
		"insufficient_permissions":   "Permisos insuficientes para esta operación",
		"failed_change_password":     "No se pudo cambiar la contraseña",
		"incorrect_current_password": "La contraseña actual es incorrecta",
		"passwords_do_not_match":     "Las contraseñas no coinciden",
		"no_file_uploaded":           "No se ha subido ningún archivo",
		"failed_save_file":           "Error al guardar el archivo",
		"company_header_required":    "El encabezado X-Company-ID es requerido",
		"invalid_company_header":     "Formato de X-Company-ID inválido",
		"failed_verify_company":      "Error al verificar los permisos de la empresa",
		"duplicate_entry":            "Ya existe un registro con esos datos en el sistema.",
		"record_not_found":           "El registro solicitado no existe.",
		"field_required":             "Uno o más campos obligatorios están vacíos.",
		"business_subscription_expired": "La suscripción de tu empresa se encuentra inactiva o vencida. Para crear nuevos beneficios y publicar promociones, por favor activa tu suscripción comunicándote con el administrador de Conexiate.",
		"membership_not_active":         "Membresía inactiva. Activa tu membresía mensual Conexiate para disfrutar de los beneficios y descuentos de comercios aliados.",
		"company_nit_duplicate":         "Ya existe una empresa registrada con este NIT.",
		"company_razon_social_duplicate": "Ya existe una empresa registrada con esta razón social.",
		"company_name_duplicate":         "Ya existe una empresa registrada con este nombre comercial.",
		"email_duplicate":                "El correo electrónico ya se encuentra registrado en la plataforma.",
		"company_has_active_benefits":    "No es posible eliminar la empresa porque tiene beneficios activos asociados. Debe desactivarlos o eliminarlos primero.",
		"company_has_active_users":       "No es posible eliminar la empresa porque tiene validadores de negocio o usuarios activos asociados. Debe desactivarlos o reasignarlos primero.",
		"user_inactive":                  "Tu cuenta se encuentra inactiva. Por favor, revisa tu correo de activación o contacta al administrador.",
		"user_not_found_email":           "No existe ningún usuario registrado con este correo electrónico.",
	},
	"en": {
		"email_required":             "Email is required",
		"invalid_email":              "Please enter a valid email",
		"password_required":          "Password is required",
		"invalid_credentials":        "Invalid credentials",
		"invalid_user_id":            "Invalid user ID",
		"user_not_found":             "User not found",
		"invalid_company_id":         "Invalid company ID",
		"company_not_found":          "Company not found",
		"invalid_role_id":            "Invalid role ID",
		"role_not_found":             "Role not found",
		"invalid_menu_id":            "Invalid menu ID",
		"menu_not_found":             "Menu not found",
		"invalid_category_id":        "Invalid category ID",
		"category_not_found":         "Category not found",
		"invalid_article_id":         "Invalid article ID",
		"article_not_found":          "Article not found",
		"access_denied":              "Access denied",
		"insufficient_permissions":   "Insufficient permissions for this operation",
		"failed_change_password":     "Failed to change password",
		"incorrect_current_password": "Current password is incorrect",
		"passwords_do_not_match":     "Passwords do not match",
		"no_file_uploaded":           "No file uploaded",
		"failed_save_file":           "Failed to save file",
		"company_header_required":    "X-Company-ID header is required",
		"invalid_company_header":     "Invalid X-Company-ID format",
		"failed_verify_company":      "Failed to verify company permissions",
		"duplicate_entry":            "An entry with these details already exists.",
		"record_not_found":           "The requested record does not exist.",
		"field_required":             "One or more required fields are empty.",
		"business_subscription_expired": "Your company subscription is inactive or expired. To create new benefits, please activate your subscription by contacting the administrator.",
		"membership_not_active":         "Inactive membership. Please activate your monthly Conexiate membership to access all benefits and discounts from allied businesses.",
		"company_nit_duplicate":         "A company is already registered with this NIT.",
		"company_razon_social_duplicate": "A company is already registered with this business name (razón social).",
		"company_name_duplicate":         "A company is already registered with this commercial name.",
		"email_duplicate":                "This email address is already registered on the platform.",
		"company_has_active_benefits":    "Cannot delete company because it has active benefits associated. Please deactivate or delete them first.",
		"company_has_active_users":       "Cannot delete company because it has active business validators or users associated. Please deactivate or reassign them first.",
		"user_inactive":                  "The user account is inactive.",
		"user_not_found_email":           "No user registered with this email address.",
	},
	"fr": {
		"email_required":             "L'adresse e-mail est requise",
		"invalid_email":              "Veuillez entrer une adresse e-mail valide",
		"password_required":          "Le mot de passe est requis",
		"invalid_credentials":        "Identifiants invalides",
		"invalid_user_id":            "ID utilisateur invalide",
		"user_not_found":             "Utilisateur non trouvé",
		"invalid_company_id":         "ID de l'entreprise invalide",
		"company_not_found":          "Entreprise non trouvée",
		"invalid_role_id":            "ID du rôle invalide",
		"role_not_found":             "Rôle non trouvé",
		"invalid_menu_id":            "ID du menu invalide",
		"menu_not_found":             "Menu non trouvé",
		"invalid_category_id":        "ID de catégorie invalide",
		"category_not_found":         "Catégorie non trouvée",
		"invalid_article_id":         "ID de l'article invalide",
		"article_not_found":          "Article non trouvé",
		"access_denied":              "Accès refusé",
		"insufficient_permissions":   "Permissions insuffisantes pour cette opération",
		"failed_change_password":     "Échec du changement de mot de passe",
		"incorrect_current_password": "Le mot de passe actuel est incorrect",
		"passwords_do_not_match":     "Les mots de passe ne correspondent pas",
		"no_file_uploaded":           "Aucun fichier téléchargé",
		"failed_save_file":           "Échec de l'enregistrement du fichier",
		"company_header_required":    "L'en-tête X-Company-ID est requis",
		"invalid_company_header":     "Format de X-Company-ID invalide",
		"failed_verify_company":      "Échec de la vérification des autorisations de l'entreprise",
		"duplicate_entry":            "Une entrée avec ces détails existe déjà",
		"record_not_found":           "Le registre demandé n'existe pas",
		"field_required":             "Un ou plusieurs champs requis sont vides",
	},
}

func init() {
	// Attempt to load from JSON files in backend/i18n directory
	for _, lang := range []string{"es", "en", "fr"} {
		filePath := filepath.Join("i18n", lang+".json")
		if _, err := os.Stat(filePath); err == nil {
			if file, err := os.Open(filePath); err == nil {
				defer file.Close()
				var data map[string]string
				if err := json.NewDecoder(file).Decode(&data); err == nil {
					for k, v := range data {
						translations[lang][k] = v
					}
				}
			}
		}
	}
}

// T translates a key based on Accept-Language header
func T(c *gin.Context, key string, defaultMsg string) string {
	lang := strings.ToLower(c.GetHeader("Accept-Language"))
	if lang == "" {
		lang = "es"
	}
	if len(lang) > 2 {
		lang = lang[:2]
	}

	if translationsForLang, ok := translations[lang]; ok {
		if val, exists := translationsForLang[key]; exists {
			return val
		}
	}

	// Fallback to "es"
	if translationsForLang, ok := translations["es"]; ok {
		if val, exists := translationsForLang[key]; exists {
			return val
		}
	}

	return defaultMsg
}

// TranslateError translates an error to the requested language
func TranslateError(c *gin.Context, err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()

	// Direct match mappings
	switch msg {
	case "invalid credentials":
		return T(c, "invalid_credentials", "El correo electrónico o la contraseña son incorrectos.")
	case "user account is inactive", "la cuenta de usuario se encuentra inactiva":
		return T(c, "user_inactive", "Tu cuenta se encuentra inactiva. Por favor, revisa tu correo de activación o contacta al administrador.")
	case "la contraseña actual y la confirmación no coinciden":
		return T(c, "passwords_do_not_match", "Las contraseñas no coinciden")
	case "la contraseña actual es incorrecta":
		return T(c, "incorrect_current_password", "La contraseña actual es incorrecta")
	case "usuario no encontrado", "user not found":
		return T(c, "user_not_found", "Usuario no encontrado")
	case "no se encontró ningún usuario registrado con este correo electrónico", "no existe ningún usuario registrado con este correo electrónico":
		return T(c, "user_not_found_email", "No existe ningún usuario registrado con este correo electrónico.")
	case "failed to generate token":
		return T(c, "failed_generate_token", "Error al generar el token de sesión")
	case "No file uploaded":
		return T(c, "no_file_uploaded", "No se ha subido ningún archivo")
	case "Failed to save file":
		return T(c, "failed_save_file", "Error al guardar el archivo")
	case "business_subscription_expired", "business subscription is expired. Contact your administrator":
		return T(c, "business_subscription_expired", "La suscripción de tu empresa se encuentra inactiva o vencida. Para crear nuevos beneficios y publicar promociones, por favor activa tu suscripción comunicándote con el administrador de Conexiate.")
	case "membership_not_active", "membership is not active. Please activate your membership to access this feature":
		return T(c, "membership_not_active", "Membresía inactiva. Activa tu membresía mensual Conexiate para disfrutar de los beneficios y descuentos de comercios aliados.")
	case "invalid company ID":
		return T(c, "invalid_company_id", "ID de empresa inválido")
	case "invalid user ID":
		return T(c, "invalid_user_id", "ID de usuario inválido")
	case "invalid menu ID":
		return T(c, "invalid_menu_id", "ID de menú inválido")
	case "invalid category ID":
		return T(c, "invalid_category_id", "ID de categoría inválido")
	case "invalid article ID":
		return T(c, "invalid_article_id", "ID de artículo inválido")
	case "authorization token is required":
		return T(c, "authorization_required", "authorization token is required")
	case "invalid or expired token":
		return T(c, "invalid_or_expired_token", "invalid or expired token")
	case "invalid token claims":
		return T(c, "invalid_token_claims", "invalid token claims")
	case "X-Company-ID header is required":
		return T(c, "company_header_required", "El encabezado X-Company-ID es requerido")
	case "invalid X-Company-ID format":
		return T(c, "invalid_company_header", "Formato de X-Company-ID inválido")
	case "failed to verify company permissions":
		return T(c, "failed_verify_company", "Error al verificar los permisos de la empresa")
	case "access denied to this company":
		return T(c, "access_denied", "Acceso denegado")
	}

	lower := strings.ToLower(msg)
	if strings.Contains(lower, "duplicate key") || strings.Contains(lower, "unique constraint") || strings.Contains(lower, "23505") || strings.Contains(lower, "1062") {
		if strings.Contains(lower, "nit") {
			return T(c, "company_nit_duplicate", "Ya existe una empresa registrada con este NIT.")
		}
		if strings.Contains(lower, "razon_social") {
			return T(c, "company_razon_social_duplicate", "Ya existe una empresa registrada con esta razón social.")
		}
		if strings.Contains(lower, "name") || strings.Contains(lower, "companies") {
			return T(c, "company_name_duplicate", "Ya existe una empresa registrada con este nombre comercial.")
		}
		if strings.Contains(lower, "email") || strings.Contains(lower, "users_email") {
			return T(c, "email_duplicate", "El correo electrónico ya se encuentra registrado en la plataforma.")
		}
		return T(c, "duplicate_entry", "Ya existe un registro con esos datos en el sistema.")
	}
	if strings.Contains(lower, "record not found") {
		return T(c, "record_not_found", "El registro solicitado no existe.")
	}
	if strings.Contains(lower, "required") {
		return T(c, "field_required", "Uno o más campos obligatorios están vacíos.")
	}
	if strings.Contains(lower, "email") {
		return T(c, "invalid_email", "Ingrese un correo electrónico válido.")
	}

	return msg
}

// Error translates an error and responds with status code and JSON {"error": ...}
func Error(c *gin.Context, status int, err error) {
	c.JSON(status, gin.H{"error": TranslateError(c, err)})
}

// ErrorString translates a string message and responds with status code and JSON {"error": ...}
func ErrorString(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": T(c, msg, msg)})
}
