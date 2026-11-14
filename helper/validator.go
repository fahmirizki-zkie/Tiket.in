package helper

import (
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateStruct(s any) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	fields := make(map[string]string)
	for _, e := range err.(validator.ValidationErrors) {
		fields[e.Field()] = msgForTag(e)
	}
	return fields
}

func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "min":
		return "minimal " + fe.Param() + " karakter"
	case "max":
		return "maksimal " + fe.Param() + " karakter"
	case "oneof":
		return "harus salah satu dari: " + fe.Param()
	case "gt":
		return "harus lebih besar dari " + fe.Param()
	case "gte":
		return "minimal bernilai " + fe.Param()
	default:
		return "tidak valid"
	}
}
