package normalizer

import (
	"bytes"
	"errors"
	"io"
	"sort"
	"unicode/utf8"
)

var JsonSyntaxError = errors.New("Syntax error")

func Normalize(src []byte) ([]byte, error) {
	r := bytes.NewReader(src)
	return parseValue(r)
}

func skipFillers(r *bytes.Reader) error {
	for {
		if c, err := r.ReadByte(); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		} else if c == ' ' || c == '\n' || c == '\r' || c == '\t' {
			continue
		}

		r.UnreadByte()
		return nil
	}
}

func parseName(r *bytes.Reader) (string, error) {
	var name []byte

	if c, err := r.ReadByte(); err != nil {
		return "", err
	} else if c != '"' {
		return "", JsonSyntaxError
	}

	if buf, err := parseString(r); err != nil {
		return "", err
	} else {
		name = buf
	}

	if err := skipFillers(r); err != nil {
		return "", err
	}

	if c, err := r.ReadByte(); err != nil {
		return "", err
	} else if c != ':' {
		return "", JsonSyntaxError
	}

	if err := skipFillers(r); err != nil {
		return "", err
	}

	return string(name), nil
}

func parseValue(r *bytes.Reader) ([]byte, error) {
	if c, err := r.ReadByte(); err != nil {
		return nil, err
	} else {
		switch c {
		case '{':
			if data, err := parseObject(r); err != nil {
				return nil, err
			} else {
				return data, nil
			}
		case '[':
			if data, err := parseArray(r); err != nil {
				return nil, err
			} else {
				return data, nil
			}
		case '"':
			if data, err := parseString(r); err != nil {
				return nil, err
			} else {
				return data, nil
			}
		case 'n':
			if data, err := parseNull(r); err != nil {
				return nil, err
			} else {
				return data, nil
			}
		case 't':
			fallthrough
		case 'f':
			if data, err := parseBool(c, r); err != nil {
				return nil, err
			} else {
				return data, nil
			}
		default:
			if c >= '0' && c <= '9' {
				r.UnreadByte()
				if data, err := parseNumber(r); err != nil {
					return nil, err
				} else {
					return data, nil
				}
			} else {
				return nil, JsonSyntaxError
			}
		}
	}
}

func parseObject(r *bytes.Reader) ([]byte, error) {
	type _ObjItem struct {
		name  string
		value []byte
	}
	obj := make([]_ObjItem, 0, 16)

	for {
		var name string

		if err := skipFillers(r); err != nil {
			return nil, err
		}
		if val, err := parseName(r); err != nil {
			return nil, err
		} else {
			if val == "" {
				return nil, JsonSyntaxError
			}
			name = val
		}

		if err := skipFillers(r); err != nil {
			return nil, err
		}
		if val, err := parseValue(r); err != nil {
			return nil, err
		} else {
			if val == nil {
				return nil, JsonSyntaxError
			}
			obj = append(obj, _ObjItem{name: name, value: val})
		}

		if err := skipFillers(r); err != nil {
			return nil, err
		}

		if c, err := r.ReadByte(); err != nil {
			return nil, err
		} else {
			if c == ',' {
				continue
			} else if c == '}' {
				break
			}
			return nil, JsonSyntaxError
		}
	}

	sort.Slice(obj, func(i, j int) bool {
		return obj[i].name < obj[j].name
	})

	data := make([]byte, 1, 256) // TODO bytes.Buffer?
	data[0] = '{'
	first := true
	for _, it := range obj {
		if first {
			first = false
		} else {
			data = append(data, ',')
		}
		data = append(data, it.name...)
		data = append(data, ':')
		data = append(data, it.value...)
	}
	data = append(data, '}')

	return data, nil
}

func parseArray(r *bytes.Reader) ([]byte, error) {
	data := make([]byte, 1, 256) // TODO bytes.Buffer?
	data[0] = '['

	for {
		if err := skipFillers(r); err != nil {
			return nil, err
		}
		if val, err := parseValue(r); err != nil {
			return nil, err
		} else {
			if val == nil {
				return nil, JsonSyntaxError
			}
			if len(data) > 1 {
				data = append(data, ',')
			}
			data = append(data, val...)
		}

		if err := skipFillers(r); err != nil {
			return nil, err
		}

		if c, err := r.ReadByte(); err != nil {
			return nil, err
		} else {
			if c == ',' {
				continue
			} else if c == ']' {
				data = append(data, ']')
				return data, nil
			}
			return nil, JsonSyntaxError
		}
	}
}

func parseString(r *bytes.Reader) ([]byte, error) {
	buf := make([]byte, 1, 128)
	escaping := false

	buf[0] = '"'

	for {
		ch, _, err := r.ReadRune()
		if err != nil {
			return nil, err
		}

		var chBuf [4]byte
		size := utf8.EncodeRune(chBuf[:], ch)
		buf = append(buf, chBuf[:size]...)

		if ch == '\\' {
			if escaping {
				escaping = false
			} else {
				escaping = true
			}
		} else {
			if ch == '"' {
				if !escaping {
					return buf, nil
				}
			}
			escaping = false
		}
	}

	return nil, nil
}

func parseBool(startByte byte, r *bytes.Reader) ([]byte, error) {
	var buf []byte
	if startByte == 't' {
		buf = []byte("true")
	} else {
		buf = []byte("false")
	}
	for _, expected := range buf[1:] {
		c, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		if c != expected {
			return nil, JsonSyntaxError
		}
	}
	return buf, nil
}

func parseNull(r *bytes.Reader) ([]byte, error) {
	buf := []byte("null")
	for _, expected := range buf[1:] {
		c, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		if c != expected {
			return nil, JsonSyntaxError
		}
	}
	return buf, nil
}

func parseNumber(r *bytes.Reader) ([]byte, error) {
	buf := make([]byte, 0, 32)
	firstPoint := true

	for {
		c, err := r.ReadByte()
		if err != nil {
			if err == io.EOF && len(buf) != 0 {
				return buf, nil
			} else {
				return nil, err
			}
		}

		if c >= '0' && c <= '9' {
			buf = append(buf, c)
		} else if c == '.' && firstPoint {
			buf = append(buf, c)
			firstPoint = false
		} else if c == ',' || c == ']' || c == '}' || c == ' ' {
			r.UnreadByte()
			return buf, nil
		} else {
			return nil, JsonSyntaxError
		}
	}
}
