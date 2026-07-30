package helper

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func init() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

func ParseValidationError(err error) map[string]string {
	var ve validator.ValidationErrors
	errs := map[string]string{}

	if !errors.As(err, &ve) {
		errs["body"] = "Isi permintaan tidak valid"
		return errs
	}

	for _, fe := range ve {
		field := strings.ToLower(fe.Field())
		switch fe.Tag() {
		case "required":
			errs[field] = fmt.Sprintf("Kolom %s wajib diisi", field)
		case "min":
			errs[field] = fmt.Sprintf("Kolom %s minimal harus %s karakter", field, fe.Param())
		case "max":
			errs[field] = fmt.Sprintf("Kolom %s maksimal harus %s karakter", field, fe.Param())
		case "email":
			errs[field] = fmt.Sprintf("Kolom %s harus berupa email yang valid", field)
		case "oneof":
			errs[field] = fmt.Sprintf("Kolom %s harus salah satu dari: %s", field, strings.ReplaceAll(fe.Param(), " ", ", "))
		case "numeric":
			errs[field] = fmt.Sprintf("Kolom %s harus berupa angka", field)
		case "gte":
			errs[field] = fmt.Sprintf("Kolom %s harus lebih besar dari atau sama dengan %s", field, fe.Param())
		case "lte":
			errs[field] = fmt.Sprintf("Kolom %s harus lebih kecil dari atau sama dengan %s", field, fe.Param())
		case "eqfield":
			errs[field] = fmt.Sprintf("Kolom %s tidak cocok", field)
		case "required_without":
			errs[field] = fmt.Sprintf("Kolom %s wajib diisi jika %s kosong", field, strings.ToLower(fe.Param()))
		default:
			errs[field] = "tidak valid"
		}
	}
	return errs
}
