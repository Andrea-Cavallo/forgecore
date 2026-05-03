package validation

import "testing"

type sampleInput struct {
	Email string `validate:"required,email"`
}

func TestValidateReturnsFieldErrors(t *testing.T) {
	err := Validate(sampleInput{Email: "non-email"})
	if err == nil {
		t.Fatal("atteso errore validazione")
	}
	if _, ok := err.(FieldErrors); !ok {
		t.Fatalf("tipo errore inatteso: %T", err)
	}
}
