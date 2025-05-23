package opt

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"runtime"
	"strings"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:thelper
func TestJSONUnmarshal(t *testing.T) {
	type Foo struct {
		One Optional[int] `json:"one"`
	}

	t.Run("malformed", func(t *testing.T) {
		o := Foo{}
		err := json.Unmarshal([]byte(`{"one": {"a":"b"}}`), &o)
		require.Error(t, err)
		assert.False(t, o.One.Valid())
	})

	tests := map[string]struct {
		asserts          func(t *testing.T, o Optional[int])
		inputJSONString  string
		expectedOptional Optional[int]
	}{
		"present and not null": {
			inputJSONString:  `{"one": 10}`,
			expectedOptional: Optional[int]{state: StateValid, Item: 10},
			asserts: func(t *testing.T, o Optional[int]) {
				val := 10
				assert.True(t, o.Valid())
				assert.Equal(t, val, o.OrElse(100))
				assert.NotNil(t, o.Ptr())
				assert.Equal(t, &val, o.Ptr())
			},
		},
		"present and null": {
			inputJSONString:  `{"one": null}`,
			expectedOptional: Optional[int]{state: StateNil, Item: 0},
			asserts: func(t *testing.T, o Optional[int]) {
				assert.False(t, o.Valid())
				assert.Equal(t, 100, o.OrElse(100))
				assert.Nil(t, o.Ptr())
			},
		},
		"absent": {
			inputJSONString:  `{"two": 123}`,
			expectedOptional: Optional[int]{state: StateAbsent, Item: 0},
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

func TestJSONMarshal(t *testing.T) {
	type foo struct {
		One Optional[int] `json:"one"`
	}

	type fooOmitEmpty struct {
		One Optional[int] `json:"one,omitempty"`
	}

	// go 1.24
	type fooOmitZero struct {
		One Optional[int] `json:"one,omitzero"`
	}

	tests := map[string]struct {
		item         any
		expectedJSON string
		goVer        string
	}{
		"plain valid": {
			item: foo{
				One: New(123),
			},
			expectedJSON: `{"one":123}`,
		},
		"plain nil": {
			item: foo{
				One: Nil[int](),
			},
			expectedJSON: `{"one":null}`,
		},
		"plain absent": {
			item: foo{
				One: None[int](),
			},
			expectedJSON: `{"one":null}`,
		},

		"omit empty valid": {
			item: fooOmitEmpty{
				One: New(123),
			},
			expectedJSON: `{"one":123}`,
		},
		"omit empty nil": {
			item: fooOmitEmpty{
				One: Nil[int](),
			},
			expectedJSON: `{"one":null}`,
		},
		"omit empty absent": {
			item: fooOmitEmpty{
				One: None[int](),
			},
			expectedJSON: `{"one":null}`,
		},

		"omit zero valid": {
			item: fooOmitZero{
				One: New(123),
			},
			expectedJSON: `{"one":123}`,
			goVer:        ">= 1.24",
		},
		"omit zero nil": {
			item: fooOmitZero{
				One: Nil[int](),
			},
			expectedJSON: `{"one":null}`,
			goVer:        ">= 1.24",
		},
		"omit zero absent": {
			item: fooOmitZero{
				One: None[int](),
			},
			expectedJSON: `{}`,
			goVer:        ">= 1.24",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			goVersionGTE(t, tc.goVer)
			b, err := json.Marshal(tc.item)
			require.NoError(t, err)
			assert.JSONEq(t, tc.expectedJSON, string(b))
		})
	}
}

func TestSQLValue(t *testing.T) {
	tests := map[string]struct {
		optValueFn     func() any
		expectValue    any
		errorAssertion require.ErrorAssertionFunc
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
				return None[int64]()
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
				return New("test")
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
				return None[string]()
			},
			expectValue: nil,
		},

		"aValuer present not nil": {
			optValueFn: func() any {
				return New(aValuer{})
			},
			expectValue: "value",
		},

		"valuer returns error": {
			optValueFn: func() any {
				return New(errorValuer{})
			},
			expectValue:    nil,
			errorAssertion: require.Error,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotErr := driver.DefaultParameterConverter.ConvertValue(tc.optValueFn())
			if tc.errorAssertion == nil {
				tc.errorAssertion = require.NoError
			}
			tc.errorAssertion(t, gotErr)
			assert.Equal(t, tc.expectValue, got)
		})
	}
}

type aValuer struct{}

func (aValuer) Value() (driver.Value, error) { return "value", nil }

type errorValuer struct{}

func (errorValuer) Value() (driver.Value, error) { return nil, errors.New("value error") }

func TestSQLScan(t *testing.T) {
	t.Run("valid int64", func(t *testing.T) {
		o := Optional[int64]{}
		err := o.Scan(int64(123))
		require.NoError(t, err)
		require.True(t, o.Present())
		require.False(t, o.Nil())
		require.True(t, o.Valid())
		require.False(t, o.IsZero())
	})

	t.Run("null int64", func(t *testing.T) {
		o := Optional[int64]{}
		err := o.Scan(nil)
		require.NoError(t, err)
		require.True(t, o.Present())
		require.True(t, o.Nil())
		require.False(t, o.Valid())
		require.False(t, o.IsZero())
	})

	t.Run("scan error", func(t *testing.T) {
		o := Optional[int64]{}
		err := o.Scan(struct{}{})
		require.Error(t, err)
		require.False(t, o.Present())
		require.False(t, o.Nil())
		require.False(t, o.Valid())
		require.True(t, o.IsZero())
	})
}

func goVersionGTE(t *testing.T, semVerConstraint string) {
	t.Helper()
	if semVerConstraint == "" {
		return
	}

	versionNum := strings.TrimPrefix(runtime.Version(), "go")

	c, err := semver.NewConstraint(semVerConstraint)
	require.NoError(t, err, "error on go version constraint")

	c.Check(semver.MustParse(versionNum))

	if !c.Check(semver.MustParse(versionNum)) {
		t.Skipf("skipping test because of golang version constraint %s for go version %s", semVerConstraint, versionNum)
	}
}
