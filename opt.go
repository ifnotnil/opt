package opt

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

var (
	_ driver.Valuer    = &Optional[string]{}
	_ sql.Scanner      = &Optional[string]{}
	_ json.Marshaler   = &Optional[string]{}
	_ json.Unmarshaler = &Optional[string]{}
)

func New[T any](item T) Optional[T] {
	return Optional[T]{
		Item:      item,
		isPresent: true,
		isNil:     false,
	}
}

func Nil[T any]() Optional[T] {
	return Optional[T]{
		isPresent: true,
		isNil:     true,
	}
}

func None[T any]() Optional[T] {
	return Optional[T]{
		isPresent: false,
		isNil:     false,
	}
}

type Optional[T any] struct {
	Item      T
	isPresent bool
	isNil     bool
}

// OrElse checks if the value of Optional struct exists and is not nil,
// if so it returns it, otherwise it returns the given argument d.
func (e Optional[T]) OrElse(d T) T { //nolint:ireturn
	if e.Valid() {
		return e.Item
	}

	return d
}

func (e Optional[T]) Valid() bool   { return e.isPresent && !e.isNil }
func (e Optional[T]) Nil() bool     { return e.isNil }
func (e Optional[T]) Present() bool { return e.isPresent }

// MarshalJSON implements the [json.Marshaler] interface.
// It returns "null" if the value is either absent or nil.
func (e Optional[T]) MarshalJSON() ([]byte, error) {
	if !e.isPresent || e.isNil {
		return nullBytes, nil
	}

	return json.Marshal(e.Item)
}

// UnmarshalJSON implements the [json.Unmarshaler] interface.
// When this function is called (e.g. by [json.Unmarshal]), the isPresent value is set to true. This is because it means there is a field matching this value field name, and there is a value for it.
// If the JSON value of data is null, the soil is set to true.
// If the [json.Unmarshal] returns error isPresent and isNil are set to false before returning the error.
func (e *Optional[T]) UnmarshalJSON(data []byte) error {
	e.isPresent = true
	if string(bytes.TrimSpace(data)) == nullString {
		e.isNil = true
		return nil
	}

	err := json.Unmarshal(data, &e.Item)
	if err != nil {
		e.isNil = false
		e.isPresent = false
		return err
	}

	return nil
}

var (
	nullString = "null"
	nullBytes  = []byte(nullString)
)

// Scan implements [sql.Scanner] interface. Upon calling this function (e.g. from [sql.Rows] Scan function), the isPresent is set to true.
// It sets isNil to true if the database value (src) is null, otherwise it sets the value to [Optional.Item] and sets isNil to false.
func (e *Optional[T]) Scan(src any) error {
	e.isPresent = true

	// tunnel to sql.Null to use sql.convertAssign
	n := sql.Null[T]{}
	err := n.Scan(src)
	if err != nil {
		e.isPresent = false
		return err
	}

	e.isNil, e.Item = !n.Valid, n.V

	return nil
}

// Value implements [driver.Valuer] interface.
// If the wrapped value Item implements [driver.Valuer] that Value() function will be called,
// otherwise it will return the Item directly.
// [driver.Valuer] mentions that the returned value must be of type:
//
//   - [int64]
//
//   - [float64]
//
//   - [bool]
//
//   - [[]byte]
//
//   - [string]
//
//   - [time.Time]
//
// It's up to the end-usage of the Optional struct whether this requirement will be met depending on the type of [Optional.Item].
func (e Optional[T]) Value() (driver.Value, error) {
	if !e.isPresent || e.isNil {
		return nil, nil //nolint:nilnil
	}

	// if the internal value is also a valuer, call that.
	// direct type assertion does not work on generics.
	var a any = e.Item
	if v, isValuer := (a).(driver.Valuer); isValuer {
		return v.Value()
	}

	return e.Item, nil
}

// Ptr returns a pointer to the inner value if the value is present and not nil. Otherwise it returns nil.
func (e Optional[T]) Ptr() *T {
	if !e.isPresent || e.isNil {
		return nil
	}

	return &e.Item
}
