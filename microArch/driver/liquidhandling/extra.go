package liquidhandling

import (
	"reflect"

	"github.com/Synthace/antha/antha/anthalib/wunit"
)

func SafeGetF64(m map[string]interface{}, key string) float64 {
	ret := 0.0

	v, ok := m[key]

	ok = ok && reflect.TypeOf(v) == reflect.TypeOf(ret)
	if ok {
		ret = v.(float64)
	}

	return ret
}

func SafeGetBool(m map[string]interface{}, key string) bool {
	ret := false

	v, ok := m[key]

	ok = ok && reflect.TypeOf(v) == reflect.TypeOf(ret)
	if ok {
		ret = v.(bool)
	}

	return ret

}

func SafeGetString(m map[string]interface{}, key string) string {
	ret := ""

	v, ok := m[key]

	ok = ok && reflect.TypeOf(v) == reflect.TypeOf(ret)
	if ok {
		ret = v.(string)
	}

	return ret

}

func SafeGetInt(m map[string]interface{}, key string) int {
	ret := 0

	v, ok := m[key]

	ok = ok && reflect.TypeOf(v) == reflect.TypeOf(ret)
	if ok {
		ret = v.(int)
	}

	return ret
}

func SafeGetVolume(m map[string]interface{}, key string) wunit.Volume {
	ret := wunit.NewVolume(0.0, "ul")

	v, ok := m[key]

	ok = ok && reflect.TypeOf(v) == reflect.TypeOf(ret)
	if ok {
		ret = v.(wunit.Volume)
	}

	return ret
}
