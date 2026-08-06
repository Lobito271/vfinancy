package masterdata

// errField is a small helper that returns a context-tagged error so
// callers can pinpoint which field is invalid.
func errField(s string) error { return fieldErr(s) }

type fieldErr string

func (e fieldErr) Error() string { return string(e) }
