package validation

import (
	"errors"
	"strings"
	"unicode"
)

var (
	ErrPasswordDemasiadoCorta = errors.New("la contraseña debe tener al menos 10 caracteres")
	ErrPasswordSinMayuscula   = errors.New("la contraseña debe contener al menos una letra mayúscula")
	ErrPasswordSinMinuscula   = errors.New("la contraseña debe contener al menos una letra minúscula")
	ErrPasswordSinNumero      = errors.New("la contraseña debe contener al menos un número")
	ErrPasswordIgualEmail     = errors.New("la contraseña no puede ser igual al correo electrónico")
	ErrPasswordContieneNombre = errors.New("la contraseña no puede contener tu nombre")
)

// ValidarPassword aplica las políticas de seguridad de contraseñas de Conexiate:
// - Longitud mínima de 10 caracteres
// - Al menos una letra mayúscula
// - Al menos una letra minúscula
// - Al menos un número
// - No puede ser idéntica al correo electrónico
// - No puede contener el nombre del usuario (si este tiene 3 o más caracteres)
func ValidarPassword(password, email, nombre string) error {
	if len(password) < 10 {
		return ErrPasswordDemasiadoCorta
	}

	var tieneMayuscula, tieneMinuscula, tieneNumero bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			tieneMayuscula = true
		case unicode.IsLower(char):
			tieneMinuscula = true
		case unicode.IsDigit(char):
			tieneNumero = true
		}
	}

	if !tieneMayuscula {
		return ErrPasswordSinMayuscula
	}
	if !tieneMinuscula {
		return ErrPasswordSinMinuscula
	}
	if !tieneNumero {
		return ErrPasswordSinNumero
	}

	pwdLower := strings.ToLower(password)

	if email != "" && pwdLower == strings.ToLower(strings.TrimSpace(email)) {
		return ErrPasswordIgualEmail
	}

	nombreLimpio := strings.ToLower(strings.TrimSpace(nombre))
	if len(nombreLimpio) >= 3 {
		// Verificar si contiene el nombre completo o alguna de sus palabras principales (ej. primer nombre)
		palabras := strings.Fields(nombreLimpio)
		for _, p := range palabras {
			if len(p) >= 3 && strings.Contains(pwdLower, p) {
				return ErrPasswordContieneNombre
			}
		}
	}

	return nil
}
