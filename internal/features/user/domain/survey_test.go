package domain

import (
	"encoding/json"
	"testing"
)

func TestUserSurveyFieldsSerialization(t *testing.T) {
	familiaresExt := true
	vivienda := "Propia"
	esEmp := true
	desc := "Taller de marroquinería y cuero"

	user := &User{
		Email:                     "user@conexiate.com",
		FirstName:                 "Laura",
		LastName:                  "Restrepo",
		FamiliaresExterior:        &familiaresExt,
		ViviendaTipo:              &vivienda,
		EsEmprendedor:             &esEmp,
		DescripcionEmprendimiento: &desc,
	}

	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Failed to marshal User: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal User JSON: %v", err)
	}

	if parsed["familiares_exterior"] != true {
		t.Errorf("Expected familiares_exterior = true, got %v", parsed["familiares_exterior"])
	}
	if parsed["vivienda_tipo"] != "Propia" {
		t.Errorf("Expected vivienda_tipo = 'Propia', got %v", parsed["vivienda_tipo"])
	}
	if parsed["es_emprendedor"] != true {
		t.Errorf("Expected es_emprendedor = true, got %v", parsed["es_emprendedor"])
	}
	if parsed["descripcion_emprendimiento"] != desc {
		t.Errorf("Expected descripcion_emprendimiento = %q, got %v", desc, parsed["descripcion_emprendimiento"])
	}
}
