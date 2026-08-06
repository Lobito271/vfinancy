package customer

func errField(s string) error { return fieldErr(s) }

type fieldErr string

func (e fieldErr) Error() string { return string(e) }
