package action

import "errors"

var (
	// ErrNameEmpty is returned when a definition has an empty name.
	ErrNameEmpty = errors.New("action: name must not be empty")

	// ErrNameInvalid is returned when a name contains characters that
	// are not valid in a JSON key (letters, digits, underscores, hyphens).
	ErrNameInvalid = errors.New("action: name is not a valid identifier")

	// ErrDescriptionTooLong is returned when description exceeds 255
	// characters.
	ErrDescriptionTooLong = errors.New("action: description must be 255 characters or fewer")

	// ErrCardinalityInvalid is returned when a cardinality value is not
	// Single or Multi.
	ErrCardinalityInvalid = errors.New("action: invalid cardinality")

	// ErrCardinalityUnknown is returned when a cardinality string cannot
	// be parsed.
	ErrCardinalityUnknown = errors.New("action: unknown cardinality")
)
