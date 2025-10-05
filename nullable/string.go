package nullable

import (
	"database/sql"
	"encoding/json/v2"
)

// String in `nullable` package
// implements: sql.Scanner, json.Marshaler, json.Unmarshaler
type String struct {
	sql.NullString // implements sql.Scanner by embedding
}

// MarshalJSON outputs string or null
func (ns *String) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.String)
	}
	return []byte("null"), nil
}

// UnmarshalJSON parses string or null into nullable.String
func (ns *String) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		ns.Valid = false
		ns.String = ""
		return nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	ns.String = str
	ns.Valid = true
	return nil
}
