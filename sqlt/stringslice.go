package sqlt

// https://gist.github.com/adharris/4163702#comment-1356268

import (
	"database/sql/driver"
	"encoding/csv"
	"errors"
	"regexp"
	"strings"
)

type StringSlice []string

var quoteEscapeRegex = regexp.MustCompile(`([^\\]([\\]{2})*)\\"`)

// Scan convert to a slice of strings
// http://www.postgresql.org/docs/9.1/static/arrays.html#ARRAYS-IO
func (s *StringSlice) Scan(src interface{}) error {
	asBytes, ok := src.([]byte)
	if !ok {
		return error(errors.New("Scan source was not []bytes"))
	}
	str := string(asBytes)

	// change quote escapes for csv parser
	str = quoteEscapeRegex.ReplaceAllString(str, `$1""`)
	str = strings.Replace(str, `\\`, `\`, -1)

	if str == "{}" {
		(*s) = make(StringSlice, 0)
		return nil
	}

	// remove braces
	str = str[1 : len(str)-1]
	csvReader := csv.NewReader(strings.NewReader(str))

	slice, err := csvReader.Read()

	if err != nil {
		return err
	}

	(*s) = StringSlice(slice)

	return nil
}

func (s StringSlice) Value() (driver.Value, error) {
	// string escapes.
	// \ => \\\
	// " => \"
	for i, elem := range s {
		s[i] = `"` + strings.Replace(strings.Replace(elem, `\`, `\\\`, -1), `"`, `\"`, -1) + `"`
	}
	return "{" + strings.Join(s, ",") + "}", nil
}
