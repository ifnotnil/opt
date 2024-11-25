package opt

import (
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
		inputJSONString  string
		expectedOptional Optional[int]
		asserts          func(t *testing.T, o Optional[int])
	}{
		"present and not null": {
			inputJSONString:  `{"one": 10}`,
			expectedOptional: Optional[int]{isPresent: true, isNil: false, Item: 10},
			asserts: func(t *testing.T, o Optional[int]) {
				assert.True(t, o.Valid())
				assert.Equal(t, 10, o.OrElse(100))
			},
		},
		"present and null": {
			inputJSONString:  `{"one": null}`,
			expectedOptional: Optional[int]{isPresent: true, isNil: true, Item: 0},
			asserts: func(t *testing.T, o Optional[int]) {
				assert.False(t, o.Valid())
				assert.Equal(t, 100, o.OrElse(100))
			},
		},
		"not present": {
			inputJSONString:  `{"two": 123}`,
			expectedOptional: Optional[int]{isPresent: false, isNil: false, Item: 0},
			asserts: func(t *testing.T, o Optional[int]) {
				assert.False(t, o.Valid())
				assert.Equal(t, 100, o.OrElse(100))
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
