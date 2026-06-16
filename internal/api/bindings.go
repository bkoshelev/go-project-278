package api

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// https://gin-gonic.com/en/docs/binding/bind-custom-unmarshaler/#using-bindunmarshaler
func (r *Range) UnmarshalParam(param string) error {
	var arr [2]int

	if err := json.Unmarshal([]byte(param), &arr); err != nil {
		return fmt.Errorf("invalid format, expected [start,end]")
	}

	if len(arr) != 2 {
		return fmt.Errorf("range must contain exactly 2 values")
	}

	r.Begin = arr[0]
	r.End = arr[1]

	return nil
}

func bindPayload(c *gin.Context, payload any) error {
	err := c.ShouldBindJSON(payload)

	if err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return ValidationError{ve[0].Field(), ve[0]}
		}

		return err
	}
	return nil
}
