package normalizer

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
)

func TestParseString(t *testing.T) {
	check := func(src, expected string, expectedError error) {
		r := bytes.NewReader([]byte(src))
		data, err := parseString(r)
		if err != expectedError {
			t.Errorf("%v != %v, src: %s", err, expectedError, src)
		} else if val := string(data); val != expected {
			t.Errorf("%v != %v", val, expected)
		}
	}

	check(`1"`, `"1"`, nil)
	check(`abc"`, `"abc"`, nil)
	check(`a\"bc"`, `"a\"bc"`, nil)
	check(`"123`, `""`, nil)

	check(`xyz`, ``, io.EOF)
}

func TestParseBool(t *testing.T) {
	check := func(src, expected string, expectedError error) {
		r := bytes.NewReader([]byte(src[1:]))
		data, err := parseBool(src[0], r)
		if err != expectedError {
			t.Errorf("%v != %v, src: %s", err, expectedError, src)
		} else if val := string(data); val != expected {
			t.Errorf("%v != %v", val, expected)
		}
	}

	check(`true`, `true`, nil)
	check(`false`, `false`, nil)
	check(`a\"bc"`, ``, JsonSyntaxError)
	check(`null`, ``, JsonSyntaxError)

	check(`t`, ``, io.EOF)
}

func TestParseNull(t *testing.T) {
	check := func(src, expected string, expectedError error) {
		r := bytes.NewReader([]byte(src))
		data, err := parseNull(r)
		if err != expectedError {
			t.Errorf("%v != %v, src: %s", err, expectedError, src)
		} else if val := string(data); val != expected {
			t.Errorf("%v != %v", val, expected)
		}
	}

	check(`ull`, `null`, nil)
	check(`false`, ``, JsonSyntaxError)
	check(`a\"bc"`, ``, JsonSyntaxError)
	check(``, ``, io.EOF)
}

func TestParseNumber(t *testing.T) {
	check := func(src, expected string, expectedError error) {
		r := bytes.NewReader([]byte(src))
		data, err := parseNumber(r)
		if err != expectedError {
			t.Errorf("%v != %v, src: %s", err, expectedError, src)
		} else if val := string(data); val != expected {
			t.Errorf("%v != %v", val, expected)
		}
	}

	check(`123`, `123`, nil)
	check(`123.456`, `123.456`, nil)
	check(`a\"bc"`, ``, JsonSyntaxError)
	check(`1.2.3"`, ``, JsonSyntaxError)
	check(``, ``, io.EOF)
}

func TestParseName(t *testing.T) {
	check := func(src, expected string, expectedError error) {
		r := bytes.NewReader([]byte(src))
		data, err := parseName(r)
		if err != expectedError {
			t.Errorf("%v != %v, src: %s", err, expectedError, src)
		} else if val := string(data); val != expected {
			t.Errorf("%v != %v", val, expected)
		}
	}

	check(`"1":`, `"1"`, nil)
	check(`"abc":`, `"abc"`, nil)
	check(`"a\"bc"  :  `, `"a\"bc"`, nil)
	check(`"xyz"`, ``, io.EOF)
	check(`xyz`, ``, JsonSyntaxError)
	check(`"xyz",`, ``, JsonSyntaxError)
	check(`"xyz"}`, ``, JsonSyntaxError)
	check(`"xyz"]`, ``, JsonSyntaxError)
}

func TestParseArray(t *testing.T) {
	check := func(src, expected string, expectedError error) {
		r := bytes.NewReader([]byte(src))
		data, err := parseArray(r)
		if err != expectedError {
			t.Errorf("%v != %v, src: %s", err, expectedError, src)
		} else if val := string(data); val != expected {
			t.Errorf("%v != %v", val, expected)
		}
	}

	check(`1]`, `[1]`, nil)
	check(`1,2]`, `[1,2]`, nil)
	check(`1, 2]`, `[1,2]`, nil)
	check(`  "1" ]`, `["1"]`, nil)
	check(`  "1", 2  , "3" ]`, `["1",2,"3"]`, nil)

	check("  1, [2, \n 3]]", `[1,[2,3]]`, nil)

	check(`1`, ``, io.EOF)
	check(`1}`, ``, JsonSyntaxError)
	check(`1,,]`, ``, JsonSyntaxError)
}

func TestParseObject(t *testing.T) {
	check := func(src, expected string, expectedError error) {
		r := bytes.NewReader([]byte(src))
		data, err := parseObject(r)
		if err != expectedError {
			t.Errorf("%v != %v, src: %s", err, expectedError, src)
		} else if val := string(data); val != expected {
			t.Errorf("%v != %v", val, expected)
		}
	}

	check(`"a":1}`, `{"a":1}`, nil)
	check(`"a":1, "b": "c" }`, `{"a":1,"b":"c"}`, nil)
	check(`"a": 1, "x": {"b": "c"} }`, `{"a":1,"x":{"b":"c"}}`, nil)
	check(`"a": 1, "x": {"b": ["c"]} }`, `{"a":1,"x":{"b":["c"]}}`, nil)
	check(`"x": 1, "a": [{"b": "c"}] }`, `{"a":[{"b":"c"}],"x":1}`, nil)
	check(`"x": 1, "a": [{"b": "c", "a": 1}] }`, `{"a":[{"a":1,"b":"c"}],"x":1}`, nil)

	check(`"c": 1, "a": 3, "b": 2}`, `{"a":3,"b":2,"c":1}`, nil)

	/*
		check(`1,2]`, `[1,2]`, nil)
		check(`1, 2]`, `[1,2]`, nil)
		check(`  "1" ]`, `["1"]`, nil)
		check(`  "1", 2  , "3" ]`, `["1",2,"3"]`, nil)

		check("  1, [2, \n 3]]", `[1,[2,3]]`, nil)

		check(`1`, ``, io.EOF)
		check(`1}`, ``, JsonSyntaxError)
		check(`1,,]`, ``, JsonSyntaxError)
	*/
}

func TestParseValue(t *testing.T) {
	check := func(src, expected string, expectedError error) {
		r := bytes.NewReader([]byte(src))
		data, err := parseValue(r)
		if err != expectedError {
			t.Errorf("%v != %v, src: %s", err, expectedError, src)
		} else if val := string(data); val != expected {
			t.Errorf("%v != %v", val, expected)
		}
	}

	check(`null`, `null`, nil)
	check(`true`, `true`, nil)
	check(`345`, `345`, nil)
	check(`345.7`, `345.7`, nil)
	check(`"abc"`, `"abc"`, nil)
	check(`[1, 3, 2]`, `[1,3,2]`, nil)
	check(`{"a":1}`, `{"a":1}`, nil)
	check(`{"b": "c", "a": 1 }`, `{"a":1,"b":"c"}`, nil)
}

func BenchmarkParseNull(b *testing.B) {
	r := bytes.NewReader([]byte("null"))

	for i := 0; i < b.N; i++ {
		r.Seek(0, io.SeekStart)
		_, err := parseValue(r)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseNumber(b *testing.B) {
	r := bytes.NewReader([]byte("12345.456"))

	for i := 0; i < b.N; i++ {
		r.Seek(0, io.SeekStart)
		_, err := parseValue(r)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseString(b *testing.B) {
	r := bytes.NewReader([]byte(`"abc 123 xyz"`))

	for i := 0; i < b.N; i++ {
		r.Seek(0, io.SeekStart)
		_, err := parseValue(r)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseIntArray(b *testing.B) {
	r := bytes.NewReader([]byte(`[1, 2, 3, 4, 5]`))

	for i := 0; i < b.N; i++ {
		r.Seek(0, io.SeekStart)
		_, err := parseValue(r)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseStringArray(b *testing.B) {
	r := bytes.NewReader([]byte(`["1", "2", "3", "4", "5"]`))

	for i := 0; i < b.N; i++ {
		r.Seek(0, io.SeekStart)
		_, err := parseValue(r)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseObject(b *testing.B) {
	r := bytes.NewReader([]byte(`{"b": 1, "a": "xyz", "d": {"y": 2, "x": "z"}, "c": [1, 3, 2]}`))

	for i := 0; i < b.N; i++ {
		r.Seek(0, io.SeekStart)
		_, err := parseValue(r)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseObjectToMap(b *testing.B) {
	src := []byte(`{"b": 1, "a": "xyz", "d": {"y": 2, "x": "z"}, "c": [1, 3, 2]}`)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		m := map[string]interface{}{}
		b.StartTimer()

		err := json.Unmarshal(src, &m)
		if err != nil {
			b.Fatal(err)
		}
	}
}
