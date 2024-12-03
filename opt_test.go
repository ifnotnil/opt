package opt

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:thelper
func TestJsonUnmarshal(t *testing.T) {
	type Foo struct {
		One Optional[int] `json:"one"`
	}

	tests := map[string]struct {
		asserts          func(t *testing.T, o Optional[int])
		inputJSONString  string
		expectedOptional Optional[int]
	}{
		"present and not null": {
			inputJSONString:  `{"one": 10}`,
			expectedOptional: Optional[int]{isPresent: true, isNil: false, Item: 10},
			asserts: func(t *testing.T, o Optional[int]) {
				val := 10
				assert.True(t, o.Valid())
				assert.Equal(t, val, o.OrElse(100))
				assert.Equal(t, &val, o.Ptr())
			},
		},
		"present and null": {
			inputJSONString:  `{"one": null}`,
			expectedOptional: Optional[int]{isPresent: true, isNil: true, Item: 0},
			asserts: func(t *testing.T, o Optional[int]) {
				assert.False(t, o.Valid())
				assert.Equal(t, 100, o.OrElse(100))
				assert.Nil(t, o.Ptr())
			},
		},
		"not present": {
			inputJSONString:  `{"two": 123}`,
			expectedOptional: Optional[int]{isPresent: false, isNil: false, Item: 0},
			asserts: func(t *testing.T, o Optional[int]) {
				assert.False(t, o.Valid())
				assert.Equal(t, 100, o.OrElse(100))
				assert.Nil(t, o.Ptr())
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			o := Foo{}
			err := json.Unmarshal([]byte(tc.inputJSONString), &o)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedOptional, o.One)
			if tc.asserts != nil {
				tc.asserts(t, o.One)
			}
		})
	}
}

func TestSqlValue(t *testing.T) {
	tests := map[string]struct {
		optValueFn  func() any
		expectValue any
	}{
		"int64 present not nil": {
			optValueFn: func() any {
				return New[int64](42)
			},
			expectValue: int64(42),
		},
		"int64 present nil": {
			optValueFn: func() any {
				return Nil[int64]()
			},
			expectValue: nil,
		},
		"int64 not present": {
			optValueFn: func() any {
				return Optional[int64]{}
			},
			expectValue: nil,
		},
		"int64 present not nil pointer": {
			optValueFn: func() any {
				n := New[int64](42)
				return n
			},
			expectValue: int64(42),
		},
		"int64 present nil pointer": {
			optValueFn: func() any {
				n := Nil[int64]()
				return n
			},
			expectValue: nil,
		},
		"int64 not pointer": {
			optValueFn: func() any {
				n := Nil[int64]()
				return &n
			},
			expectValue: nil,
		},

		"string present not nil": {
			optValueFn: func() any {
				return New[string]("test")
			},
			expectValue: "test",
		},
		"string present nil": {
			optValueFn: func() any {
				return Nil[string]()
			},
			expectValue: nil,
		},
		"string not present": {
			optValueFn: func() any {
				return Optional[string]{}
			},
			expectValue: nil,
		},

		"aValuer present not nil": {
			optValueFn: func() any {
				return New[aValuer](aValuer{})
			},
			expectValue: "value",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotErr := driver.DefaultParameterConverter.ConvertValue(tc.optValueFn())
			require.NoError(t, gotErr)
			assert.Equal(t, tc.expectValue, got)
		})
	}
}

type aValuer struct{}

func (aValuer) Value() (driver.Value, error) { return "value", nil }
