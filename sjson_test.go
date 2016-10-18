package sjson

import (
	"encoding/hex"
	"math/rand"
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
)

func testRaw(t *testing.T, kind int, expect, json, path string, value interface{}) {
	var json2 string
	var err error
	switch kind {
	default:
		json2, err = Set(json, path, value)
	case setRaw:
		json2, err = SetRaw(json, path, value.(string))
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
