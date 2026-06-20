# `validator` — Tag-Based Struct Validation

The `validator` package lets you declare validation rules directly on your Go
struct fields using struct tags. No external DSL, no code generation, no
additional registration step — just annotate your structs and call
`validator.Validate` or `validator.ValidateAll`.

---

## Table of Contents

1. [Installation](#installation)
2. [Core Concepts](#core-concepts)
3. [Built-in Validation Tags](#built-in-validation-tags)
   - [required](#required)
   - [min / max](#min--max)
   - [email](#email)
   - [minLength / maxLength](#minlength--maxlength)
   - [regex](#regex)
   - [oneOf](#oneof)
   - [eqField](#eqfield)
4. [Functions](#functions)
   - [Validate](#validate)
   - [ValidateAll](#validateall)
   - [Register (Custom Validators)](#register-custom-validators)
   - [RegisterStructValidator](#registerstructvalidator)
5. [Error Types](#error-types)
   - [ValidationError](#validationerror)
   - [Result](#result)
6. [Complete Example Structs](#complete-example-structs)
7. [Example Outputs](#example-outputs)
8. [Choosing Validate vs ValidateAll](#choosing-validate-vs-validateall)

---

## Installation

```bash
go get github.com/sgdevelopers29-afk/GoKit-Lite
```

```go
import "github.com/sgdevelopers29-afk/GoKit-Lite/validator"
```

---

## Core Concepts

- **Tag-based** — rules live on the struct field itself, right next to the type.
- **Fail-fast (`Validate`)** — stops at the first broken rule; fast for login,
  internal checks, etc.
- **Aggregate (`ValidateAll`)** — collects every broken rule on every field;
  ideal for registration forms where the user should see all errors at once.
- **Recursive** — nested structs, slices of structs, and maps of structs are
  validated automatically.
- **Extensible** — add custom field-level or struct-level rules via `Register`
  and `RegisterStructValidator`.

---

## Built-in Validation Tags

### `required`

**Applies to:** any field type  
**Tag syntax:** `required:"true"`

The field must not hold its zero value:
- strings: must not be `""`
- numbers: must not be `0`
- booleans: must not be `false`
- slices/maps: must be non-nil and non-empty
- pointers: must not be `nil`

```go
type User struct {
    Name string `required:"true"`
}

validator.Validate(User{Name: ""})
// Error: field Name failed rule "required": field Name is required but has a zero value
```

---

### `min` / `max`

**Applies to:** numeric fields (`int`, `uint`, `float32`, `float64` and their variants)  
**Tag syntax:** `min:"<n>"`, `max:"<n>"` (integer values)

```go
type Product struct {
    Quantity int     `min:"1" max:"100"`
    Price    float64 `min:"0"`
}

validator.Validate(Product{Quantity: 0, Price: -5.0})
// Error (fail-fast): field Quantity failed rule "min": field Quantity must be >= 1, got 0
```

| Tag | Meaning |
|---|---|
| `min:"18"` | value must be ≥ 18 |
| `max:"120"` | value must be ≤ 120 |

---

### `email`

**Applies to:** `string` fields  
**Tag syntax:** `email:"true"`

The string must match standard e-mail format (local-part `@` domain `.` TLD).

```go
type Contact struct {
    Email string `required:"true" email:"true"`
}

validator.Validate(Contact{Email: "not-an-email"})
// Error: field Email failed rule "email": field Email must be a valid email address
```

---

### `minLength` / `maxLength`

**Applies to:** `string` fields  
**Tag syntax:** `minLength:"<n>"`, `maxLength:"<n>"`

Length is measured in **Unicode code points** (runes), not bytes, so multi-byte
characters like `é` or `中` count as one character each.

```go
type Post struct {
    Title   string `required:"true" minLength:"3" maxLength:"100"`
    Content string `minLength:"10"`
}

validator.Validate(Post{Title: "Hi"})
// Error: field Title failed rule "minLength": field Title must have at least 3 character(s), got 2
```

---

### `regex`

**Applies to:** `string` fields  
**Tag syntax:** `regex:"<pattern>"` (standard Go regular expression)

The field value must match the regular expression.

```go
type Account struct {
    // Username: only lowercase letters and underscores
    Username string `required:"true" regex:"^[a-z_]+$"`
}

validator.Validate(Account{Username: "Bad User!"})
// Error: field Username failed rule "regex": field Username must match pattern "^[a-z_]+$"
```

> **Tip:** Patterns are compiled once and cached — no performance penalty for
> repeated validation.

---

### `oneOf`

**Applies to:** `string` fields  
**Tag syntax:** `oneOf:"<a>,<b>,<c>,..."`

The field value must exactly match one of the comma-separated allowed values.

```go
type Order struct {
    Status string `required:"true" oneOf:"pending,processing,shipped,delivered,cancelled"`
}

validator.Validate(Order{Status: "unknown"})
// Error: field Status failed rule "oneOf":
//   field Status must be one of [pending,processing,shipped,delivered,cancelled], got "unknown"
```

---

### `eqField`

**Applies to:** any field type  
**Tag syntax:** `eqField:"<OtherFieldName>"`

The tagged field's value must be deeply equal to the value of the named sibling
field in the same struct. Perfect for password confirmation.

```go
type ChangePassword struct {
    NewPassword     string `required:"true" minLength:"8"`
    ConfirmPassword string `required:"true" eqField:"NewPassword"`
}

validator.Validate(ChangePassword{
    NewPassword:     "my-secure-pw",
    ConfirmPassword: "different",
})
// Error: field ConfirmPassword failed rule "eqField":
//   field ConfirmPassword must be equal to field NewPassword
```

> **Note:** The value in the `eqField` tag is the **Go field name** (e.g.
> `"NewPassword"`), not the JSON key.

---

## Functions

### `Validate`

```go
func Validate(data any) error
```

Validate inspects every exported field of `data` using its struct tags and
**returns the first rule violation found** (fail-fast).

- **Input:** a struct value or a non-nil pointer to a struct.
- **Returns:** `nil` if all rules pass, or a `*ValidationError` on the first failure.

```go
err := validator.Validate(myStruct)
if err != nil {
    // Type-assert to inspect field and rule:
    var ve *validator.ValidationError
    if errors.As(err, &ve) {
        fmt.Printf("Field: %s, Rule: %s\n", ve.Field, ve.Rule)
    }
}
```

**When to use `Validate`:**
- Login endpoints (you just need to know the payload is structurally valid).
- Internal service calls where you trust the caller more.
- Any situation where you want to bail out immediately.

---

### `ValidateAll`

```go
func ValidateAll(data any) *Result
```

ValidateAll inspects every exported field and **collects all rule violations**
across all fields. Within a single field, it still stops at the first failing
rule (per-field fail-fast), but it continues to the next field regardless.

- **Input:** a struct value or a non-nil pointer to a struct.
- **Returns:** a `*Result` with `Valid: true` (and an empty `Errors` slice) if
  everything passes, or `Valid: false` plus a populated `Errors` slice on failure.

```go
result := validator.ValidateAll(myStruct)
if !result.Valid {
    for _, e := range result.Errors {
        fmt.Printf("[%s] %s: %s\n", e.Rule, e.Field, e.Message)
    }
}
```

**When to use `ValidateAll`:**
- Registration / sign-up forms — the user benefits from seeing all problems at once.
- Admin forms with many fields.
- Any UX where you want to highlight every broken field in a single round-trip.

---

### `Register` (Custom Validators)

```go
func Register(name string, fn ValidatorFunc) error
```

Register adds a custom field-level validator under a new tag name. The tag is
activated by setting it to `"true"` on a field.

```go
// Register a "noSpaces" rule
validator.Register("noSpaces", func(value any) bool {
    s, ok := value.(string)
    if !ok {
        return true // only applies to strings
    }
    return !strings.Contains(s, " ")
})

type Username struct {
    Handle string `required:"true" noSpaces:"true"`
}

validator.Validate(Username{Handle: "hello world"})
// Error: field Handle failed rule "noSpaces": field Handle failed custom validation "noSpaces"
```

**Constraints:**
- `name` must not be empty.
- `name` must not collide with a built-in tag (`required`, `min`, `max`, etc.).
- `fn` must not be `nil`.

Use `validator.Unregister(name)` to remove a custom rule.

---

### `RegisterStructValidator`

```go
func RegisterStructValidator(t reflect.Type, fn StructValidatorFunc) error
```

RegisterStructValidator attaches a **cross-field (struct-level)** validation
function to a specific struct type. It runs after all field-level rules for that
struct have been evaluated.

```go
type DateRange struct {
    Start int
    End   int
}

validator.RegisterStructValidator(reflect.TypeOf(DateRange{}), func(s any) bool {
    dr := s.(DateRange)
    return dr.End >= dr.Start
})

validator.Validate(DateRange{Start: 2025, End: 2024})
// Error: field DateRange failed rule "struct": struct validation failed for DateRange
```

Use `validator.UnregisterStructValidator(t)` to remove a struct-level rule.

---

## Error Types

### `ValidationError`

```go
type ValidationError struct {
    Field   string // qualified struct field name, e.g. "Address.ZipCode"
    Rule    string // the violated rule, e.g. "required", "email", "minLength"
    Message string // human-readable description
}

func (e *ValidationError) Error() string
// Returns: `field <Field> failed rule "<Rule>": <Message>`
```

You can extract a `*ValidationError` from any `error` returned by `Validate`:

```go
err := validator.Validate(req)
var ve *validator.ValidationError
if errors.As(err, &ve) {
    fmt.Println(ve.Field)   // "Email"
    fmt.Println(ve.Rule)    // "email"
    fmt.Println(ve.Message) // "field Email must be a valid email address"
}
```

---

### `Result`

```go
type Result struct {
    Valid  bool
    Errors []ValidationError
}

func (r *Result) Error() string   // combined error string
func (r *Result) First() *ValidationError
```

`Result` is returned by `ValidateAll`. Check `result.Valid` first; if `false`,
iterate `result.Errors` for the full list of failures.

`result.First()` is a convenience method that returns the first error as a
`*ValidationError`, or `nil` if the result is valid.

---

## Complete Example Structs

### User Registration

```go
type RegisterRequest struct {
    Name            string `json:"name"             required:"true" minLength:"2" maxLength:"50"`
    Email           string `json:"email"            required:"true" email:"true"`
    Password        string `json:"password"         required:"true" minLength:"8"`
    ConfirmPassword string `json:"confirm_password" required:"true" eqField:"Password"`
    Age             int    `json:"age"              min:"13" max:"120"`
    Role            string `json:"role"             oneOf:"user,moderator,admin"`
}
```

### Product Listing

```go
type Product struct {
    SKU      string  `json:"sku"      required:"true" regex:"^[A-Z]{2}[0-9]{4}$"`
    Name     string  `json:"name"     required:"true" minLength:"3" maxLength:"120"`
    Price    float64 `json:"price"    min:"0"`
    Quantity int     `json:"quantity" min:"0" max:"9999"`
    Category string  `json:"category" required:"true" oneOf:"electronics,clothing,food,books"`
}
```

### Login

```go
type LoginRequest struct {
    Email    string `json:"email"    required:"true" email:"true"`
    Password string `json:"password" required:"true"`
}
```

---

## Example Outputs

### `Validate` — fail-fast

```go
type User struct {
    Name  string `required:"true"`
    Email string `required:"true" email:"true"`
    Age   int    `min:"18"`
}

err := validator.Validate(User{Name: "", Email: "bad", Age: 15})
// err.Error() → `field Name failed rule "required": field Name is required but has a zero value`
```

### `ValidateAll` — all errors

```go
result := validator.ValidateAll(User{Name: "", Email: "bad", Age: 15})
// result.Valid → false
// result.Errors:
//   [0] Field:"Name",  Rule:"required",  Message:"field Name is required but has a zero value"
//   [1] Field:"Email", Rule:"email",     Message:"field Email must be a valid email address"
//   [2] Field:"Age",   Rule:"min",       Message:"field Age must be >= 18, got 15"
```

---

## Choosing `Validate` vs `ValidateAll`

| | `Validate` | `ValidateAll` |
|---|---|---|
| **Behaviour** | Fail-fast — stops at first error | Collects all errors |
| **Return type** | `error` | `*Result` |
| **Best for** | Login, internal checks | Registration, forms with many fields |
| **Performance** | Slightly faster (stops early) | Always scans all fields |
| **UX** | Shows one error per submit | Shows all errors in one shot |
