# opt
[![CI Status](https://github.com/ifnotnil/opt/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/ifnotnil/opt/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/ifnotnil/opt/graph/badge.svg?token=eMp3iLkJ37)](https://codecov.io/gh/ifnotnil/opt)
[![Go Report Card](https://goreportcard.com/badge/github.com/ifnotnil/opt)](https://goreportcard.com/report/github.com/ifnotnil/opt)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/ifnotnil/opt)](https://pkg.go.dev/github.com/ifnotnil/opt)

Install: `go get -u github.com/ifnotnil/opt`

Golang generic optional type with support for [`json.Marshaler`](https://pkg.go.dev/encoding/json#Marshaler), [`json.Unmarshaler`](https://pkg.go.dev/encoding/json#Unmarshaler), [`driver.Valuer`](https://pkg.go.dev/database/sql/driver#Valuer), [`sql.Scanner`](https://pkg.go.dev/database/sql#Scanner) and golang 1.24 json [`omitzero`](https://tip.golang.org/doc/go1.24#:~:text=with%20the%20new-,omitzero,-option%20in%20the) tag.


The main ideas are:
  * Separate present-but-nil from not present (e.g., for the HTTP PATCH method).
  * Implement optional values without using pointers.

### States
| State  | Constructor function | Description                          |
|--------|----------------------|--------------------------------------|
| None   | `None[T]()`          | Value is absent, not preset nor nil. |
| Nil    | `Nil[T]()`           | Value is present but nil             |
| Valid  | `New[T](t T)`        | Value is preset and valid (not nil)  |


### Json unmarshaling

Given a struct 

```golang
type Foo struct {
	A Optional[int] `json:"a"`
}
```
| Json          | Resulted state of `A` |
|---------------|-----------------------|
| `{}`          | None                  |
| `{"a": null}` | Nil                   |
| `{"a": 123}`  | Valid                 |
