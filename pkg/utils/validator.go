package utils

import (
	"quiz_app/internal/models"
	"reflect"
)

func IsAnyUserFieldEmpty(admin models.Admin) bool {
	adminVal := reflect.ValueOf(admin)

	for i := 1; i < adminVal.NumField(); i++ {
		if adminVal.Field(i) == reflect.ValueOf("") {
			return true
		}
	}
	return false
}
