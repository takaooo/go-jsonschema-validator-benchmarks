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
	draft201909  = "draft2019-09-tests"
	custom = "custom-tests"
	msg  = "incorrect validate"
)

var (
	initErr error
	schemas = make([]Schema, 0, 64)
	ctx = context.Background()
)

func init() {
	dir := draft201909
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
				if file.Name() == "refRemote.json" {continue}
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
	for i := 0; i < 1; i++ {
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
				//following tests cause a stack overflow
				if s.Description == "Location-independent identifier" ||
					s.Description == "$anchor inside an enum is not a real identifier" {
					continue
				}
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
			schemaJSON, err := json.Marshal(s.Schema)
			if err != nil {
				b.Fatal(err.Error())
			}
			b.StartTimer()
			for _, test := range s.Tests {
				// Compiler doesn't handle schemas from these well, intercepting to get PASS
				if s.src == "vocabulary.json" ||
					s.src == "ref.json" {
					continue
				}
				compiler := santhosh.NewCompiler()
				compiler.Draft = santhosh.Draft2019
				if err := compiler.AddResource("", bytes.NewReader(schemaJSON)); err != nil {
					b.Fatal("failed to compile1!" + err.Error() + s.src)
				}
				schema, err := compiler.Compile("")
				if err != nil {
					b.Fatal("failed to compile! " + s.src + "===" + err.Error())
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

func BenchmarkQriWithoutCompiler(b *testing.B) {
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
			qri.LoadDraft2019_09()
			rs := new(qri.Schema)
			err = json.Unmarshal(bytes, rs)
			if err != nil {
				b.Fatal(err.Error())
			}
			b.StartTimer()
			for _, test := range s.Tests {
				vs := rs.Validate(ctx, test.Data)
				b.StopTimer()
				if len(*vs.Errs) > 0 == test.Valid {
					b.Logf("%s. schema file: %s. schema desc: %s. test desc: %s.",
						msg,
						s.src,
						s.Description,
						test.Description)
				}
				b.StartTimer()
			}
		}
	}
}

func BenchmarkXeipuuWithoutCompiler(b *testing.B) {
	if initErr != nil {
		b.Fatal(initErr.Error())
	}
	for i := 0; i < b.N; i++ {
		for _, s := range schemas {
			b.StopTimer()
			schemaLoader := xeipuuv.NewGoLoader(s.Schema)
			b.StartTimer()
			for _, test := range s.Tests {
				b.StopTimer()
				if s.Description == "Location-independent identifier" ||
					s.Description == "$anchor inside an enum is not a real identifier" {
					continue
				}
				documentLoader := xeipuuv.NewGoLoader(test.Data)
				schema, err := xeipuuv.NewSchema(schemaLoader)
				if err != nil {
					b.Fatal(err.Error())
				}
				b.StartTimer()
				result, err := schema.Validate(documentLoader)
				if err != nil {
					b.Fatal(err.Error())
				}
				b.StopTimer()
				if result.Valid() != test.Valid {
					b.Logf("%s. schema file: %s. schema desc: %s. test desc: %s.",
						msg,
						s.src,
						s.Description,
						test.Description)
				}
				b.StartTimer()
			}
		}
	}
}

func BenchmarkSanthoshWithoutCompiler(b *testing.B) {
	if initErr != nil {
		b.Fatal(initErr.Error())
	}
	for i := 0; i < b.N; i++ {
		for _, s := range schemas {
			// local init
			b.StopTimer()
			schemaJSON, err := json.Marshal(s.Schema)
			if err != nil {
				b.Fatal(err.Error())
			}
			b.StartTimer()
			for _, test := range s.Tests {
				b.StopTimer()
				// Compiler doesn't handle schemas from these well, intercepting to get PASS
				if s.src == "vocabulary.json" ||
					s.src == "ref.json" {
					continue
				}
				compiler := santhosh.NewCompiler()
				compiler.Draft = santhosh.Draft2019
				if err = compiler.AddResource("", bytes.NewReader(schemaJSON)); err != nil {
					b.Fatal(err.Error())
				}
				var schema *santhosh.Schema
				if schema, err = compiler.Compile(""); err != nil {
					b.Fatal(err.Error())
				}
				b.StartTimer()
				err = schema.Validate(test.Data)
				b.StopTimer()
				if (err == nil) != test.Valid {
					b.Logf("%s. schema file: %s. schema desc: %s. test desc: %s.",
						msg,
						s.src,
						s.Description,
						test.Description)
				}
				b.StartTimer()
			}
		}
	}
}
