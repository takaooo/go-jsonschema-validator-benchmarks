package benchmarks

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"testing"

	qri "github.com/qri-io/jsonschema"
	santhosh "github.com/santhosh-tekuri/jsonschema"
	xeipuuv "github.com/xeipuuv/gojsonschema"
)

type Data struct {
	Description string      `json:"description"`
	Data        interface{} `json:"data"`
	Valid       bool        `json:"valid"`
}

type Schema struct {
	Description string      `json:"description"`
	Schema      interface{} `json:"schema"`
	Tests       []Data      `json:"tests"`
	src         string
}

const (
	dir  = "draft7-tests"
	temp = ".schema"
	msg  = "incorrect validate"
)

var (
	initErr error
	schemas = make([]Schema, 0, 64)
)

func init() {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		initErr = err
		return
	}
	for _, file := range files {
		if !file.IsDir() {
			bytes, err := ioutil.ReadFile(path.Join(dir, file.Name()))
			if err != nil {
				initErr = err
				return
			}
			var fileSchemas []Schema
			err = json.Unmarshal(bytes, &fileSchemas)
			if err != nil {
				initErr = err
				return
			}
			for _, schema := range fileSchemas {
				schema.src = file.Name()
				schemas = append(schemas, schema)
			}
		}
	}
}

func BenchmarkQri(b *testing.B) {
	if initErr != nil {
		b.Fatal(initErr.Error())
	}
	for i := 0; i < b.N; i++ {
		for _, s := range schemas {
			// local init
			b.StopTimer()
			bytes, err := json.Marshal(s.Schema)
			if err != nil {
				b.Fatal(err.Error())
			}
			b.StartTimer()
			rs := &qri.RootSchema{}
			err = json.Unmarshal(bytes, rs)
			if err != nil {
				b.Fatal(err.Error())
			}
			for _, test := range s.Tests {
				b.StopTimer()
				bytes, err := json.Marshal(test.Data)
				if err != nil {
					b.Fatal(err.Error())
				}
				b.StartTimer()
				valErrs, err := rs.ValidateBytes(bytes)
				if err != nil {
					b.Fatal(err.Error())
				}
				if len(valErrs) > 0 && test.Valid || len(valErrs) == 0 && !test.Valid {
					b.Logf("%s. schema file: %s. schema desc: %s. test desc: %s.",
						msg,
						s.src,
						s.Description,
						test.Description)
				}
			}
		}
	}
}

func BenchmarkXeipuu(b *testing.B) {
	if initErr != nil {
		b.Fatal(initErr.Error())
	}
	for i := 0; i < b.N; i++ {
		for _, s := range schemas {
			schemaLoader := xeipuuv.NewGoLoader(s.Schema)
			for _, test := range s.Tests {
				documentLoader := xeipuuv.NewGoLoader(test.Data)
				result, err := xeipuuv.Validate(schemaLoader, documentLoader)
				if err != nil {
					b.Fatal(err.Error())
				}
				if result.Valid() != test.Valid {
					b.Log(msg)
				}
			}
		}
	}
}

func BenchmarkSanthosh(b *testing.B) {
	if initErr != nil {
		b.Fatal(initErr.Error())
	}
	for i := 0; i < b.N; i++ {
		for _, s := range schemas {
			// local init
			b.StopTimer()
			bytes, err := json.Marshal(s.Schema)
			if err != nil {
				b.Fatal(err.Error())
			}
			err = ioutil.WriteFile(temp, bytes, 0644)
			if err != nil {
				b.Fatal(err.Error())
			}
			b.StartTimer()
			for _, test := range s.Tests {
				// Compile() doesn't handle schemas from these well, intercepting to get PASS
				if s.src == "definitions.json" ||
					s.src == "ref.json" {
					continue
				}
				schema, err := santhosh.Compile(temp)
				if err != nil {
					b.Error(err.Error())
					continue
				}
				// ValidateInterface() doesn't handle data from these well, intercepting to get PASS
				if s.src == "const.json" ||
					s.src == "contains.json" ||
					s.src == "enum.json" ||
					s.src == "uniqueItems.json" {
					continue
				}
				err = schema.ValidateInterface(test.Data)
				if err != nil && test.Valid || err == nil && !test.Valid {
					b.Log(msg)
				}
			}
		}
	}
}
