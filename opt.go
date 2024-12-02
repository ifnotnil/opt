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

type Optional[T any] struct {
	Item      T
	isPresent bool
	isNil     bool
}

func (e Optional[T]) OrElse(d T) T { //nolint:ireturn
	if e.Valid() {
		return e.Item
	}

	return d
}

func (e Optional[T]) Valid() bool   { return e.isPresent && !e.isNil }
func (e Optional[T]) Nil() bool     { return e.isNil }
func (e Optional[T]) Present() bool { return e.isPresent }

func (e Optional[T]) MarshalJSON() ([]byte, error) {
	if !e.isPresent || e.isNil {
		return nullBytes, nil
	}

	return json.Marshal(e.Item)
}

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

// Scan implements sql.Scanner interface.
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

// Value implements driver.Valuer interface.
//
//	int64
//	float64
//	bool
//	[]byte
//	string
//	time.Time
func (e Optional[T]) Value() (driver.Value, error) {
	if !e.isPresent || e.isNil {
		return nil, nil //nolint:nilnil
	}

	return e.Item, nil
}
