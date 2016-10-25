package sjson

import (
	"bytes"
	"encoding/hex"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestInvalidPaths(t *testing.T) {
	var err error
	_, err = SetRaw(`{"hello":"world"}`, "", `"planet"`)
	if err == nil || err.Error() != "path cannot be empty" {
		t.Fatalf("expecting '%v', got '%v'", "path cannot be empty", err)
	}
	_, err = SetRaw("", "name.last.#", "")
	if err == nil || err.Error() != "array access character not allowed in path" {
		t.Fatalf("expecting '%v', got '%v'", "array access character not allowed in path", err)
	}
	_, err = SetRaw("", "name.last.\\1#", "")
	if err == nil || err.Error() != "array access character not allowed in path" {
		t.Fatalf("expecting '%v', got '%v'", "array access character not allowed in path", err)
	}
	_, err = SetRaw("", "name.las?t", "")
	if err == nil || err.Error() != "wildcard characters not allowed in path" {
		t.Fatalf("expecting '%v', got '%v'", "wildcard characters not allowed in path", err)
	}
	_, err = SetRaw("", "name.la\\s?t", "")
	if err == nil || err.Error() != "wildcard characters not allowed in path" {
		t.Fatalf("expecting '%v', got '%v'", "wildcard characters not allowed in path", err)
	}
	_, err = SetRaw("", "name.las*t", "")
	if err == nil || err.Error() != "wildcard characters not allowed in path" {
		t.Fatalf("expecting '%v', got '%v'", "wildcard characters not allowed in path", err)
	}
	_, err = SetRaw("", "name.las\\a*t", "")
	if err == nil || err.Error() != "wildcard characters not allowed in path" {
		t.Fatalf("expecting '%v', got '%v'", "wildcard characters not allowed in path", err)
	}
}

const (
	setRaw    = 1
	setBool   = 2
	setInt    = 3
	setFloat  = 4
	setString = 5
	setDelete = 6
)

func testRaw(t *testing.T, kind int, expect, json, path string, value interface{}) {
	var json2 string
	var err error
	switch kind {
	default:
		json2, err = Set(json, path, value)
	case setRaw:
		json2, err = SetRaw(json, path, value.(string))
	case setDelete:
		json2, err = Delete(json, path)
	}
	if err != nil {
		t.Fatal(err)
	} else if json2 != expect {
		t.Fatalf("expected '%v', got '%v'", expect, json2)
	}

	var json3 []byte
	switch kind {
	default:
		json3, err = SetBytes([]byte(json), path, value)
	case setRaw:
		json3, err = SetRawBytes([]byte(json), path, []byte(value.(string)))
	case setDelete:
		json3, err = DeleteBytes([]byte(json), path)
	}
	if err != nil {
		t.Fatal(err)
	} else if string(json3) != expect {
		t.Fatalf("expected '%v', got '%v'", expect, string(json3))
	}
}
func TestBasic(t *testing.T) {
	testRaw(t, setRaw, `[{"hiw":"planet","hi":"world"}]`, `[{"hi":"world"}]`, "0.hiw", `"planet"`)
	testRaw(t, setRaw, `[true]`, ``, "0", `true`)
	testRaw(t, setRaw, `[null,true]`, ``, "1", `true`)
	testRaw(t, setRaw, `[1,null,true]`, `[1]`, "2", `true`)
	testRaw(t, setRaw, `[1,true,false]`, `[1,null,false]`, "1", `true`)
	testRaw(t, setRaw,
		`[1,{"hello":"when","this":[0,null,2]},false]`,
		`[1,{"hello":"when","this":[0,1,2]},false]`,
		"1.this.1", `null`)
	testRaw(t, setRaw,
		`{"a":1,"b":{"hello":"when","this":[0,null,2]},"c":false}`,
		`{"a":1,"b":{"hello":"when","this":[0,1,2]},"c":false}`,
		"b.this.1", `null`)
	testRaw(t, setRaw,
		`{"a":1,"b":{"hello":"when","this":[0,null,2,null,4]},"c":false}`,
		`{"a":1,"b":{"hello":"when","this":[0,null,2]},"c":false}`,
		"b.this.4", `4`)
	testRaw(t, setRaw,
		`{"b":{"this":[null,null,null,null,4]}}`,
		``,
		"b.this.4", `4`)
	testRaw(t, setRaw,
		`[null,{"this":[null,null,null,null,4]}]`,
		``,
		"1.this.4", `4`)
	testRaw(t, setRaw,
		`{"1":{"this":[null,null,null,null,4]}}`,
		``,
		":1.this.4", `4`)
	testRaw(t, setRaw,
		`{":1":{"this":[null,null,null,null,4]}}`,
		``,
		"\\:1.this.4", `4`)
	testRaw(t, setRaw,
		`{":\1":{"this":[null,null,null,null,{".HI":4}]}}`,
		``,
		"\\:\\\\1.this.4.\\.HI", `4`)
	testRaw(t, setRaw,
		`{"b":{"this":{"😇":""}}}`,
		``,
		"b.this.😇", `""`)
	testRaw(t, setRaw,
		`[ 1,2  ,3]`,
		`  [ 1,2  ] `,
		"-1", `3`)
	testRaw(t, setInt, `[1234]`, ``, `0`, int64(1234))
	testRaw(t, setFloat, `[1234.5]`, ``, `0`, float64(1234.5))
	testRaw(t, setString, `["1234.5"]`, ``, `0`, "1234.5")
	testRaw(t, setBool, `[true]`, ``, `0`, true)
	testRaw(t, setBool, `[null]`, ``, `0`, nil)
}

func TestDelete(t *testing.T) {
	testRaw(t, setDelete, `[456]`, `[123,456]`, `0`, nil)
	testRaw(t, setDelete, `[123,789]`, `[123,456,789]`, `1`, nil)
	testRaw(t, setDelete, `[123,456]`, `[123,456,789]`, `-1`, nil)
	testRaw(t, setDelete, `{"a":[123,456]}`, `{"a":[123,456,789]}`, `a.-1`, nil)
	testRaw(t, setDelete, `{"and":"another"}`, `{"this":"that","and":"another"}`, `this`, nil)
	testRaw(t, setDelete, `{"this":"that"}`, `{"this":"that","and":"another"}`, `and`, nil)
	testRaw(t, setDelete, `{}`, `{"and":"another"}`, `and`, nil)
	testRaw(t, setDelete, `{"1":"2"}`, `{"1":"2"}`, `3`, nil)
}

// TestRandomData is a fuzzing test that throws random data at SetRaw
// function looking for panics.
func TestRandomData(t *testing.T) {
	var lstr string
	defer func() {
		if v := recover(); v != nil {
			println("'" + hex.EncodeToString([]byte(lstr)) + "'")
			println("'" + lstr + "'")
			panic(v)
		}
	}()
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 200)
	for i := 0; i < 2000000; i++ {
		n, err := rand.Read(b[:rand.Int()%len(b)])
		if err != nil {
			t.Fatal(err)
		}
		lstr = string(b[:n])
		SetRaw(lstr, "zzzz.zzzz.zzzz", "123")
	}
}

var json = `
{
    "sha": "d25341478381063d1c76e81b3a52e0592a7c997f",
    "commit": {
      "author": {
        "name": "Tom Tom Anderson",
        "email": "tomtom@anderson.edu",
        "date": "2013-06-22T16:30:59Z"
      },
      "committer": {
        "name": "Tom Tom Anderson",
        "email": "jeffditto@anderson.edu",
        "date": "2013-06-22T16:30:59Z"
      },
      "message": "Merge pull request #162 from stedolan/utf8-fixes\n\nUtf8 fixes. Closes #161",
      "tree": {
        "sha": "6ab697a8dfb5a96e124666bf6d6213822599fb40",
        "url": "https://api.github.com/repos/stedolan/jq/git/trees/6ab697a8dfb5a96e124666bf6d6213822599fb40"
      },
      "url": "https://api.github.com/repos/stedolan/jq/git/commits/d25341478381063d1c76e81b3a52e0592a7c997f",
      "comment_count": 0
    }
}
`
var path = "commit.committer.email"
var value = "tomtom@anderson.com"
var rawValue = `"tomtom@anderson.com"`
var rawValueBytes = []byte(rawValue)
var expect = strings.Replace(json, "jeffditto@anderson.edu", "tomtom@anderson.com", 1)
var jsonBytes = []byte(json)
var jsonBytes2 = []byte(json)
var expectBytes = []byte(expect)
var opts = &Options{Optimistic: true}
var optsInPlace = &Options{Optimistic: true, ReplaceInPlace: true}

func BenchmarkSet(t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		res, err := Set(json, path, value)
		if err != nil {
			t.Fatal(err)
		}
		if res != expect {
			t.Fatal("expected '%v', got '%v'", expect, res)
		}
	}
}

func BenchmarkSetRaw(t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		res, err := SetRaw(json, path, rawValue)
		if err != nil {
			t.Fatal(err)
		}
		if res != expect {
			t.Fatal("expected '%v', got '%v'", expect, res)
		}
	}
}

func BenchmarkSetBytes(t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		res, err := SetBytes(jsonBytes, path, value)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Compare(res, expectBytes) != 0 {
			t.Fatal("expected '%v', got '%v'", expect, res)
		}
	}
}

func BenchmarkSetRawBytes(t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		res, err := SetRawBytes(jsonBytes, path, rawValueBytes)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Compare(res, expectBytes) != 0 {
			t.Fatal("expected '%v', got '%v'", expect, res)
		}
	}
}

func BenchmarkSetOptimistic(t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		res, err := SetOptions(json, path, value, opts)
		if err != nil {
			t.Fatal(err)
		}
		if res != expect {
			t.Fatal("expected '%v', got '%v'", expect, res)
		}
	}
}

func BenchmarkSetInPlace(t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		res, err := SetOptions(json, path, value, optsInPlace)
		if err != nil {
			t.Fatal(err)
		}
		if res != expect {
			t.Fatal("expected '%v', got '%v'", expect, res)
		}
	}
}

func BenchmarkSetRawOptimistic(t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		res, err := SetRawOptions(json, path, rawValue, opts)
		if err != nil {
			t.Fatal(err)
		}
		if res != expect {
			t.Fatal("expected '%v', got '%v'", expect, res)
		}
	}
}

func BenchmarkSetRawInPlace(t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		res, err := SetRawOptions(json, path, rawValue, optsInPlace)
		if err != nil {
			t.Fatal(err)
		}
		if res != expect {
			t.Fatal("expected '%v', got '%v'", expect, res)
		}
	}
}

func BenchmarkSetBytesOptimistic(t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		res, err := SetBytesOptions(jsonBytes, path, value, opts)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Compare(res, expectBytes) != 0 {
			t.Fatal("expected '%v', got '%v'", string(expectBytes), string(res))
		}
	}
}

func BenchmarkSetBytesInPlace(t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		copy(jsonBytes2, jsonBytes)
		res, err := SetBytesOptions(jsonBytes2, path, value, optsInPlace)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Compare(res, expectBytes) != 0 {
			t.Fatal("expected '%v', got '%v'", string(expectBytes), string(res))
		}
	}
}

func BenchmarkSetRawBytesOptimistic(t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		res, err := SetRawBytesOptions(jsonBytes, path, rawValueBytes, opts)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Compare(res, expectBytes) != 0 {
			t.Fatal("expected '%v', got '%v'", string(expectBytes), string(res))
		}
	}
}

func BenchmarkSetRawBytesInPlace(t *testing.B) {
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		copy(jsonBytes2, jsonBytes)
		res, err := SetRawBytesOptions(jsonBytes2, path, rawValueBytes, optsInPlace)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Compare(res, expectBytes) != 0 {
			t.Fatal("expected '%v', got '%v'", string(expectBytes), string(res))
		}
	}
}
