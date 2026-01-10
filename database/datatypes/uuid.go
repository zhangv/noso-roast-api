package datatypes

import (
	"database/sql/driver"

	"github.com/google/uuid"
)

//需要mysql8支持 参考来源：https://medium.com/wesionary-team/implementation-of-uuid-and-binary-16-in-gorm-v2-1c329c352c91
//MYTYPE -> new datatype
type UUID uuid.UUID

// StringToMYTYPE -> parse string to MYTYPE
func StringToMYTYPE(s string) (UUID, error) {
	id, err := uuid.Parse(s)
	return UUID(id), err
}

//String -> String Representation of Binary16
func (my UUID) String() string {
	return uuid.UUID(my).String()
}

//GormDataType -> sets type to binary(16)
func (my UUID) GormDataType() string {
	return "binary(16)"
}

func (my UUID) MarshalJSON() ([]byte, error) {
	s := uuid.UUID(my)
	str := "\"" + s.String() + "\""
	return []byte(str), nil
}

func (my *UUID) UnmarshalJSON(by []byte) error {
	s, err := uuid.ParseBytes(by)
	*my = UUID(s)
	return err
}

// Scan --> tells GORM how to receive from the database
func (my *UUID) Scan(value interface{}) error {

	bytes, _ := value.([]byte)
	parseByte, err := uuid.FromBytes(bytes)
	*my = UUID(parseByte)
	return err
}

// Value -> tells GORM how to save into the database
func (my UUID) Value() (driver.Value, error) {
	return uuid.UUID(my).MarshalBinary()
}
