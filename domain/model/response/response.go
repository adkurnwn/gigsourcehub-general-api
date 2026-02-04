package response

import "net/http"

type Base struct {
	Status     int               `json:"status"`
	Message    string            `json:"message"`
	Validation map[string]string `json:"validation"`
	Data       interface{}       `json:"data"`
}

type List struct {
	List  []interface{} `json:"list"`
	Limit int64         `json:"limit"`
	Page  int64         `json:"page"`
	Total int64         `json:"total"`
}

// constructors / helpers

func Error(status int, message string) Base {
	return Base{
		Status:  status,
		Message: message,
	}
}

func ErrorValidation(validation map[string]string, message string) Base {
	return Base{
		Status:     http.StatusBadRequest,
		Message:    message,
		Validation: validation,
	}
}

func Success(data interface{}) Base {
	return Base{
		Status:  http.StatusOK,
		Message: "success",
		Data:    data,
	}
}
