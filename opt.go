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
		Item:  item,
		state: stateValid,
	}
}

func Nil[T any]() Optional[T] {
	return Optional[T]{
		state: stateNil,
	}
}

func None[T any]() Optional[T] {
	return Optional[T]{
		state: stateNone,
	}
}

// state represent the different states of an optional field.
//   - None ([stateNone])
//   - Present but nil ([stateNil])
//   - Present with value ([stateValid])
type state int8

const (
	stateNone  state = iota // Field is not preset.
	stateNil                // Field is present but nil.
	stateValid              // Field is present with a not nil value.
)

type Optional[T any] struct {
	Item  T
	state state
}

// Ptr returns a pointer to the inner value if the value is present and not nil. Otherwise it returns nil.
func (e Optional[T]) Ptr() *T {
	if e.IsValid() {
		return &e.Item
	}

	return nil
}

// OrElse checks if the value of Optional struct exists and is not nil,
// if so it returns it, otherwise it returns the given argument d.
func (e Optional[T]) OrElse(d T) T { //nolint:ireturn
	if e.IsValid() {
		return e.Item
	}

	return d
}

func (e Optional[T]) IsValid() bool   { return e.state == stateValid }
func (e Optional[T]) IsNil() bool     { return e.state == stateNil }
func (e Optional[T]) IsPresent() bool { return e.state > stateNone }

// IsZero implements the interface used by go 1.24 [encoding/json] marshall when `omitzero` tag is present.
func (e Optional[T]) IsZero() bool {
	return e.state == stateNone
}

// MarshalJSON implements the [json.Marshaler] interface.
// It returns "null" if the state is either none or nil.
func (e Optional[T]) MarshalJSON() ([]byte, error) {
	if e.IsValid() {
		return json.Marshal(e.Item)
	}

	return nullBytes, nil
}

// UnmarshalJSON implements the [json.Unmarshaler] interface.
// When this function is called (e.g. by [json.Unmarshal]), the isPresent value is set to true. This is because it means there is a field matching this value field name, and there is a value for it.
// If the JSON value of data is null, the soil is set to true.
// If the [json.Unmarshal] returns error isPresent and isNil are set to false before returning the error.
func (e *Optional[T]) UnmarshalJSON(data []byte) error {
	e.state = stateNil // state is > None
	if string(bytes.TrimSpace(data)) == nullString {
		return nil
	}

	err := json.Unmarshal(data, &e.Item)
	if err != nil {
		e.state = stateNone
		return err
	}

	e.state = stateValid

	return nil
}

var (
	nullString = "null"
	nullBytes  = []byte(nullString)
)

// Scan implements [sql.Scanner] interface. Upon calling this function (e.g. from [sql.Rows] Scan function), the isPresent is set to true.
// It sets isNil to true if the database value (src) is null, otherwise it sets the value to [Optional.Item] and sets isNil to false.
func (e *Optional[T]) Scan(src any) error {
	e.state = stateNil // state is > None

	// tunnel to sql.Null to use sql.convertAssign
	n := sql.Null[T]{}
	err := n.Scan(src)
	if err != nil {
		e.state = stateNone
		return err
	}

	if n.Valid {
		e.state = stateValid
		e.Item = n.V
	}

	return nil
}

// Value implements [driver.Valuer] interface.
// If the wrapped value Item implements [driver.Valuer] that Value() function will be called.
func (e Optional[T]) Value() (driver.Value, error) {
	if !e.IsValid() {
		return nil, nil //nolint:nilnil
	}

	var val any
	var err error

	val = e.Item

	// if the internal value is also a valuer, call that.
	// direct type assertion does not work on generics.
	if v, isValuer := (any(e.Item)).(driver.Valuer); isValuer {
		// this could panic if the item is nil pointer,
		// todo: consider sql.callValuerValue implementation (database/sql/convert.go)
		val, err = v.Value()
		if err != nil {
			return nil, err
		}
	}

	return driver.DefaultParameterConverter.ConvertValue(val)
}
