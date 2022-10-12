package cashu

import "encoding/json"

type ErrorResponse struct {
	Err  string `json:"error"`
	Code int    `json:"code"`
}
type ErrorOptions func(err *ErrorResponse)

func WithCode(code int) ErrorOptions {
	return func(err *ErrorResponse) {
		err.Code = code
	}
}
func NewErrorResponse(err error, options ...ErrorOptions) ErrorResponse {
	e := ErrorResponse{
		Err: err.Error(),
	}
	for _, o := range options {
		o(&e)
	}
	return e
}

func (e ErrorResponse) String() string {
	return ToJson(e)
}

func (e ErrorResponse) Error() string {
	return e.Err
}

func ToJson(i interface{}) string {
	b, err := json.Marshal(i)
	if err != nil {
		return err.Error()
	}
	return string(b)
}
