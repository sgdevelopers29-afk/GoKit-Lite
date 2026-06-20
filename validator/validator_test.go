package validator

import (
	"errors"
	"reflect"
	"strings"
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

// ═══════════════════════════════════════════════════════════════════════════════
// Validator V3 tests
// ═══════════════════════════════════════════════════════════════════════════════

// ── V3 test fixtures ──────────────────────────────────────────────────────────

type minLengthUser struct {
	Username string `minLength:"3"`
}

type maxLengthUser struct {
	Username string `maxLength:"20"`
}

type regexUser struct {
	Phone string `regex:"^[0-9]{10}$"`
}

type oneOfUser struct {
	Role string `oneOf:"admin,user,guest"`
}

// v3FullUser exercises every V3 tag alongside V2 tags.
type v3FullUser struct {
	Name     string `required:"true" minLength:"2" maxLength:"50"`
	Age      int    `min:"18" max:"120"`
	Email    string `required:"true" email:"true"`
	Phone    string `regex:"^[0-9]{10}$"`
	Role     string `oneOf:"admin,user,guest"`
}

// ── minLength ─────────────────────────────────────────────────────────────────

func TestValidate_MinLength_BelowLimit(t *testing.T) {
	err := Validate(minLengthUser{Username: "ab"})
	assertValidationError(t, err, "Username", "minLength")
}

func TestValidate_MinLength_AtLimit(t *testing.T) {
	if err := Validate(minLengthUser{Username: "abc"}); err != nil {
		t.Fatalf("expected nil at boundary (3 chars), got %v", err)
	}
}

func TestValidate_MinLength_AboveLimit(t *testing.T) {
	if err := Validate(minLengthUser{Username: "ganesh"}); err != nil {
		t.Fatalf("expected nil above limit, got %v", err)
	}
}

func TestValidate_MinLength_EmptyString_Fails(t *testing.T) {
	err := Validate(minLengthUser{Username: ""})
	assertValidationError(t, err, "Username", "minLength")
}

func TestValidate_MinLength_Zero_AlwaysPasses(t *testing.T) {
	type s struct {
		Name string `minLength:"0"`
	}
	// Any string (including empty) satisfies minLength:"0".
	if err := Validate(s{Name: ""}); err != nil {
		t.Fatalf("expected nil for minLength:0 with empty string, got %v", err)
	}
}

func TestValidate_MinLength_UnicodeMultibyte(t *testing.T) {
	// "日本語" is 3 runes but more than 3 bytes.
	// minLength:"3" should pass because rune count == 3.
	type s struct {
		Text string `minLength:"3"`
	}
	if err := Validate(s{Text: "日本語"}); err != nil {
		t.Fatalf("expected nil for 3-rune multibyte string with minLength:3, got %v", err)
	}
}

func TestValidate_MinLength_UnicodeBelow(t *testing.T) {
	// "日本" is only 2 runes — should fail minLength:"3".
	type s struct {
		Text string `minLength:"3"`
	}
	err := Validate(s{Text: "日本"})
	assertValidationError(t, err, "Text", "minLength")
}

// ── maxLength ─────────────────────────────────────────────────────────────────

func TestValidate_MaxLength_AboveLimit(t *testing.T) {
	err := Validate(maxLengthUser{Username: "thisusernameiswaytoolong"})
	assertValidationError(t, err, "Username", "maxLength")
}

func TestValidate_MaxLength_AtLimit(t *testing.T) {
	// Exactly 20 chars.
	if err := Validate(maxLengthUser{Username: "twentycharacteruserx"}); err != nil {
		t.Fatalf("expected nil at boundary (20 chars), got %v", err)
	}
}

func TestValidate_MaxLength_BelowLimit(t *testing.T) {
	if err := Validate(maxLengthUser{Username: "ganesh"}); err != nil {
		t.Fatalf("expected nil below limit, got %v", err)
	}
}

func TestValidate_MaxLength_EmptyString_Passes(t *testing.T) {
	if err := Validate(maxLengthUser{Username: ""}); err != nil {
		t.Fatalf("expected nil for empty string with maxLength:20, got %v", err)
	}
}

func TestValidate_MaxLength_ZeroLimit_EmptyStringPasses(t *testing.T) {
	type s struct {
		Name string `maxLength:"0"`
	}
	if err := Validate(s{Name: ""}); err != nil {
		t.Fatalf("expected nil for empty string with maxLength:0, got %v", err)
	}
}

func TestValidate_MaxLength_ZeroLimit_NonEmptyFails(t *testing.T) {
	type s struct {
		Name string `maxLength:"0"`
	}
	err := Validate(s{Name: "a"})
	assertValidationError(t, err, "Name", "maxLength")
}

func TestValidate_MaxLength_UnicodeMultibyte(t *testing.T) {
	// "日本語テスト" is 6 runes; maxLength:"5" should fail.
	type s struct {
		Text string `maxLength:"5"`
	}
	err := Validate(s{Text: "日本語テスト"})
	assertValidationError(t, err, "Text", "maxLength")
}

func TestValidate_MaxLength_UnicodeAtLimit(t *testing.T) {
	// "日本語テス" is exactly 5 runes; maxLength:"5" should pass.
	type s struct {
		Text string `maxLength:"5"`
	}
	if err := Validate(s{Text: "日本語テス"}); err != nil {
		t.Fatalf("expected nil for 5-rune string with maxLength:5, got %v", err)
	}
}

// ── regex ─────────────────────────────────────────────────────────────────────

func TestValidate_Regex_Valid(t *testing.T) {
	if err := Validate(regexUser{Phone: "9876543210"}); err != nil {
		t.Fatalf("expected nil for valid 10-digit phone, got %v", err)
	}
}

func TestValidate_Regex_Invalid_Alpha(t *testing.T) {
	err := Validate(regexUser{Phone: "abc123"})
	assertValidationError(t, err, "Phone", "regex")
}

func TestValidate_Regex_Invalid_TooShort(t *testing.T) {
	err := Validate(regexUser{Phone: "12345"})
	assertValidationError(t, err, "Phone", "regex")
}

func TestValidate_Regex_Invalid_TooLong(t *testing.T) {
	err := Validate(regexUser{Phone: "98765432101"})
	assertValidationError(t, err, "Phone", "regex")
}

func TestValidate_Regex_EmptyString_Fails(t *testing.T) {
	err := Validate(regexUser{Phone: ""})
	assertValidationError(t, err, "Phone", "regex")
}

func TestValidate_Regex_CacheHit(t *testing.T) {
	// Call Validate twice with the same struct to exercise the regexpCache
	// Store/Load path without panicking.
	u := regexUser{Phone: "1234567890"}
	for i := 0; i < 3; i++ {
		if err := Validate(u); err != nil {
			t.Fatalf("iteration %d: expected nil, got %v", i, err)
		}
	}
}

func TestValidate_Regex_InvalidPattern_ReturnsError(t *testing.T) {
	type s struct {
		Field string `regex:"[invalid"`
	}
	err := Validate(s{Field: "anything"})
	if err == nil {
		t.Fatal("expected error for malformed regex pattern")
	}
	// Must NOT be a *ValidationError — it is a programming error, not a field error.
	var ve *ValidationError
	if errors.As(err, &ve) {
		t.Fatalf("expected plain error for malformed pattern, got *ValidationError: %v", ve)
	}
}

func TestValidate_Regex_AnchoredPattern(t *testing.T) {
	// Ensures anchored patterns work: "^foo$" must not match "foobar".
	type s struct {
		Code string `regex:"^foo$"`
	}
	if err := Validate(s{Code: "foo"}); err != nil {
		t.Fatalf("expected nil for exact match, got %v", err)
	}
	err := Validate(s{Code: "foobar"})
	assertValidationError(t, err, "Code", "regex")
}

// ── oneOf ─────────────────────────────────────────────────────────────────────

func TestValidate_OneOf_Valid(t *testing.T) {
	roles := []string{"admin", "user", "guest"}
	for _, r := range roles {
		t.Run(r, func(t *testing.T) {
			if err := Validate(oneOfUser{Role: r}); err != nil {
				t.Errorf("expected nil for role %q, got %v", r, err)
			}
		})
	}
}

func TestValidate_OneOf_Invalid(t *testing.T) {
	err := Validate(oneOfUser{Role: "superadmin"})
	assertValidationError(t, err, "Role", "oneOf")
}

func TestValidate_OneOf_EmptyValue_Fails(t *testing.T) {
	err := Validate(oneOfUser{Role: ""})
	assertValidationError(t, err, "Role", "oneOf")
}

func TestValidate_OneOf_CaseSensitive(t *testing.T) {
	// "Admin" (capital A) is not in "admin,user,guest".
	err := Validate(oneOfUser{Role: "Admin"})
	assertValidationError(t, err, "Role", "oneOf")
}

func TestValidate_OneOf_WhitespaceTrimming(t *testing.T) {
	// Tag with spaces around values: "admin, user, guest"
	type s struct {
		Role string `oneOf:"admin, user, guest"`
	}
	// "user" (no extra spaces) must still match " user" entry after trimming.
	if err := Validate(s{Role: "user"}); err != nil {
		t.Fatalf("expected nil after whitespace trimming, got %v", err)
	}
}

func TestValidate_OneOf_SingleOption(t *testing.T) {
	type s struct {
		Status string `oneOf:"active"`
	}
	if err := Validate(s{Status: "active"}); err != nil {
		t.Fatalf("expected nil for single-option oneOf, got %v", err)
	}
	err := Validate(s{Status: "inactive"})
	assertValidationError(t, err, "Status", "oneOf")
}

// ── V3 combined / integration tests ──────────────────────────────────────────

func TestValidate_V3FullUser_Valid(t *testing.T) {
	u := v3FullUser{
		Name:  "Ganesh",
		Age:   22,
		Email: "ganesh@gmail.com",
		Phone: "9876543210",
		Role:  "admin",
	}
	if err := Validate(u); err != nil {
		t.Fatalf("expected nil for fully valid v3FullUser, got %v", err)
	}
}

func TestValidate_V3FullUser_RequiredFails(t *testing.T) {
	u := v3FullUser{Age: 22, Email: "g@g.com", Phone: "9876543210", Role: "admin"}
	assertValidationError(t, Validate(u), "Name", "required")
}

func TestValidate_V3FullUser_MinLengthFails(t *testing.T) {
	u := v3FullUser{Name: "G", Age: 22, Email: "g@g.com", Phone: "9876543210", Role: "admin"}
	assertValidationError(t, Validate(u), "Name", "minLength")
}

func TestValidate_V3FullUser_MaxLengthFails(t *testing.T) {
	// 51 ASCII characters — exceeds maxLength:"50".
	u := v3FullUser{
		Name:  "GaneshKumarWithAReallyReallyReallyLongNameOverFifty",
		Age:   22,
		Email: "g@g.com",
		Phone: "9876543210",
		Role:  "admin",
	}
	assertValidationError(t, Validate(u), "Name", "maxLength")
}

func TestValidate_V3FullUser_RegexFails(t *testing.T) {
	u := v3FullUser{Name: "Ganesh", Age: 22, Email: "g@g.com", Phone: "BADPHONE", Role: "admin"}
	assertValidationError(t, Validate(u), "Phone", "regex")
}

func TestValidate_V3FullUser_OneOfFails(t *testing.T) {
	u := v3FullUser{Name: "Ganesh", Age: 22, Email: "g@g.com", Phone: "9876543210", Role: "superadmin"}
	assertValidationError(t, Validate(u), "Role", "oneOf")
}

func TestValidate_V3FullUser_PointerValid(t *testing.T) {
	u := &v3FullUser{
		Name:  "Ganesh",
		Age:   22,
		Email: "ganesh@gmail.com",
		Phone: "9876543210",
		Role:  "user",
	}
	if err := Validate(u); err != nil {
		t.Fatalf("expected nil for valid *v3FullUser, got %v", err)
	}
}

// ── V3 edge-cases ─────────────────────────────────────────────────────────────

func TestValidate_MinLength_OnNonString_NoOp(t *testing.T) {
	// minLength on an int field must be silently ignored (no-op).
	type s struct {
		Age int `minLength:"3"`
	}
	if err := Validate(s{Age: 1}); err != nil {
		t.Fatalf("minLength on int should be a no-op, got %v", err)
	}
}

func TestValidate_MaxLength_OnNonString_NoOp(t *testing.T) {
	type s struct {
		Age int `maxLength:"3"`
	}
	if err := Validate(s{Age: 9999}); err != nil {
		t.Fatalf("maxLength on int should be a no-op, got %v", err)
	}
}

func TestValidate_Regex_OnNonString_NoOp(t *testing.T) {
	type s struct {
		Age int `regex:"^[0-9]+$"`
	}
	if err := Validate(s{Age: 42}); err != nil {
		t.Fatalf("regex on int should be a no-op, got %v", err)
	}
}

func TestValidate_OneOf_OnNonString_NoOp(t *testing.T) {
	type s struct {
		Level int `oneOf:"1,2,3"`
	}
	if err := Validate(s{Level: 99}); err != nil {
		t.Fatalf("oneOf on int should be a no-op, got %v", err)
	}
}

func TestValidate_MinMaxLength_BothPresent_MinCheckedFirst(t *testing.T) {
	// minLength is applied before maxLength; with an impossible range and a
	// value that violates both, minLength error must be returned.
	type s struct {
		// degenerate: min > max; value "a" (len 1) violates minLength:"5" first.
		Name string `minLength:"5" maxLength:"3"`
	}
	assertValidationError(t, Validate(s{Name: "a"}), "Name", "minLength")
}

func TestValidate_RuleOrderV3(t *testing.T) {
	// A field that violates both required and minLength: required is checked
	// first, so the required error should be returned.
	type s struct {
		Name string `required:"true" minLength:"5"`
	}
	assertValidationError(t, Validate(s{Name: ""}), "Name", "required")
}

// ═══════════════════════════════════════════════════════════════════════════════
// Validator V4 tests — Nested Struct, Slice, Map, Recursive Engine
// ═══════════════════════════════════════════════════════════════════════════════

// ── V4 test fixtures ──────────────────────────────────────────────────────────

// v4Address is a nested struct used by V4 tests.
type v4Address struct {
	City    string `required:"true"`
	Pincode string `regex:"^[0-9]{6}$"`
}

// v4User is a top-level struct that embeds a nested struct, a slice, and a map.
type v4User struct {
	Name     string            `required:"true" minLength:"2"`
	Age      int               `min:"18" max:"120"`
	Address  v4Address
	Skills   []string          `required:"true"`
	Metadata map[string]string `required:"true"`
}

// v4Tag is a struct used as a slice/map element with its own validation tags.
type v4Tag struct {
	Label string `required:"true" minLength:"1"`
	Value string `oneOf:"low,medium,high"`
}

// v4Deep is used for 3-level nesting tests.
type v4Deep struct {
	Level1 struct {
		Level2 struct {
			Name string `required:"true"`
		}
	}
}

// ── Nested struct validation ───────────────────────────────────────────────────

func TestValidate_V4_NestedStruct_Valid(t *testing.T) {
	u := v4User{
		Name:     "Ganesh",
		Age:      22,
		Address:  v4Address{City: "Mumbai", Pincode: "400001"},
		Skills:   []string{"Go"},
		Metadata: map[string]string{"role": "dev"},
	}
	if err := Validate(u); err != nil {
		t.Fatalf("expected nil for fully valid v4User, got %v", err)
	}
}

func TestValidate_V4_NestedStruct_RequiredFieldMissing(t *testing.T) {
	// Address.City is required but empty — error must carry qualified path.
	u := v4User{
		Name:     "Ganesh",
		Age:      22,
		Address:  v4Address{City: "", Pincode: "400001"},
		Skills:   []string{"Go"},
		Metadata: map[string]string{"role": "dev"},
	}
	err := Validate(u)
	assertValidationError(t, err, "Address.City", "required")
}

func TestValidate_V4_NestedStruct_RegexFailure(t *testing.T) {
	// Address.Pincode fails regex — qualified path must be "Address.Pincode".
	u := v4User{
		Name:     "Ganesh",
		Age:      22,
		Address:  v4Address{City: "Mumbai", Pincode: "BAD"},
		Skills:   []string{"Go"},
		Metadata: map[string]string{"role": "dev"},
	}
	err := Validate(u)
	assertValidationError(t, err, "Address.Pincode", "regex")
}

func TestValidate_V4_NestedStruct_TopLevelCheckedBeforeNested(t *testing.T) {
	// Root-level Name is empty — must be caught before descending into Address.
	u := v4User{
		Name:     "",
		Age:      22,
		Address:  v4Address{City: "Mumbai", Pincode: "400001"},
		Skills:   []string{"Go"},
		Metadata: map[string]string{"role": "dev"},
	}
	err := Validate(u)
	assertValidationError(t, err, "Name", "required")
}

func TestValidate_V4_NestedStruct_ThreeLevelsDeep(t *testing.T) {
	// Level1.Level2.Name is required but empty.
	d := v4Deep{}
	err := Validate(d)
	assertValidationError(t, err, "Level1.Level2.Name", "required")
}

func TestValidate_V4_NestedStruct_ThreeLevelsDeep_Valid(t *testing.T) {
	d := v4Deep{}
	d.Level1.Level2.Name = "deep"
	if err := Validate(d); err != nil {
		t.Fatalf("expected nil for valid 3-level nested struct, got %v", err)
	}
}

// Pointer-to-struct nested field — valid (non-nil, valid inner field).
func TestValidate_V4_PointerToNestedStruct_Valid(t *testing.T) {
	type profile struct {
		Bio string `required:"true"`
	}
	type user struct {
		Name    string   `required:"true"`
		Profile *profile
	}
	u := user{Name: "Ganesh", Profile: &profile{Bio: "developer"}}
	if err := Validate(u); err != nil {
		t.Fatalf("expected nil for valid pointer-to-nested-struct, got %v", err)
	}
}

// Pointer-to-struct nested field — nil pointer must be skipped (no panic),
// unless the field itself is marked required.
func TestValidate_V4_PointerToNestedStruct_NilSkipped(t *testing.T) {
	type profile struct {
		Bio string `required:"true"`
	}
	type user struct {
		Name    string   `required:"true"`
		Profile *profile // no required tag — nil is acceptable
	}
	u := user{Name: "Ganesh", Profile: nil}
	if err := Validate(u); err != nil {
		t.Fatalf("nil pointer-to-struct without required tag must be skipped, got %v", err)
	}
}

// Pointer-to-struct nested field — nil pointer WITH required tag must fail.
func TestValidate_V4_PointerToNestedStruct_NilRequired(t *testing.T) {
	type profile struct {
		Bio string `required:"true"`
	}
	type user struct {
		Name    string   `required:"true"`
		Profile *profile `required:"true"`
	}
	u := user{Name: "Ganesh", Profile: nil}
	err := Validate(u)
	assertValidationError(t, err, "Profile", "required")
}

// Pointer-to-struct nested field — invalid inner field must be caught.
func TestValidate_V4_PointerToNestedStruct_InvalidInner(t *testing.T) {
	type profile struct {
		Bio string `required:"true"`
	}
	type user struct {
		Name    string   `required:"true"`
		Profile *profile
	}
	u := user{Name: "Ganesh", Profile: &profile{Bio: ""}}
	err := Validate(u)
	assertValidationError(t, err, "Profile.Bio", "required")
}

// ── Slice validation ───────────────────────────────────────────────────────────

func TestValidate_V4_Slice_Required_NilFails(t *testing.T) {
	// A nil slice with required:"true" must fail.
	type s struct {
		Tags []string `required:"true"`
	}
	err := Validate(s{Tags: nil})
	assertValidationError(t, err, "Tags", "required")
}

func TestValidate_V4_Slice_Required_EmptyFails(t *testing.T) {
	// An initialised but empty slice with required:"true" must also fail.
	type s struct {
		Tags []string `required:"true"`
	}
	err := Validate(s{Tags: []string{}})
	assertValidationError(t, err, "Tags", "required")
}

func TestValidate_V4_Slice_Required_NonEmpty_Passes(t *testing.T) {
	type s struct {
		Tags []string `required:"true"`
	}
	if err := Validate(s{Tags: []string{"go", "backend"}}); err != nil {
		t.Fatalf("expected nil for non-empty required slice, got %v", err)
	}
}

func TestValidate_V4_Slice_NoRequiredTag_NilPasses(t *testing.T) {
	// Without required, a nil slice must pass without error.
	type s struct {
		Tags []string
	}
	if err := Validate(s{Tags: nil}); err != nil {
		t.Fatalf("expected nil for nil slice without required tag, got %v", err)
	}
}

func TestValidate_V4_Slice_OfStructs_Valid(t *testing.T) {
	type s struct {
		Tags []v4Tag
	}
	st := s{Tags: []v4Tag{
		{Label: "speed", Value: "high"},
		{Label: "cost", Value: "low"},
	}}
	if err := Validate(st); err != nil {
		t.Fatalf("expected nil for valid slice of structs, got %v", err)
	}
}

func TestValidate_V4_Slice_OfStructs_InvalidElement(t *testing.T) {
	// Tags[1].Label is empty — required:"true" must fail with qualified path.
	type s struct {
		Tags []v4Tag
	}
	st := s{Tags: []v4Tag{
		{Label: "speed", Value: "high"},
		{Label: "", Value: "low"}, // invalid
	}}
	err := Validate(st)
	assertValidationError(t, err, "Tags[1].Label", "required")
}

func TestValidate_V4_Slice_OfStructs_InvalidElement_OneOf(t *testing.T) {
	// Tags[0].Value is not in the oneOf list.
	type s struct {
		Tags []v4Tag
	}
	st := s{Tags: []v4Tag{
		{Label: "speed", Value: "ultra"}, // invalid Value
	}}
	err := Validate(st)
	assertValidationError(t, err, "Tags[0].Value", "oneOf")
}

func TestValidate_V4_Slice_OfPointerToStruct_NilElementSkipped(t *testing.T) {
	// A nil pointer element inside a slice must not panic; it must be skipped.
	type inner struct {
		Name string `required:"true"`
	}
	type s struct {
		Items []*inner
	}
	st := s{Items: []*inner{nil, {Name: "ok"}}}
	if err := Validate(st); err != nil {
		t.Fatalf("nil pointer element in slice must be skipped, got %v", err)
	}
}

func TestValidate_V4_Slice_OfPointerToStruct_InvalidInner(t *testing.T) {
	type inner struct {
		Name string `required:"true"`
	}
	type s struct {
		Items []*inner
	}
	st := s{Items: []*inner{{Name: "ok"}, {Name: ""}}}
	err := Validate(st)
	assertValidationError(t, err, "Items[1].Name", "required")
}

// ── Map validation ─────────────────────────────────────────────────────────────

func TestValidate_V4_Map_Required_NilFails(t *testing.T) {
	type s struct {
		Meta map[string]string `required:"true"`
	}
	err := Validate(s{Meta: nil})
	assertValidationError(t, err, "Meta", "required")
}

func TestValidate_V4_Map_Required_EmptyFails(t *testing.T) {
	type s struct {
		Meta map[string]string `required:"true"`
	}
	err := Validate(s{Meta: map[string]string{}})
	assertValidationError(t, err, "Meta", "required")
}

func TestValidate_V4_Map_Required_NonEmpty_Passes(t *testing.T) {
	type s struct {
		Meta map[string]string `required:"true"`
	}
	if err := Validate(s{Meta: map[string]string{"env": "prod"}}); err != nil {
		t.Fatalf("expected nil for non-empty required map, got %v", err)
	}
}

func TestValidate_V4_Map_NoRequiredTag_NilPasses(t *testing.T) {
	type s struct {
		Meta map[string]string
	}
	if err := Validate(s{Meta: nil}); err != nil {
		t.Fatalf("expected nil for nil map without required tag, got %v", err)
	}
}

func TestValidate_V4_Map_OfStructValues_Valid(t *testing.T) {
	type inner struct {
		Name string `required:"true"`
	}
	type s struct {
		Registry map[string]inner
	}
	st := s{Registry: map[string]inner{
		"alpha": {Name: "Alpha Service"},
		"beta":  {Name: "Beta Service"},
	}}
	if err := Validate(st); err != nil {
		t.Fatalf("expected nil for valid map of structs, got %v", err)
	}
}

func TestValidate_V4_Map_OfStructValues_InvalidValue(t *testing.T) {
	// One map value has an empty required field — must return a qualified error.
	type inner struct {
		Name string `required:"true"`
	}
	type s struct {
		Registry map[string]inner
	}
	st := s{Registry: map[string]inner{
		"alpha": {Name: ""},
	}}
	err := Validate(st)
	if err == nil {
		t.Fatal("expected ValidationError for invalid map struct value, got nil")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
	// Field must be "Registry[alpha].Name"
	if ve.Field != "Registry[alpha].Name" {
		t.Errorf("expected Field %q, got %q", "Registry[alpha].Name", ve.Field)
	}
	if ve.Rule != "required" {
		t.Errorf("expected Rule %q, got %q", "required", ve.Rule)
	}
}

func TestValidate_V4_Map_OfPointerToStructValues_NilValueSkipped(t *testing.T) {
	// nil pointer values inside a map must not panic; they must be skipped.
	type inner struct {
		Name string `required:"true"`
	}
	type s struct {
		Registry map[string]*inner
	}
	st := s{Registry: map[string]*inner{
		"alpha": nil,
		"beta":  {Name: "Beta"},
	}}
	if err := Validate(st); err != nil {
		t.Fatalf("nil pointer map value must be skipped, got %v", err)
	}
}

// ── Integration / mixed-nesting tests ─────────────────────────────────────────

func TestValidate_V4_Mixed_AllValid(t *testing.T) {
	// v4User exercises nested struct, required slice, and required map together.
	u := v4User{
		Name:     "Ganesh",
		Age:      25,
		Address:  v4Address{City: "Pune", Pincode: "411001"},
		Skills:   []string{"Go", "Docker"},
		Metadata: map[string]string{"team": "backend"},
	}
	if err := Validate(u); err != nil {
		t.Fatalf("expected nil for fully valid mixed v4User, got %v", err)
	}
}

func TestValidate_V4_Mixed_NestedStructFailsFirst(t *testing.T) {
	// Address is validated after top-level fields; Skills is validated after
	// Address. With Address.City empty, we expect the nested error first.
	u := v4User{
		Name:     "Ganesh",
		Age:      25,
		Address:  v4Address{City: "", Pincode: "411001"}, // invalid
		Skills:   []string{},                              // also invalid, checked after
		Metadata: map[string]string{"team": "backend"},
	}
	err := Validate(u)
	assertValidationError(t, err, "Address.City", "required")
}

func TestValidate_V4_Mixed_SliceFailsAfterNested(t *testing.T) {
	// Address is valid but Skills is empty (required).
	u := v4User{
		Name:     "Ganesh",
		Age:      25,
		Address:  v4Address{City: "Pune", Pincode: "411001"},
		Skills:   nil,                                     // invalid — required
		Metadata: map[string]string{"team": "backend"},
	}
	err := Validate(u)
	assertValidationError(t, err, "Skills", "required")
}

func TestValidate_V4_Mixed_MapFailsAfterSlice(t *testing.T) {
	// Address valid, Skills valid, Metadata empty — Metadata error expected.
	u := v4User{
		Name:     "Ganesh",
		Age:      25,
		Address:  v4Address{City: "Pune", Pincode: "411001"},
		Skills:   []string{"Go"},
		Metadata: map[string]string{}, // invalid — required
	}
	err := Validate(u)
	assertValidationError(t, err, "Metadata", "required")
}

func TestValidate_V4_Pointer_ToRootStruct_Valid(t *testing.T) {
	// Validate still accepts *T for the root struct in V4.
	u := &v4User{
		Name:     "Ganesh",
		Age:      25,
		Address:  v4Address{City: "Pune", Pincode: "411001"},
		Skills:   []string{"Go"},
		Metadata: map[string]string{"env": "prod"},
	}
	if err := Validate(u); err != nil {
		t.Fatalf("expected nil for valid *v4User, got %v", err)
	}
}

// ── V4 edge-case tests ─────────────────────────────────────────────────────────

func TestValidate_V4_EmptyNestedStruct_NoTags_Passes(t *testing.T) {
	// A nested struct with no validation tags must never produce an error.
	type inner struct {
		Comment string
	}
	type outer struct {
		Name  string `required:"true"`
		Extra inner
	}
	o := outer{Name: "Ganesh", Extra: inner{Comment: ""}}
	if err := Validate(o); err != nil {
		t.Fatalf("expected nil for nested struct with no tags, got %v", err)
	}
}

func TestValidate_V4_QualifiedName_MultiLevelSliceStruct(t *testing.T) {
	// Slice of structs, each of which has its own nested struct.
	type inner struct {
		Tag   v4Tag
	}
	type outer struct {
		Items []inner
	}
	o := outer{Items: []inner{
		{Tag: v4Tag{Label: "ok", Value: "high"}},
		{Tag: v4Tag{Label: "", Value: "low"}}, // Items[1].Tag.Label fails required
	}}
	err := Validate(o)
	assertValidationError(t, err, "Items[1].Tag.Label", "required")
}

func TestValidate_V4_Slice_Array_OfStructs_Valid(t *testing.T) {
	// [2]v4Tag (fixed-size array) must also be recursed into.
	type s struct {
		Tags [2]v4Tag
	}
	st := s{Tags: [2]v4Tag{
		{Label: "a", Value: "low"},
		{Label: "b", Value: "high"},
	}}
	if err := Validate(st); err != nil {
		t.Fatalf("expected nil for valid fixed-size array of structs, got %v", err)
	}
}

func TestValidate_V4_Slice_Array_OfStructs_InvalidElement(t *testing.T) {
	type s struct {
		Tags [2]v4Tag
	}
	st := s{Tags: [2]v4Tag{
		{Label: "a", Value: "low"},
		{Label: "", Value: "high"}, // Tags[1].Label required fails
	}}
	err := Validate(st)
	assertValidationError(t, err, "Tags[1].Label", "required")
}

func TestValidate_V4_qualifiedName_Helper(t *testing.T) {
	// Internal helper — verify dot-path construction via the observable Field.
	type child struct {
		Val string `required:"true"`
	}
	type parent struct {
		Child child
	}
	err := Validate(parent{Child: child{Val: ""}})
	assertValidationError(t, err, "Child.Val", "required")
}

func TestValidate_V4_StructField_WithAllV3Tags_InsideNested(t *testing.T) {
	// Nested struct uses all V3 string tags — ensure rules still fire correctly
	// even when the struct is accessed through recursive descent.
	type inner struct {
		Role  string `oneOf:"admin,user"`
		Phone string `regex:"^[0-9]{10}$"`
	}
	type outer struct {
		Name  string `required:"true"`
		Inner inner
	}
	// Inner.Role violates oneOf.
	err := Validate(outer{Name: "Ganesh", Inner: inner{Role: "superadmin", Phone: "9876543210"}})
	assertValidationError(t, err, "Inner.Role", "oneOf")
}

func TestValidate_V4_DepthGuard_Triggers(t *testing.T) {
	// validateStruct must not panic on deeply nested structures; the depth guard
	// returns a descriptive error instead.
	// We build a chain 33 levels deep (one beyond maxRecursionDepth=32) using
	// an interface field to force reflect descent at each level.
	//
	// Since Go does not support recursive struct types directly, we use the
	// public Validate entry-point with a manually-constructed reflect.Value via
	// a helper type chain generated to exceed the depth limit.
	//
	// The simplest safe test: call validateStruct directly at depth > limit.
	type leaf struct{ X string }
	v := reflect.ValueOf(leaf{X: "ok"})
	_, err := collectErrors(v, "test", maxRecursionDepth+1, true)
	if err == nil {
		t.Fatal("expected error when depth > maxRecursionDepth, got nil")
	}
	// Must be a plain error, not a *ValidationError.
	var ve *ValidationError
	if errors.As(err, &ve) {
		t.Fatalf("depth guard must return plain error, got *ValidationError: %v", ve)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// Validator V5 tests — Custom Validators, Struct Validators, eqField, Aggregation
// ═══════════════════════════════════════════════════════════════════════════════

// ── Custom Field Validators ───────────────────────────────────────────────────

func TestRegister_CustomValidator_Valid(t *testing.T) {
	err := Register("is_even", func(v any) bool {
		i, ok := v.(int)
		return ok && i%2 == 0
	})
	if err != nil {
		t.Fatalf("expected nil error on valid Register, got %v", err)
	}
	defer Unregister("is_even")

	type s struct {
		Val int `is_even:"true"`
	}

	if err := Validate(s{Val: 4}); err != nil {
		t.Errorf("expected nil for valid custom validator, got %v", err)
	}

	err = Validate(s{Val: 5})
	assertValidationError(t, err, "Val", "is_even")
}

func TestRegister_EmptyName_Error(t *testing.T) {
	err := Register("", func(any) bool { return true })
	if err == nil {
		t.Fatal("expected error when registering with empty name")
	}
}

func TestRegister_NilFunc_Error(t *testing.T) {
	err := Register("nil_func", nil)
	if err == nil {
		t.Fatal("expected error when registering with nil function")
	}
}

func TestRegister_ReservedTag_Error(t *testing.T) {
	err := Register("required", func(any) bool { return true })
	if err == nil {
		t.Fatal("expected error when overwriting reserved tag")
	}
}

func TestApplyCustomValidator_WrongValue_Ignored(t *testing.T) {
	// Only tag value "true" activates the custom validator
	Register("dummy", func(any) bool { return false })
	defer Unregister("dummy")

	type s struct {
		Val int `dummy:"false"`
	}

	if err := Validate(s{Val: 1}); err != nil {
		t.Fatalf("expected nil when custom tag value is not 'true', got %v", err)
	}
}

// ── Struct-Level Validators ───────────────────────────────────────────────────

func TestRegisterStructValidator_Valid(t *testing.T) {
	type rangeStruct struct {
		Min int
		Max int
	}
	rt := reflect.TypeOf(rangeStruct{})

	err := RegisterStructValidator(rt, func(s any) bool {
		r, ok := s.(rangeStruct)
		return ok && r.Min <= r.Max
	})
	if err != nil {
		t.Fatalf("expected nil error on valid RegisterStructValidator, got %v", err)
	}
	defer UnregisterStructValidator(rt)

	if err := Validate(rangeStruct{Min: 1, Max: 5}); err != nil {
		t.Errorf("expected nil for valid struct validator, got %v", err)
	}

	err = Validate(rangeStruct{Min: 5, Max: 1})
	assertValidationError(t, err, "rangeStruct", "struct")
}

func TestRegisterStructValidator_NilType_Error(t *testing.T) {
	err := RegisterStructValidator(nil, func(any) bool { return true })
	if err == nil {
		t.Fatal("expected error when registering nil type")
	}
}

func TestRegisterStructValidator_NonStructType_Error(t *testing.T) {
	rt := reflect.TypeOf(42) // int type
	err := RegisterStructValidator(rt, func(any) bool { return true })
	if err == nil {
		t.Fatal("expected error when registering non-struct type")
	}
}

func TestRegisterStructValidator_NilFunc_Error(t *testing.T) {
	rt := reflect.TypeOf(struct{}{})
	err := RegisterStructValidator(rt, nil)
	if err == nil {
		t.Fatal("expected error when registering nil func")
	}
}

func TestRegisterStructValidator_NestedStruct(t *testing.T) {
	type inner struct {
		A int
		B int
	}
	type outer struct {
		Nested inner
	}

	rt := reflect.TypeOf(inner{})
	RegisterStructValidator(rt, func(s any) bool {
		in, ok := s.(inner)
		return ok && in.A == in.B
	})
	defer UnregisterStructValidator(rt)

	// Valid
	if err := Validate(outer{Nested: inner{A: 1, B: 1}}); err != nil {
		t.Errorf("expected nil for valid nested struct validator, got %v", err)
	}

	// Invalid
	err := Validate(outer{Nested: inner{A: 1, B: 2}})
	assertValidationError(t, err, "Nested", "struct")
}

// ── eqField tag ───────────────────────────────────────────────────────────────

func TestValidate_EqField_Valid(t *testing.T) {
	type s struct {
		P1 string
		P2 string `eqField:"P1"`
	}
	if err := Validate(s{P1: "pass", P2: "pass"}); err != nil {
		t.Fatalf("expected nil for equal fields, got %v", err)
	}
}

func TestValidate_EqField_Invalid(t *testing.T) {
	type s struct {
		P1 string
		P2 string `eqField:"P1"`
	}
	err := Validate(s{P1: "pass", P2: "fail"})
	assertValidationError(t, err, "P2", "eqField")
}

func TestValidate_EqField_NonStringTypes(t *testing.T) {
	type s struct {
		A int
		B int `eqField:"A"`
	}
	if err := Validate(s{A: 42, B: 42}); err != nil {
		t.Fatalf("expected nil for equal int fields, got %v", err)
	}
	err := Validate(s{A: 42, B: 100})
	assertValidationError(t, err, "B", "eqField")
}

func TestValidate_EqField_MissingTarget_ReturnsError(t *testing.T) {
	type s struct {
		A string `eqField:"NonExistent"`
	}
	err := Validate(s{A: "test"})
	if err == nil {
		t.Fatal("expected error for missing target field")
	}
	var ve *ValidationError
	if errors.As(err, &ve) {
		t.Fatalf("expected plain error, got *ValidationError: %v", ve)
	}
}

// ── ValidateAll Error Aggregation ─────────────────────────────────────────────

func TestValidateAll_Valid(t *testing.T) {
	type s struct {
		A string `required:"true"`
	}
	res := ValidateAll(s{A: "ok"})
	if !res.Valid {
		t.Errorf("expected Valid=true, got false")
	}
	if len(res.Errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(res.Errors))
	}
	if res.First() != nil {
		t.Errorf("expected First() to return nil, got %v", res.First())
	}
	if res.Error() != "validator: valid" {
		t.Errorf("expected 'validator: valid', got %q", res.Error())
	}
}

func TestValidateAll_MultipleErrors(t *testing.T) {
	type s struct {
		A string `required:"true"`
		B int    `min:"10"`
		C string `email:"true"`
	}
	res := ValidateAll(s{A: "", B: 5, C: "bad"})
	if res.Valid {
		t.Fatal("expected Valid=false")
	}
	if len(res.Errors) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(res.Errors))
	}

	// Order is guaranteed by struct definition
	if res.Errors[0].Field != "A" || res.Errors[0].Rule != "required" {
		t.Errorf("error 0 mismatch: %+v", res.Errors[0])
	}
	if res.Errors[1].Field != "B" || res.Errors[1].Rule != "min" {
		t.Errorf("error 1 mismatch: %+v", res.Errors[1])
	}
	if res.Errors[2].Field != "C" || res.Errors[2].Rule != "email" {
		t.Errorf("error 2 mismatch: %+v", res.Errors[2])
	}

	// First()
	first := res.First()
	if first == nil || first.Field != "A" {
		t.Errorf("First() mismatch: %+v", first)
	}

	// Error()
	msg := res.Error()
	if !strings.Contains(msg, "3 validation errors") {
		t.Errorf("Error() format unexpected: %q", msg)
	}
}

func TestValidateAll_PerFieldFailFast(t *testing.T) {
	type s struct {
		A string `required:"true" minLength:"5"`
	}
	// A fails both required and minLength, but only required should be collected
	res := ValidateAll(s{A: ""})
	if res.Valid {
		t.Fatal("expected Valid=false")
	}
	if len(res.Errors) != 1 {
		t.Fatalf("expected 1 error per field, got %d", len(res.Errors))
	}
	if res.Errors[0].Rule != "required" {
		t.Errorf("expected required error, got %s", res.Errors[0].Rule)
	}
}

func TestValidateAll_NestedErrors(t *testing.T) {
	type inner struct {
		X string `required:"true"`
		Y int    `min:"10"`
	}
	type outer struct {
		Inner inner
		Z     string `required:"true"`
	}

	res := ValidateAll(outer{
		Inner: inner{X: "", Y: 5},
		Z:     "",
	})

	if res.Valid || len(res.Errors) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(res.Errors))
	}

	if res.Errors[0].Field != "Inner.X" {
		t.Errorf("error 0 mismatch: %s", res.Errors[0].Field)
	}
	if res.Errors[1].Field != "Inner.Y" {
		t.Errorf("error 1 mismatch: %s", res.Errors[1].Field)
	}
	if res.Errors[2].Field != "Z" {
		t.Errorf("error 2 mismatch: %s", res.Errors[2].Field)
	}
}

func TestValidateAll_NilInput(t *testing.T) {
	res := ValidateAll(nil)
	if res.Valid {
		t.Fatal("expected Valid=false for nil input")
	}
	if len(res.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(res.Errors))
	}
	if res.Errors[0].Rule != "input" {
		t.Errorf("expected 'input' rule error, got %s", res.Errors[0].Rule)
	}
}
