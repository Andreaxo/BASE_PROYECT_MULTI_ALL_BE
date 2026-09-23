package validation

import (
	"testing"
)

func TestValidarPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		email       string
		nombre      string
		expectedErr error
	}{
		{
			name:        "Válida normal",
			password:    "Segura12345!",
			email:       "juan@example.com",
			nombre:      "Carlos Gomez",
			expectedErr: nil,
		},
		{
			name:        "Corta (menos de 10 caracteres)",
			password:    "Abc1234",
			email:       "juan@example.com",
			nombre:      "Carlos",
			expectedErr: ErrPasswordDemasiadoCorta,
		},
		{
			name:        "Sin mayúscula",
			password:    "contraseñasegura123",
			email:       "juan@example.com",
			nombre:      "Carlos",
			expectedErr: ErrPasswordSinMayuscula,
		},
		{
			name:        "Sin minúscula",
			password:    "CONTRASENA12345",
			email:       "juan@example.com",
			nombre:      "Carlos",
			expectedErr: ErrPasswordSinMinuscula,
		},
		{
			name:        "Sin número",
			password:    "ContrasenaSegura",
			email:       "juan@example.com",
			nombre:      "Carlos",
			expectedErr: ErrPasswordSinNumero,
		},
		{
			name:        "Igual al email",
			password:    "Juan12345@Example.Com",
			email:       "juan12345@example.com",
			nombre:      "Pedro",
			expectedErr: ErrPasswordIgualEmail,
		},
		{
			name:        "Contiene el nombre",
			password:    "MiCarlos12345",
			email:       "juan@example.com",
			nombre:      "Carlos Gomez",
			expectedErr: ErrPasswordContieneNombre,
		},
		{
			name:        "Contiene el apellido",
			password:    "SuperGomez99!",
			email:       "juan@example.com",
			nombre:      "Carlos Gomez",
			expectedErr: ErrPasswordContieneNombre,
		},
		{
			name:        "Nombre corto (menor a 3 letras) no genera falso positivo",
			password:    "Alonso12345!",
			email:       "juan@example.com",
			nombre:      "Al",
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidarPassword(tt.password, tt.email, tt.nombre)
			if err != tt.expectedErr {
				t.Errorf("ValidarPassword(%q, %q, %q) = %v; esperado %v",
					tt.password, tt.email, tt.nombre, err, tt.expectedErr)
			}
		})
	}
}
