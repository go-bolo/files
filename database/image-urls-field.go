package files_database

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type ImageURLsField map[string]string

func (j ImageURLsField) ToJSON() []byte {
	jsonString, _ := json.Marshal(j)
	return jsonString
}

func (j *ImageURLsField) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal ImageURLsField JSON value:", value))
	}

	return json.Unmarshal(bytes, &j)
}

func (j ImageURLsField) Value() (driver.Value, error) {
	if len(j) == 0 {
		return []byte(`{}`), nil
	}
	return json.Marshal(j)
}

func NewImageURLsField() ImageURLsField {
	return ImageURLsField{}
}

func (j ImageURLsField) GetValue(key string) string {
	if v, ok := j[key]; ok {
		return v
	}
	return ""
}
