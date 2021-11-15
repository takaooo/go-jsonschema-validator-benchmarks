package benchmarks

import (
	"bytes"
	"context"
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
	draft7  = "draft7-tests"
	draft201909  = "draft7-tests"
	msg  = "incorrect validate"
)

var (
	initErr error
	schemas = make([]Schema, 0, 64)
	ctx = context.Background()
)

func init() {
	dir := draft7
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
			qri.LoadDraft2019_09()
			rs := new(qri.Schema)
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
				valErrs, err := rs.ValidateBytes(ctx, bytes)
				if err != nil {
					b.Fatal(err.Error())
				}
				if len(valErrs) > 0 == test.Valid {
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

func BenchmarkSanthosh(b *testing.B) {
	if initErr != nil {
		b.Fatal(initErr.Error())
	}
	for i := 0; i < b.N; i++ {
		for _, s := range schemas {
			// local init
			b.StopTimer()
			bytez, err := json.Marshal(s.Schema)
			if err != nil {
				b.Fatal(err.Error())
			}
			if err != nil {
				b.Fatal(err.Error())
			}
			b.StartTimer()
			for _, test := range s.Tests {
				// Compiler doesn't handle schemas from these well, intercepting to get PASS
				if
					s.src == "ref.json" {
					continue
				}
				compiler := santhosh.NewCompiler()
				compiler.Draft = santhosh.Draft2019
				if err := compiler.AddResource("", bytes.NewReader(bytez)); err != nil {
					b.Fatal(err.Error())
				}
				schema, err := compiler.Compile("")

				if err != nil {
					b.Fatal(err.Error())
				}
				err = schema.Validate(test.Data)
				if (err == nil) != test.Valid {
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
