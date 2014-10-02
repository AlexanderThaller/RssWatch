package service

import (
	"net/url"
	"os"
	"strconv"
	"time"
)

func ParseSettings(ur *url.URL) (set url.Values, err error) {
	set, err = url.ParseQuery(ur.Fragment)
	if err != nil {
		return
	}

	return
}

// ParseDuration will parse the duration for the given settings name.
// When the setting is not defined we will use the given default.
func ParseDuration(na, de string, se url.Values) (out time.Duration, def bool, err error) {
	u, def := GetSettingValue(na, de, se)

	out, err = time.ParseDuration(u)
	if err != nil {
		return
	}

	return
}

// ParseBool will parse the boolean for the given settings name. When
// the setting is not defined we will use the given default.
func ParseBool(na, de string, se url.Values) (out bool, def bool, err error) {
	u, def := GetSettingValue(na, de, se)

	out, err = strconv.ParseBool(u)
	if err != nil {
		return
	}

	return
}

// ParseFileMode will parse the os.FileMode for the given settings name.
// When the setting is not defined we will use the given default.
func ParseFileMode(na, de string, se url.Values) (out os.FileMode, def bool, err error) {
	u, def := GetSettingValue(na, de, se)

	o, err := strconv.ParseUint(u, 8, 32)
	if err != nil {
		return
	}

	out = os.FileMode(o)

	return
}

// GetSettingValue will get the value from the given settings. If the
// setting is not defined the function will return the given default
// value.
func GetSettingValue(na, de string, se url.Values) (out string, def bool) {
	out = de

	// Return default value if setting os not present.
	if _, d := se[na]; !d {
		def = true
		return
	}

	out = se.Get(na)
	def = false

	return
}

// ParseInt will parse the uint in base 10 for the given settings name.
// When the setting is not defined we will use the given default.
func ParseInt(na, de string, se url.Values) (out int, def bool, err error) {
	u, def := GetSettingValue(na, de, se)

	o, err := strconv.ParseInt(u, 10, 64)
	if err != nil {
		return
	}
	out = int(o)

	return
}

// ParseUint will parse the uint in base 10 for the given settings name.
// When the setting is not defined we will use the given default.
func ParseUint(na, de string, se url.Values) (out uint, def bool, err error) {
	u, def := GetSettingValue(na, de, se)

	o, err := strconv.ParseUint(u, 10, 64)
	if err != nil {
		return
	}
	out = uint(o)

	return
}
