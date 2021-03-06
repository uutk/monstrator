package main

import (
	"encoding"
	"encoding/json"
	"io"
	"os"
	"reflect"
)

var config = new(configuration)

type configuration struct {
	Token                string   `json:"token" env:"TOKEN"`
	Address              string   `json:"address" env:"ADDRESS"`
	Port                 string   `json:"port" env:"PORT"`
	ReadTimeout          duration `json:"readTimeout" env:"READ_TIMEOUT"`
	WriteTimeout         duration `json:"writeTimeout" env:"WRITE_TIMEOUT"`
	InlineQueryCacheTime duration `json:"inlineQueryCacheTime" env:"INLINE_QUERY_CACHE_TIME"`
	TLS                  struct {
		Certificate string `json:"certificate" env:"TLS_CERTIFICATE"`
		Key         string `json:"key" env:"TLS_KEY"`
	} `json:"TLS"`
	Shorteners struct {
		Google struct {
			APIKey  string   `json:"APIKey" env:"GOOGLE_SHORTENER_API_KEY"`
			Timeout duration `json:"timeout" env:"GOOGLE_SHORTENER_TIMEOUT"`
		} `json:"Google"`
		Isgd struct {
			Timeout duration `json:"timeout" env:"ISGD_SHORTENER_TIMEOUT"`
		} `json:"isgd"`
	} `json:"shorteners"`
}

func loadConfigFromJSON(r io.Reader) error {
	err := json.NewDecoder(r).Decode(config)
	return err
}

func loadConfigFromEnv() error {
	var unmarshal func(structValue reflect.Value) error
	unmarshal = func(structValue reflect.Value) error {
		structType := structValue.Type()
		for i := 0; i < structValue.NumField(); i++ {
			fieldValue := structValue.Field(i)
			field := structType.Field(i)
			kind := field.Type.Kind()
			if !fieldValue.IsValid() || !fieldValue.CanSet() {
				continue
			}
			if key, ok := field.Tag.Lookup("env"); ok {
				if value, ok := os.LookupEnv(key); ok {
					switch kind {
					case reflect.String:
						fieldValue.SetString(value)
					default:
						if u, ok := fieldValue.Addr().Interface().(encoding.TextUnmarshaler); ok {
							if err := u.UnmarshalText([]byte(value)); err != nil {
								return err
							}
						} else {
							return &envUnmarshalTypeError{structType, field}
						}
					}
				}
			} else if kind == reflect.Struct {
				if err := unmarshal(fieldValue); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return unmarshal(reflect.ValueOf(config).Elem())
}

type envUnmarshalTypeError struct {
	Struct reflect.Type
	Field  reflect.StructField
}

func (e *envUnmarshalTypeError) Error() string {
	return "cannot unmarshal environment variable into Go struct field " + e.Struct.Name() + "." + e.Field.Name + " of type " + e.Field.Type.String()
}
