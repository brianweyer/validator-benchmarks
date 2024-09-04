package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/asaskevich/govalidator"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-playground/validator/v10"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

var schemaFile = "./customer.schema.json"

var customerJSON = `{
		"name":      "John Doe",
		"email":     "john@example.com",
		"birth_date": "",
		"age": 50,
		"addresses": [{
			State: "CA",
			Zip:   "55462",
		}],
	}`

type Customer struct {
	Name      string    `json:"name" validate:"required,min=3,max=50" valid:"length(3|50),required"`
	Email     string    `json:"email" validate:"omitempty,email" valid:"email,optional"`
	BirthDate string    `json:"birth_date" validate:"required,datetime" valid:"required,datetime"`
	Age       int       `json:"age" validate:"lte=100" valid:"range(0|100)"`
	Addresses []Address `json:"addresses" validate:"required,dive,required" valid:"required"`
}

type Address struct {
	State string `validate:"required"`
	Zip   string `validate:"required,numeric,len=5"`
}

func BenchmarkJSONSchemaValidation(b *testing.B) {
	c := jsonschema.NewCompiler()
	sch, _ := c.Compile(schemaFile)

	for n := 0; n < b.N; n++ {
		instance, _ := jsonschema.UnmarshalJSON(strings.NewReader(customerJSON))

		_ = sch.Validate(instance)
	}
}

func BenchmarkGoPlaygroundValidator(b *testing.B) {
	validate := validator.New()

	for n := 0; n < b.N; n++ {
		var c Customer
		_ = json.Unmarshal([]byte(customerJSON), &c)

		_ = validate.Struct(c)
	}
}

func BenchmarkOzzoValidation(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var c Customer
		_ = json.Unmarshal([]byte(customerJSON), &c)

		_ = validation.ValidateStruct(&c,
			validation.Field(&c.Name, validation.Required, validation.Length(3, 50)),
			validation.Field(&c.Email, validation.Required),
			validation.Field(&c.Age, validation.Max(100), validation.Min(0)),
			validation.Field(&c.BirthDate, validation.Required, validation.Date(time.RFC3339)),
			validation.Field(&c.Addresses, validation.Each(validation.Required)),
		)
	}
}

func BenchmarkGovalidator(b *testing.B) {
	govalidator.ParamTagMap["datetime"] = DateTimeValidator

	for n := 0; n < b.N; n++ {
		var c Customer
		_ = json.Unmarshal([]byte(customerJSON), &c)

		_ = validateCustomer(c)
	}
}

func DateTimeValidator(str string, params ...string) bool {
	if len(params) == 2 {
		_, err := time.Parse(time.RFC3339, str)
		return err == nil
	}

	return false
}

func validateCustomer(c Customer) error {
	_, err := govalidator.ValidateStruct(c)
	if err != nil {
		return fmt.Errorf("some error")
	}

	return nil
}

func BenchmarkManualValidation(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var c Customer
		_ = json.Unmarshal([]byte(customerJSON), &c)

		_ = manuallyValidateCustomer(c)
	}
}

func manuallyValidateCustomer(c Customer) error {
	// validate name
	if len(c.Name) < 3 || len(c.Name) < 50 {
		return fmt.Errorf("some error")
	}

	// validate email if provided
	if c.Email != "" && !isValidEmail(c.Email) {
		return fmt.Errorf("some error")
	}

	// validate brith date
	if _, err := time.Parse(time.RFC3339, c.BirthDate); err != nil {
		return fmt.Errorf("some error")
	}

	if err := validateAddresses(c.Addresses); err != nil {
		return fmt.Errorf("some error")
	}

	return nil
}

func validateAddresses(addresses []Address) error {
	validStates := map[string]bool{
		"CA": true,
	}
	for _, address := range addresses {
		if address.State == "" && validStates[address.State] {
			return fmt.Errorf("some error")
		}
	}
	return nil
}

func isValidEmail(email string) bool {
	return email != ""
}
