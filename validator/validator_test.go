package validator

import (
	"errors"
	"testing"
)

// ── Test fixtures ─────────────────────────────────────────────────────────────

// basicUser mirrors V1 tests so all existing tests remain green.
type basicUser struct {
	Name  string `required:"true"`
	Email string `required:"true"`
	Age   int
}

// fullUser exercises every V2 tag in combination.
type fullUser struct {
	Name  string  `required:"true"`
	Age   int     `min:"18" max:"120"`
	Score float64 `min:"0" max:"100"`
	Email string  `required:"true" email:"true"`
}

// minOnlyUser is used to isolate min tests.
type minOnlyUser struct {
	Age int `min:"18"`
}

// maxOnlyUser is used to isolate max tests.
type maxOnlyUser struct {
	Score int `max:"100"`
}

// emailOnlyUser is used to isolate email tests.
type emailOnlyUser struct {
	Email string `email:"true"`
}

// unsignedUser tests min/max on unsigned integer kinds.
type unsignedUser struct {
	Level uint `min:"1" max:"10"`
}

// floatUser tests min/max on float kinds.
type floatUser struct {
	Score float64 `min:"0" max:"100"`
}

// noTagUser has no validation tags; Validate must return nil.
type noTagUser struct {
	Name string
	Age  int
}

// unexportedUser ensures unexported fields are skipped without panic.
type unexportedUser struct {
	Name    string `required:"true"`
	secret  string `required:"true"` //nolint:unused // intentionally unexported
}

// ── Helper ────────────────────────────────────────────────────────────────────

// assertValidationError extracts *ValidationError from err (using errors.As)
// and checks that Field and Rule match the expected values.
func assertValidationError(t *testing.T, err error, wantField, wantRule string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected a ValidationError{Field:%q Rule:%q}, got nil", wantField, wantRule)
	}
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
	if ve.Field != wantField {
		t.Errorf("ValidationError.Field: want %q, got %q", wantField, ve.Field)
	}
	if ve.Rule != wantRule {
		t.Errorf("ValidationError.Rule: want %q, got %q", wantRule, ve.Rule)
	}
}

// ── V1 backward-compatibility tests ──────────────────────────────────────────

func TestValidate_V1_Success(t *testing.T) {
	u := basicUser{Name: "Ganesh", Email: "ganesh@gmail.com", Age: 22}
	if err := Validate(u); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidate_V1_RequiredString_Empty(t *testing.T) {
	u := basicUser{Name: "", Email: "ganesh@gmail.com"}
	err := Validate(u)
	assertValidationError(t, err, "Name", "required")
}

func TestValidate_V1_RequiredAllEmpty(t *testing.T) {
	u := basicUser{}
	err := Validate(u)
	assertValidationError(t, err, "Name", "required")
}

func TestValidate_V1_Pointer_Valid(t *testing.T) {
	u := &basicUser{Name: "Ganesh", Email: "ganesh@gmail.com"}
	if err := Validate(u); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidate_V1_Nil(t *testing.T) {
	if err := Validate(nil); err == nil {
		t.Fatal("expected error for nil input")
	}
}

func TestValidate_V1_NonStruct_String(t *testing.T) {
	if err := Validate("hello"); err == nil {
		t.Fatal("expected error for string input")
	}
}

func TestValidate_V1_NonStruct_Int(t *testing.T) {
	if err := Validate(100); err == nil {
		t.Fatal("expected error for int input")
	}
}

func TestValidate_V1_NilPointer(t *testing.T) {
	var u *basicUser
	if err := Validate(u); err == nil {
		t.Fatal("expected error for nil pointer")
	}
}

// ── required rule ─────────────────────────────────────────────────────────────

func TestValidate_Required_Int_Zero(t *testing.T) {
	type s struct {
		Count int `required:"true"`
	}
	err := Validate(s{Count: 0})
	assertValidationError(t, err, "Count", "required")
}

func TestValidate_Required_Int_NonZero(t *testing.T) {
	type s struct {
		Count int `required:"true"`
	}
	if err := Validate(s{Count: 1}); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidate_Required_Bool_False(t *testing.T) {
	type s struct {
		Active bool `required:"true"`
	}
	err := Validate(s{Active: false})
	assertValidationError(t, err, "Active", "required")
}

func TestValidate_Required_Bool_True(t *testing.T) {
	type s struct {
		Active bool `required:"true"`
	}
	if err := Validate(s{Active: true}); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

// ── min rule ──────────────────────────────────────────────────────────────────

func TestValidate_Min_BelowLimit(t *testing.T) {
	err := Validate(minOnlyUser{Age: 17})
	assertValidationError(t, err, "Age", "min")
}

func TestValidate_Min_AtLimit(t *testing.T) {
	if err := Validate(minOnlyUser{Age: 18}); err != nil {
		t.Fatalf("expected nil at boundary, got %v", err)
	}
}

func TestValidate_Min_AboveLimit(t *testing.T) {
	if err := Validate(minOnlyUser{Age: 25}); err != nil {
		t.Fatalf("expected nil above limit, got %v", err)
	}
}

func TestValidate_Min_NegativeValue(t *testing.T) {
	err := Validate(minOnlyUser{Age: -5})
	assertValidationError(t, err, "Age", "min")
}

func TestValidate_Min_Uint_BelowLimit(t *testing.T) {
	err := Validate(unsignedUser{Level: 0})
	assertValidationError(t, err, "Level", "min")
}

func TestValidate_Min_Uint_AtLimit(t *testing.T) {
	if err := Validate(unsignedUser{Level: 1}); err != nil {
		t.Fatalf("expected nil at uint boundary, got %v", err)
	}
}

func TestValidate_Min_Float_BelowLimit(t *testing.T) {
	type s struct {
		Score float64 `min:"0"`
	}
	err := Validate(s{Score: -0.1})
	assertValidationError(t, err, "Score", "min")
}

func TestValidate_Min_Float_AtLimit(t *testing.T) {
	type s struct {
		Score float64 `min:"0"`
	}
	if err := Validate(s{Score: 0.0}); err != nil {
		t.Fatalf("expected nil at float boundary, got %v", err)
	}
}

// ── max rule ──────────────────────────────────────────────────────────────────

func TestValidate_Max_AboveLimit(t *testing.T) {
	err := Validate(maxOnlyUser{Score: 120})
	assertValidationError(t, err, "Score", "max")
}

func TestValidate_Max_AtLimit(t *testing.T) {
	if err := Validate(maxOnlyUser{Score: 100}); err != nil {
		t.Fatalf("expected nil at boundary, got %v", err)
	}
}

func TestValidate_Max_BelowLimit(t *testing.T) {
	if err := Validate(maxOnlyUser{Score: 50}); err != nil {
		t.Fatalf("expected nil below limit, got %v", err)
	}
}

func TestValidate_Max_Uint_AboveLimit(t *testing.T) {
	err := Validate(unsignedUser{Level: 11})
	assertValidationError(t, err, "Level", "max")
}

func TestValidate_Max_Uint_AtLimit(t *testing.T) {
	if err := Validate(unsignedUser{Level: 10}); err != nil {
		t.Fatalf("expected nil at uint max boundary, got %v", err)
	}
}

func TestValidate_Max_Float_AboveLimit(t *testing.T) {
	type s struct {
		Score float64 `max:"100"`
	}
	err := Validate(s{Score: 100.1})
	assertValidationError(t, err, "Score", "max")
}

func TestValidate_Max_Float_AtLimit(t *testing.T) {
	type s struct {
		Score float64 `max:"100"`
	}
	if err := Validate(s{Score: 100.0}); err != nil {
		t.Fatalf("expected nil at float max boundary, got %v", err)
	}
}

// ── email rule ────────────────────────────────────────────────────────────────

func TestValidate_Email_Valid(t *testing.T) {
	emails := []string{
		"ganesh@gmail.com",
		"user.name+tag@example.co.uk",
		"user@sub.domain.org",
		"123@numbers.io",
	}
	for _, e := range emails {
		t.Run(e, func(t *testing.T) {
			if err := Validate(emailOnlyUser{Email: e}); err != nil {
				t.Errorf("expected nil for %q, got %v", e, err)
			}
		})
	}
}

func TestValidate_Email_Invalid(t *testing.T) {
	invalids := []string{
		"hello",
		"@nodomain",
		"missing-at-sign",
		"user@",
		"@",
		"user@domain",   // no TLD
		"user @gmail.com", // space inside
		"",
	}
	for _, e := range invalids {
		t.Run(e, func(t *testing.T) {
			err := Validate(emailOnlyUser{Email: e})
			assertValidationError(t, err, "Email", "email")
		})
	}
}

// ── ValidationError type ──────────────────────────────────────────────────────

func TestValidationError_ErrorString(t *testing.T) {
	ve := &ValidationError{Field: "Age", Rule: "min", Message: "field Age must be >= 18, got 5"}
	want := `field Age failed rule "min": field Age must be >= 18, got 5`
	if ve.Error() != want {
		t.Errorf("Error() = %q, want %q", ve.Error(), want)
	}
}

func TestValidate_ErrorsAs(t *testing.T) {
	err := Validate(minOnlyUser{Age: 5})
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("errors.As failed: got %T", err)
	}
	if ve.Field != "Age" || ve.Rule != "min" {
		t.Errorf("unexpected ValidationError: %+v", ve)
	}
}

// ── Combined / integration tests ──────────────────────────────────────────────

func TestValidate_FullUser_Valid(t *testing.T) {
	u := fullUser{
		Name:  "Ganesh",
		Age:   22,
		Score: 95.5,
		Email: "ganesh@gmail.com",
	}
	if err := Validate(u); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidate_FullUser_RequiredFails(t *testing.T) {
	u := fullUser{Age: 22, Score: 95, Email: "ganesh@gmail.com"} // Name empty
	assertValidationError(t, Validate(u), "Name", "required")
}

func TestValidate_FullUser_MinFails(t *testing.T) {
	u := fullUser{Name: "Ganesh", Age: 16, Score: 95, Email: "ganesh@gmail.com"}
	assertValidationError(t, Validate(u), "Age", "min")
}

func TestValidate_FullUser_MaxFails(t *testing.T) {
	u := fullUser{Name: "Ganesh", Age: 22, Score: 200, Email: "ganesh@gmail.com"}
	assertValidationError(t, Validate(u), "Score", "max")
}

func TestValidate_FullUser_EmailFails(t *testing.T) {
	u := fullUser{Name: "Ganesh", Age: 22, Score: 95, Email: "not-an-email"}
	assertValidationError(t, Validate(u), "Email", "email")
}

// ── Edge-case / guard tests ───────────────────────────────────────────────────

func TestValidate_NoTags_ReturnsNil(t *testing.T) {
	if err := Validate(noTagUser{Name: "", Age: 0}); err != nil {
		t.Fatalf("expected nil for struct with no tags, got %v", err)
	}
}

func TestValidate_UnexportedFields_Skipped(t *testing.T) {
	// The unexported field carries required:"true" but must be skipped.
	u := unexportedUser{Name: "Ganesh"}
	if err := Validate(u); err != nil {
		t.Fatalf("expected nil (unexported fields must be skipped), got %v", err)
	}
}

func TestValidate_PointerToFullUser_Valid(t *testing.T) {
	u := &fullUser{Name: "Ganesh", Age: 22, Score: 95, Email: "ganesh@gmail.com"}
	if err := Validate(u); err != nil {
		t.Fatalf("expected nil for pointer to valid struct, got %v", err)
	}
}

func TestValidate_MinMax_BothPresent_BothViolated_ReturnsMin(t *testing.T) {
	// min is checked before max, so min error must be returned first.
	type s struct {
		Val int `min:"10" max:"5"` // degenerate range; Val=3 violates min
	}
	assertValidationError(t, Validate(s{Val: 3}), "Val", "min")
}

func TestValidate_Email_OnEmptyString_Fails(t *testing.T) {
	// An empty string is not a valid e-mail even if required is not set.
	err := Validate(emailOnlyUser{Email: ""})
	assertValidationError(t, err, "Email", "email")
}
