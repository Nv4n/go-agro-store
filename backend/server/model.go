package server

import (
	"github.com/go-playground/validator/v10"
	"log/slog"
	"regexp"
)

type UserRegister struct {
	Email     string `json:"email" form:"email" validate:"required,email"`
	Password  string `json:"password" form:"password" validate:"required,min=8,max=32"`
	FirstName string `json:"first_name" form:"fname" validate:"required,name"`
	LastName  string `json:"last_name" form:"lname" validate:"required,name"`
}

type UserLogin struct {
	Email    string `json:"email" form:"email" validate:"required,email"`
	Password string `json:"password" form:"password" validate:"required,min=8,max=32"`
}

var nameRegex = `^[A-ZА-Я][a-zа-я]{1,49}$`

func nameValidator(fl validator.FieldLevel) bool {
	match, err := regexp.MatchString(nameRegex, fl.Field().String())
	if err != nil {
		slog.Warn(err.Error())
	}
	return match
}

func NewValidator() (*validator.Validate, error) {
	validate := validator.New()

	err := validate.RegisterValidation("name", nameValidator)
	if err != nil {
		return nil, err
	}
	return validate, nil
}
