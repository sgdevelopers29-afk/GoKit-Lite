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
