package blog

import "github.com/go-playground/validator/v10"

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ValidateUsername(username string) error {
	return validate.Var(username, "required")
}

func ValidatePassword(password string) error {
	return validate.Var(password, "required,min=8")
}

func ValidateEmail(email string) error {
	return validate.Var(email, "required,email")
}
