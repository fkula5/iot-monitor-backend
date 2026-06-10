package validator

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name string `json:"name" validate:"required,min=3"`
	Age  int    `json:"age" validate:"gt=0"`
}

func TestValidateStruct(t *testing.T) {
	t.Run("Valid struct", func(t *testing.T) {
		s := TestStruct{Name: "John", Age: 30}
		errs := ValidateStruct(s)
		assert.Nil(t, errs)
	})

	t.Run("Invalid struct - missing required", func(t *testing.T) {
		s := TestStruct{Age: 30}
		errs := ValidateStruct(s)
		assert.NotNil(t, errs)
		assert.Equal(t, "is required", errs["name"])
	})

	t.Run("Invalid struct - min length", func(t *testing.T) {
		s := TestStruct{Name: "Jo", Age: 30}
		errs := ValidateStruct(s)
		assert.NotNil(t, errs)
		assert.Equal(t, "must be at least 3", errs["name"])
	})
}
