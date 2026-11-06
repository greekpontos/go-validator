package main

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func Validate(v any) error {
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	if val.Kind() != reflect.Struct {
		return errors.New("v must be a struct")
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}

		rules := strings.Split(tag, ";")
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)
			if rule == "" {
				continue
			}

			parts := strings.SplitN(rule, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid rule format: %s", rule)
			}

			key := parts[0]
			expectedValueStr := parts[1]

			switch key {
			case "min":
				limit, err := strconv.Atoi(expectedValueStr)
				if err != nil {
					return fmt.Errorf("invalid min value: %s", expectedValueStr)
				}

				switch fieldVal.Kind() {
				case reflect.String:
					runes := []rune(fieldVal.String())
					if len(runes) < limit {
						return fmt.Errorf("field %s length is less than %d", field.Name, limit)
					}
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if fieldVal.Int() < int64(limit) {
						return fmt.Errorf("field %s is less than %d", field.Name, limit)
					}
				default:
					return fmt.Errorf("min rule is not supported for type %s", fieldVal.Kind())
				}

			case "max":
				limit, err := strconv.Atoi(expectedValueStr)
				if err != nil {
					return fmt.Errorf("invalid max value: %s", expectedValueStr)
				}

				switch fieldVal.Kind() {
				case reflect.String:
					runes := []rune(fieldVal.String())
					if len(runes) > limit {
						return fmt.Errorf("field %s length is greater than %d", field.Name, limit)
					}
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if fieldVal.Int() > int64(limit) {
						return fmt.Errorf("field %s is greater than %d", field.Name, limit)
					}
				default:
					return fmt.Errorf("max rule is not supported for type %s", fieldVal.Kind())
				}

			case "regexp":
				if fieldVal.Kind() != reflect.String {
					return fmt.Errorf("regexp rule is only supported for string fields, got %s", fieldVal.Kind())
				}

				re, err := regexp.Compile(expectedValueStr)
				if err != nil {
					return fmt.Errorf("invalid regexp pattern: %s", expectedValueStr)
				}

				if !re.MatchString(fieldVal.String()) {
					return fmt.Errorf("field %s does not match pattern", field.Name)
				}

			default:
				return fmt.Errorf("unknown validation rule: %s", key)
			}
		}
	}
	return nil
}

type User struct {
	Name  string `validate:"min=3"`
	Age   int    `validate:"min=18;max=65"`
	Email string `validate:"regexp=^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`
}

func main() {
	if err := Validate(User{Name: "Ив", Age: 18, Email: "test@example.com"}); err != nil {
		fmt.Println("Validation error:", err)
	}
	if err := Validate(User{Name: "Иван", Age: 70, Email: "test@example.com"}); err != nil {
		fmt.Println("Validation error:", err)
	}
	if err := Validate(User{Name: "Иван", Age: 35, Email: "invalid email"}); err != nil {
		fmt.Println("Validation error:", err)
	}
	if err := Validate(User{Name: "Иван", Age: 35, Email: "test@example.com"}); err != nil {
		fmt.Println("Validation error:", err)
	} else {
		fmt.Println("User 4 is valid")
	}
}
