package validator

import "testing"

type User struct {
	Name  string `required:"true"`
	Email string `required:"true"`
	Age   int
}

func TestValidateSuccess(t *testing.T) {
	user := User{
		Name:  "Ganesh",
		Email: "ganesh@gmail.com",
		Age:   22,
	}

	err := Validate(user)

	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateRequiredField(t *testing.T) {
	user := User{
		Name:  "",
		Email: "ganesh@gmail.com",
	}

	err := Validate(user)

	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateMultipleRequiredFields(t *testing.T) {
	user := User{}

	err := Validate(user)

	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidatePointer(t *testing.T) {
	user := &User{
		Name:  "Ganesh",
		Email: "ganesh@gmail.com",
	}

	err := Validate(user)

	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateNil(t *testing.T) {
	err := Validate(nil)

	if err == nil {
		t.Fatal("expected error for nil input")
	}
}

func TestValidateNonStructString(t *testing.T) {
	err := Validate("hello")

	if err == nil {
		t.Fatal("expected error for string input")
	}
}

func TestValidateNonStructInt(t *testing.T) {
	err := Validate(100)

	if err == nil {
		t.Fatal("expected error for int input")
	}
}

func TestValidateNilPointer(t *testing.T) {
	var user *User

	err := Validate(user)

	if err == nil {
		t.Fatal("expected error for nil pointer")
	}
}
