package datatypes

import (
	"database/sql/driver"
	"encoding/json"
	"log"
	"strings"
)

type Slot []string

//第二种方式
// type Slot struct {
// 	Slots []string
// }

// GormDataType -> sets type to binary(16)
func (my Slot) GormDataType() string {
	return "varchar(512)"
}

// Scan --> tells GORM how to receive from the database
func (str *Slot) Scan(value interface{}) error {

	log.Println("value", value)
	if value == nil {
		*str = []string{}
		return nil
	}
	bytes, _ := value.([]byte)
	strvalue := string(bytes)
	if len(strvalue) < 2 {
		*str = []string{}
		return nil
	}
	tmp := strvalue[1 : len(strvalue)-1]
	*str = strings.Split(tmp, "|")
	// *str = Slot{Slots: strings.Split(tmp, "|")}
	return nil
}

// Value -> tells GORM how to save into the database
func (s Slot) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "||", nil
	}

	str := strings.Join(s, "|")
	return "|" + str + "|", nil
}

// UnmarshalJSON implements json.Unmarshaler interface
// This allows Slot to be unmarshaled from JSON array format
func (s *Slot) UnmarshalJSON(data []byte) error {
	// Handle null or empty JSON
	if len(data) == 0 || string(data) == "null" {
		*s = []string{}
		return nil
	}

	// Try to unmarshal as JSON array
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	*s = Slot(arr)
	return nil
}

// MarshalJSON implements json.Marshaler interface
// This allows Slot to be marshaled to JSON array format
func (s Slot) MarshalJSON() ([]byte, error) {
	if s == nil {
		return json.Marshal([]string{})
	}
	return json.Marshal([]string(s))
}
