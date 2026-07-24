package dx

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// MustLoad inspects environment variables and populates struct T, raising a detailed panic if required values are missing.
func MustLoad[T any]() T {
	var cfg T
	val := reflect.ValueOf(&cfg).Elem()
	typ := val.Type()

	if typ.Kind() != reflect.Struct {
		panic("zyra.Env.MustLoad: generic type T must be a struct")
	}

	var missingKeys []string

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		structField := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		tag := structField.Tag.Get("env")
		envKey := strings.ToUpper(structField.Name)
		required := false
		defaultVal := ""

		if tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] != "" {
				envKey = parts[0]
			}
			for _, opt := range parts[1:] {
				opt = strings.TrimSpace(opt)
				if opt == "required" {
					required = true
				} else if strings.HasPrefix(opt, "default=") {
					defaultVal = strings.TrimPrefix(opt, "default=")
				}
			}
		}

		envVal := os.Getenv(envKey)
		if envVal == "" {
			if defaultVal != "" {
				envVal = defaultVal
			} else if required {
				missingKeys = append(missingKeys, envKey)
				continue
			}
		}

		if envVal != "" {
			if err := setFieldValue(field, envVal); err != nil {
				panic(fmt.Sprintf("zyra.Env.MustLoad: failed to set field '%s' from env '%s': %v", structField.Name, envKey, err))
			}
		}
	}

	if len(missingKeys) > 0 {
		panic(fmt.Sprintf("zyra.Env.MustLoad: missing required environment variables: %s", strings.Join(missingKeys, ", ")))
	}

	return cfg
}

func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Second) {
			d, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			field.SetInt(int64(d))
		} else {
			n, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			field.SetInt(n)
		}
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			parts := strings.Split(value, ",")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			field.Set(reflect.ValueOf(parts))
		}
	}
	return nil
}
