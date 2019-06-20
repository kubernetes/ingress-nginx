package seq

import (
	"fmt"
	"reflect"
	"sort"
)

// Compare returns 0 if a and b are equal, -1 if a < b, or 1 if a > b.
// Panics if a and b are not of the same type, or are of a type not listed here.
//   * Bools are compared assuming false < true.
//   * Strings, integer and float values are compared as Go compares them.
//   * Two nil pointers are equal; one nil pointer is treated as smaller than a non-nil pointer.
//   * Non-nil pointers are compared by comparing the values they point to.
//   * Structures are compared by comparing their fields in order.
//   * Slices are compared by comparing elements sequentially.  If the slices are of different
//   length and all elements are the same up to the shorter length, the shorter slice is treated as
//   smaller.
//   * Maps can only be compared if they have string keys, in which case the ordered list of
//   keys are first compared as string slices, and if they're equal then the values are compared
//   sequentially in key order.
func Compare(a, b interface{}) int {
	return compareValue(reflect.ValueOf(a), reflect.ValueOf(b))
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func compareValue(ir1, ir2 reflect.Value) int {
	var zerovalue reflect.Value
	r1, r2 := reflect.Indirect(ir1), reflect.Indirect(ir2)

	if r1 == zerovalue {
		if r2 == zerovalue {
			return 0
		}
		return -1
	}
	if r2 == zerovalue {
		return 1
	}

	switch r1.Kind() {
	case reflect.Bool:
		v1, v2 := boolToInt(r1.Bool()), boolToInt(r2.Bool())
		if v1 < v2 {
			return -1
		}
		if v1 > v2 {
			return 1
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v1, v2 := r1.Int(), r2.Int()
		if v1 < v2 {
			return -1
		}
		if v1 > v2 {
			return 1
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		v1, v2 := r1.Uint(), r2.Uint()
		if v1 < v2 {
			return -1
		}
		if v1 > v2 {
			return 1
		}
	case reflect.Float32, reflect.Float64:
		v1, v2 := r1.Float(), r2.Float()
		if v1 < v2 {
			return -1
		}
		if v1 > v2 {
			return 1
		}
	case reflect.Map:
		return compareMap(r1, r2)
	case reflect.Struct:
		return compareStruct(r1, r2)
	case reflect.Slice:
		if r1.Type().Elem().Kind() == reflect.Uint8 {
			// Not using bytes.Compare because that fails on unexported fields:
			// return bytes.Compare(r1.Interface().([]byte), r2.Interface().([]byte))
			var s string
			strtype := reflect.TypeOf(s)
			v1, v2 := r1.Convert(strtype).String(), r2.Convert(strtype).String()
			if v1 < v2 {
				return -1
			}
			if v1 > v2 {
				return 1
			}
		}
		return compareSlice(r1, r2)
	case reflect.String:
		v1, v2 := r1.String(), r2.String()
		if v1 < v2 {
			return -1
		}
		if v1 > v2 {
			return 1
		}
	default:
		panic(fmt.Sprintf("don't know how to compare values of type %v", r1.Type()))
	}
	return 0
}

func compareStruct(r1, r2 reflect.Value) int {
	if r1.Type() != r2.Type() {
		panic(fmt.Sprintf("s1 and s2 are not of the same type: %v, %v", r1.Type(), r2.Type()))
	}

	n := r1.NumField()
	for i := 0; i < n; i++ {
		c := compareValue(r1.Field(i), r2.Field(i))
		if c != 0 {
			return c
		}
	}
	return 0
}

func compareSlice(r1, r2 reflect.Value) int {
	maxlen := r1.Len()
	if r2.Len() < maxlen {
		maxlen = r2.Len()
	}
	for i := 0; i < maxlen; i++ {
		c := compareValue(r1.Index(i), r2.Index(i))
		if c != 0 {
			return c
		}
	}
	if r1.Len() > maxlen {
		return 1
	}
	if r2.Len() > maxlen {
		return -1
	}
	return 0
}

func sortedKeys(r1 reflect.Value) []string {
	keys := make([]string, 0, r1.Len())
	for _, k := range r1.MapKeys() {
		keys = append(keys, k.String())
	}
	sort.Strings(keys)
	return keys
}

func compareMap(r1, r2 reflect.Value) int {
	if r1.Type().Key().Kind() != reflect.String {
		panic("can only compare maps with keys of type string")
	}

	s1, s2 := sortedKeys(r1), sortedKeys(r2)
	c := compareSlice(reflect.ValueOf(s1), reflect.ValueOf(s2))
	if c != 0 {
		return c
	}

	for _, k := range s1 {
		vk := reflect.ValueOf(k)
		c := compareValue(r1.MapIndex(vk), r2.MapIndex(vk))
		if c != 0 {
			return c
		}
	}
	return 0
}
