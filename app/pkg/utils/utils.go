package utils

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

func FormatQuery(q string) string {
	return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(q, "\t", ""), "\n", " "))
}

func DoWithAttempts(fn func() error, maxAttempts int, delay time.Duration) error {
	var err error
	for maxAttempts > 0 {
		if err = fn(); err != nil {
			time.Sleep(delay)
			maxAttempts--
			continue
		}
		return nil
	}
	return err
}

func TranslateValidationError(err error, fieldName string) string {
	buffer := bytes.Buffer{}
	for _, v := range err.(validator.ValidationErrors) {
		if v.Field() != "" {
			fieldName = v.Field()
		}
		buffer.WriteString(fmt.Sprintf("Field validation for '%s' failed on the '%s' tag. ", fieldName, v.Tag()))
	}
	return buffer.String()
}
