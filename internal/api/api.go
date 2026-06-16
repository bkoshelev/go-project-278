package api

type ValidationError struct {
	FieldName string
	Err       error
}

func (verr ValidationError) toJSON() map[string]any {
	return map[string]any{verr.FieldName: verr.Err.Error()}
}

func (e ValidationError) Error() string {
	return e.Err.Error()
}

var (
	InvalidErr = "invalid request"
)
