package openpoiapi_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/yone-k/openpoiapi-go"
)

func TestAPIError_Error(t *testing.T) {
	t.Run("ステータスコードとメッセージを含む文字列を返す", func(t *testing.T) {
		err := &openpoiapi.APIError{StatusCode: 429, Message: "Too Many Requests"}

		want := "openpoiapi: status 429: Too Many Requests"
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("包まれていてもerrors.Asで取り出せる", func(t *testing.T) {
		wrapped := fmt.Errorf("wrap: %w", &openpoiapi.APIError{StatusCode: 400, Message: "bad"})

		var apiErr *openpoiapi.APIError
		if !errors.As(wrapped, &apiErr) {
			t.Fatalf("errors.As() = false, want true")
		}
		if apiErr.StatusCode != 400 || apiErr.Message != "bad" {
			t.Errorf("APIError = %+v, want {StatusCode:400 Message:bad}", *apiErr)
		}
	})
}

func TestValidationError_Error(t *testing.T) {
	t.Run("項目名と理由を含む文字列を返す", func(t *testing.T) {
		err := &openpoiapi.ValidationError{Field: "center", Reason: "x"}

		want := "openpoiapi: invalid center: x"
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("包まれていてもerrors.Asで取り出せる", func(t *testing.T) {
		wrapped := fmt.Errorf("wrap: %w", &openpoiapi.ValidationError{Field: "radius", Reason: "y"})

		var validationErr *openpoiapi.ValidationError
		if !errors.As(wrapped, &validationErr) {
			t.Fatalf("errors.As() = false, want true")
		}
		if validationErr.Field != "radius" || validationErr.Reason != "y" {
			t.Errorf("ValidationError = %+v, want {Field:radius Reason:y}", *validationErr)
		}
	})
}
