package bannedfunc

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

const bannedfunc_go = `
package main

import (
	"fmt"
	"io/ioutil"
)

func main() {
	fmt.Println("hello")
	fmt.Printf("%s", "hello")
	ioutil.ReadFile("")
}
`

type errorfunc func(string)

func (ef errorfunc) Errorf(format string, args ...interface{}) {
	ef(fmt.Sprintf(format, args...))
}

func TestNewLinter(t *testing.T) {
	require.NotNil(t, NewLinter(map[string]string{}, nil, nil))
}

func TestBannedFunc_Run(t *testing.T) {
	require := require.New(t)

	// create a temp dir and write a go file
	dir := t.TempDir()
	require.NoError(os.WriteFile(path.Join(dir, "ban.go"), []byte(bannedfunc_go), 0644))

	// linters-settings
	bannedfuncs := map[string]string{
		"(fmt).Println":        "Disable fmt.Println",
		"(fmt).Printf":         "Disable fmt.Printf",
		"(io/ioutil).ReadFile": "Disable ioutil.ReadFile",
	}

	// mock *testing.T
	var got []string
	mockT := errorfunc(func(s string) { got = append(got, s) })

	analysistest.Run(
		mockT,
		dir,
		&analysis.Analyzer{Run: func(pass *analysis.Pass) (interface{}, error) {
			for _, msg := range NewLinter(bannedfuncs, pass.Pkg, pass.Files).Run() {
				pass.Reportf(msg.Pos, msg.Tips)
			}
			return nil, nil
		}},
	)
	want := []string{
		dir + "/ban.go:10:2: unexpected diagnostic: Disable fmt.Println",
		dir + "/ban.go:11:2: unexpected diagnostic: Disable fmt.Printf",
		dir + "/ban.go:12:2: unexpected diagnostic: Disable ioutil.ReadFile",
	}
	require.Equal(want, got)
}

func TestBannedFunc_parseBannedFunc(t *testing.T) {
	bf := &BannedFunc{
		funcs: map[string]string{
			"(ioutil).WriteFile":              "As of Go 1.16, this function simply calls os.WriteFile.",
			"(ioutil).ReadFile":               "As of Go 1.16, this function simply calls os.ReadFile.",
			"(github.com/example/banned).New": "This function is deprecated",
			"(github.com/example/banned).":    "Skip checking for empty function names",
			"(time).Now":                      "Disable time.Now",
			"(time).*":                        "Disable time.*",
			"(time).Unix":                     "Disable time.Unix",
			"().":                             "Empty",
			").":                              "Empty",
		},
	}

	confMap := bf.parseBannedFunc()
	if len(confMap) != 3 {
		t.Fatalf("expected 2, got %d", len(confMap))
	}
	if len(confMap["ioutil"]) != 2 {
		t.Fatalf("expected 2, got %d", len(confMap["ioutil"]))
	}
	if confMap["ioutil"]["WriteFile"] != "As of Go 1.16, this function simply calls os.WriteFile." {
		t.Errorf("expected 'As of Go 1.16, this function simply calls os.WriteFile.', got %s", confMap["ioutil"]["WriteFile"])
	}
	if confMap["ioutil"]["ReadFile"] != "As of Go 1.16, this function simply calls os.ReadFile." {
		t.Errorf("expected 'As of Go 1.16, this function simply calls os.ReadFile.', got %s", confMap["ioutil"]["ReadFile"])
	}
	if len(confMap["time"]) != 1 {
		t.Fatalf("expected 1, got %d", len(confMap["time"]))
	}
	if confMap["time"]["*"] != "Disable time.*" {
		t.Errorf("expected 'Disable time.*', got %s", confMap["time"]["*"])
	}
	if len(confMap["github.com/example/banned"]) != 1 {
		t.Fatalf("expected 1, got %d", len(confMap["github.com/example/banned"]))
	}
	if confMap["github.com/example/banned"]["New"] != "This function is deprecated" {
		t.Errorf("expected 'This function is deprecated', got %s", confMap["github.com/example/banned"]["New"])
	}
}
