package forms

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
)

//Form creates a custom form struct,embeds a url.Value object
type Form struct {
	url.Values
	Errors errors
}

//Valid returns true if there are no errors (valid form)
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

//New initializes a form struct
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

//Required checks for required fields and sends a message if empty
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field can not be empty")
		}
	}
}

//HasARequiredField checks if the form field is not empty
func (f *Form) HasARequiredField(field string) bool {
	x := f.Get(field)
	if x == "" {
		//f.Errors.Add(field, "This field can not be empty")
		return false
	}
	return true
}

//MinLength checks if a field is a min length
func (f *Form) MinLength(field string, length int) bool {
	x := f.Get(field)
	if len(x) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long", length))
		return false
	}

	return true
}

//IsEmail checks for a valid email address
func (f *Form) IsEmail(field string) {
	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "Invalid Email Address")
	}
}
