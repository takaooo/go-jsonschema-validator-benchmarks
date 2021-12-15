package benchmarks

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"path"
	"testing"

	qri "github.com/qri-io/jsonschema"
	santhosh "github.com/santhosh-tekuri/jsonschema/v5"
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
	draft201909 = "draft2019-09-tests"
	draft7      = "draft7-tests"
	custom      = "custom-tests"
	msg         = "incorrect validate"
)

var (
	ctx = context.Background()
)

func getSchemas(dir string) ([]Schema, error) {
	schemas := make([]Schema, 0, 64)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if !file.IsDir() {
			fileBytes, err := ioutil.ReadFile(path.Join(dir, file.Name()))
			if err != nil {
				return nil, err
			}
			var fileSchemas []Schema
			err = json.Unmarshal(fileBytes, &fileSchemas)
			if err != nil {
				return nil, err
			}

			for _, schema := range fileSchemas {
				if file.Name() == "refRemote.json" || file.Name() == "vocabulary.json" {
					continue
				}
				schema.src = file.Name()
				schemas = append(schemas, schema)
			}
		}
	}
	return schemas, nil
}

func get2019(b *testing.B) []Schema {
	b.StopTimer()
	schemas, err := getSchemas(draft201909)
	if err != nil {
		b.Fatal(err.Error())
	}
	b.StartTimer()
	return schemas
}

func getCustom(b *testing.B) []Schema {
	b.StopTimer()
	schemas, err := getSchemas(custom)
	if err != nil {
		b.Fatal(err.Error())
	}
	b.StartTimer()
	return schemas
}

func BenchmarkQri2019(b *testing.B) {
	schemas := get2019(b)
	benchmarkQri(schemas, b)
}

//func BenchmarkXeipuu2019(b *testing.B) {
//	schemas := get2019(b)
//	benchmarkXeipuu(schemas, b)
//}

func BenchmarkSanthosh2019(b *testing.B) {
	schemas := get2019(b)
	benchmarkSanthosh(schemas, b)
}

func BenchmarkQri2019NoCompiler(b *testing.B) {
	schemas := get2019(b)
	qri.LoadDraft2019_09()
	benchmarkQriWithoutCompiler(schemas, b)
}

//func BenchmarkXeipuu2019NoCompiler(b *testing.B) {
//	schemas := get2019(b)
//	benchmarkXeipuuWithoutCompiler(schemas, b)
//}

func BenchmarkSanthosh2019NoCompiler(b *testing.B) {
	schemas := get2019(b)
	benchmarkSanthoshWithoutCompiler(schemas, b)
}

func BenchmarkQriCustom(b *testing.B) {
	schemas := getCustom(b)
	qri.LoadDraft2019_09()
	benchmarkQri(schemas, b)
}

func BenchmarkXeipuuCustom(b *testing.B) {
	schemas := getCustom(b)
	benchmarkXeipuu(schemas, b)
}

func BenchmarkSanthoshCustom(b *testing.B) {
	schemas := getCustom(b)
	benchmarkSanthosh(schemas, b)
}

func BenchmarkQriCustomNoCompiler(b *testing.B) {
	schemas := getCustom(b)
	benchmarkQriWithoutCompiler(schemas, b)
}

func BenchmarkSanthoshCustomNoCompiler(b *testing.B) {
	schemas := getCustom(b)
	benchmarkSanthoshWithoutCompiler(schemas, b)
}

func benchmarkQri(schemas []Schema, b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, s := range schemas {
			schemaJSON, err := json.Marshal(s.Schema)
			if err != nil {
				b.Fatal(err.Error())
			}
			b.StartTimer()
			qri.LoadDraft2019_09()
			rs := new(qri.Schema)

			err = json.Unmarshal(schemaJSON, rs)
			if err != nil {
				b.Fatal(err.Error())
			}
			for _, test := range s.Tests {
				_ = rs.Validate(ctx, test.Data)
			}
		}
	}
}

func benchmarkXeipuu(schemas []Schema, b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, s := range schemas {
			schemaLoader := xeipuuv.NewGoLoader(s.Schema)
			for _, test := range s.Tests {
				documentLoader := xeipuuv.NewGoLoader(test.Data)
				_, err := xeipuuv.Validate(schemaLoader, documentLoader)
				if err != nil {
					b.Fatal(err.Error())
				}
			}
		}
	}
}

func benchmarkSanthosh(schemas []Schema, b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, s := range schemas {
			b.StopTimer()
			schemaJSON, err := json.Marshal(s.Schema)
			if err != nil {
				b.Fatal(err.Error())
			}
			b.StartTimer()
			for _, test := range s.Tests {
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
			}
		}
	}
}

func benchmarkQriWithoutCompiler(schemas []Schema, b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, s := range schemas {
			b.StopTimer()
			schemaJSON, err := json.Marshal(s.Schema)
			if err != nil {
				b.Fatal(err.Error())
			}
			rs := new(qri.Schema)
			if err = json.Unmarshal(schemaJSON, rs); err != nil {
				b.Fatal(err.Error())
			}
			b.StartTimer()
			for _, test := range s.Tests {
				_ = rs.Validate(ctx, test.Data)
				if err != nil {
					b.Fatal(err.Error())
				}
			}
		}
	}
}

func benchmarkSanthoshWithoutCompiler(schemas []Schema, b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, s := range schemas {
			b.StopTimer()
			schemaJSON, err := json.Marshal(s.Schema)
			if err != nil {
				b.Fatal(err.Error())
			}
			compiler := santhosh.NewCompiler()
			compiler.Draft = santhosh.Draft2019
			if err = compiler.AddResource("", bytes.NewReader(schemaJSON)); err != nil {
				b.Fatal(err.Error())
			}
			schema, err := compiler.Compile("")
			if err != nil {
				b.Fatal(err.Error())
			}
			b.StartTimer()
			for _, test := range s.Tests {
				err = schema.Validate(test.Data)
			}
		}
	}
}
