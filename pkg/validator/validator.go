package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	govalidator "github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *govalidator.Validate
}

func New() *Validator {
	v := govalidator.New()
	// Gunakan nama field dari tag json agar pesan error konsisten dengan API.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return fld.Name
		}
		return name
	})
	return &Validator{validate: v}
}

// Validate mengembalikan map field -> pesan error dalam bahasa Indonesia.
func (v *Validator) Validate(s any) map[string]string {
	if err := v.validate.Struct(s); err != nil {
		out := make(map[string]string)
		var verrs govalidator.ValidationErrors
		if errors.As(err, &verrs) {
			for _, fe := range verrs {
				out[fe.Field()] = message(fe)
			}
			return out
		}
		out["_error"] = err.Error()
		return out
	}
	return nil
}

func message(fe govalidator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s wajib diisi", fe.Field())
	case "email":
		return "Email tidak valid"
	case "min":
		return fmt.Sprintf("%s minimal %s karakter", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s maksimal %s karakter", fe.Field(), fe.Param())
	case "gt":
		return fmt.Sprintf("%s harus lebih besar dari %s", fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf("%s minimal bernilai %s", fe.Field(), fe.Param())
	case "numeric":
		return fmt.Sprintf("%s harus berupa angka", fe.Field())
	case "oneof":
		return fmt.Sprintf("%s harus salah satu dari: %s", fe.Field(), fe.Param())
	case "eqfield":
		return fmt.Sprintf("%s tidak sama dengan %s", fe.Field(), fe.Param())
	default:
		return fmt.Sprintf("%s tidak valid", fe.Field())
	}
}
