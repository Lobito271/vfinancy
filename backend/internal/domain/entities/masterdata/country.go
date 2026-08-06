package masterdata

// Country is a small ISO 3166 lookup. It is global (no CompanyID).
type Country struct {
	Code                  string
	Name                  string
	Locale                string
	Currency              string
	TaxIDLabel             string
	PersonalIDLabel        string
	DefaultDocumentTypes  []string
	CreatedAt             string  // ISO timestamp; string to keep this a value object
}

// NewCountry validates the code and constructs a Country.
func NewCountry(code, name, locale, currency, taxLabel, personalLabel string, docTypes []string) (*Country, error) {
	if len(code) != 2 {
		return nil, errField("country code must be 2 letters")
	}
	if name == "" {
		return nil, errField("country name is required")
	}
	return &Country{
		Code:                 code,
		Name:                 name,
		Locale:               locale,
		Currency:             currency,
		TaxIDLabel:            taxLabel,
		PersonalIDLabel:       personalLabel,
		DefaultDocumentTypes: docTypes,
	}, nil
}
