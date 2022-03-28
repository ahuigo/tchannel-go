// Copyright (c) 2015 Uber Technologies, Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uber/tchannel-go/testutils"
)

// These tests ensure that the code generator generates valid code that can be built
// in combination with Thrift's autogenerated code.

const _tchannelPackage = "github.com/uber/tchannel-go"

var (
	_testGoPath     string
	_testGoPathOnce sync.Once
)

func TestMain(m *testing.M) {
	exitCode := m.Run()

	// If we created a fake GOPATH, we should clean it up on success.
	if _testGoPath != "" && exitCode == 0 {
		os.RemoveAll(_testGoPath)
	}

	os.Exit(exitCode)
}

func getTChannelDir(goPath string) string {
	return filepath.Join(goPath, "src", _tchannelPackage)
}

func getCurrentTChannelPath(t *testing.T) string {
	wd, err := os.Getwd()
	require.NoError(t, err, "Failed to get working directory")

	// Walk up "wd" till we find "tchannel-go".
	for filepath.Base(wd) != filepath.Base(_tchannelPackage) {
		wd = filepath.Dir(wd)
		if wd == "" {
			t.Fatalf("Failed to find tchannel-go in parents of current directory")
		}
	}

	return wd
}

func createGoPath(t *testing.T) {
	goPath, err := ioutil.TempDir("", "thrift-gen")
	require.NoError(t, err, "TempDir failed")

	// Create $GOPATH/src/github.com/uber/tchannel-go and symlink everything.
	// And then create a dummy directory for all the test output.
	tchannelDir := getTChannelDir(goPath)
	require.NoError(t, os.MkdirAll(tchannelDir, 0755), "MkDirAll failed")

	// Symlink the contents of tchannel-go into the temp directory.
	realTChannelDir := getCurrentTChannelPath(t)
	realDirContents, err := ioutil.ReadDir(realTChannelDir)
	require.NoError(t, err, "Failed to read real tchannel-go dir")

	for _, f := range realDirContents {
		realPath := filepath.Join(realTChannelDir, f.Name())
		err := os.Symlink(realPath, filepath.Join(tchannelDir, filepath.Base(f.Name())))
		require.NoError(t, err, "Failed to symlink %v", f.Name())
	}

	_testGoPath = goPath

	// None of the other tests in this package should use GOPATH, so we don't
	// restore this.
	os.Setenv("GOPATH", goPath)
}

func getOutputDir(t *testing.T) (dir, pkg string) {
	_testGoPathOnce.Do(func() { createGoPath(t) })

	// Create a random directory inside of the GOPATH in tmp
	randStr := testutils.RandString(10)
	randDir := filepath.Join(getTChannelDir(_testGoPath), randStr)
	// In case it's not empty.
	os.RemoveAll(randDir)

	return randDir, filepath.Join(_tchannelPackage, randStr)
}

func TestAllThrift(t *testing.T) {
	files, err := ioutil.ReadDir("test_files")
	require.NoError(t, err, "Cannot read test_files directory: %v", err)

	for _, f := range files {
		fname := f.Name()
		if f.IsDir() || filepath.Ext(fname) != ".thrift" {
			continue
		}

		if err := runBuildTest(t, filepath.Join("test_files", fname)); err != nil {
			t.Errorf("Thrift file %v failed: %v", fname, err)
		}
	}
}

func TestIncludeThrift(t *testing.T) {
	dirs, err := ioutil.ReadDir("test_files/include_test")
	require.NoError(t, err, "Cannot read test_files/include_test directory: %v", err)

	for _, d := range dirs {
		dname := d.Name()
		if !d.IsDir() {
			continue
		}

		thriftFile := filepath.Join(dname, path.Base(dname)+".thrift")
		if err := runBuildTest(t, filepath.Join("test_files/include_test/", thriftFile)); err != nil {
			t.Errorf("Thrift test %v failed: %v", dname, err)
		}
	}
}

func TestMultipleFiles(t *testing.T) {
	if err := runBuildTest(t, filepath.Join("test_files", "multi_test", "file1.thrift")); err != nil {
		t.Errorf("Multiple file test failed: %v", err)
	}
}

func TestExternalTemplate(t *testing.T) {
	template1 := `package {{ .Package }}

{{ range .AST.Services }}
// Service {{ .Name }} has {{ len .Methods }} methods.
{{ range .Methods }}
// func {{ .Name | goPublicName }} ({{ range .Arguments }}{{ .Type | goType }}, {{ end }}) ({{ if .ReturnType }}{{ .ReturnType | goType }}{{ end }}){{ end }}
{{ end }}
	`
	templateFile := writeTempFile(t, template1)
	defer os.Remove(templateFile)

	expected := `package service_extend

// Service S1 has 1 methods.

// func M1 ([]byte, ) ([]byte)

// Service S2 has 1 methods.

// func M2 (*S, int32, ) (*S)

// Service S3 has 1 methods.

// func M3 () ()
`

	opts := processOptions{
		InputFile:     "test_files/service_extend.thrift",
		TemplateFiles: []string{templateFile},
	}
	checks := func(dir string) error {
		dir = filepath.Join(dir, "service_extend")
		if err := checkDirectoryFiles(dir, 6); err != nil {
			return err
		}

		// Verify the contents of the extra file.
		outFile := filepath.Join(dir, defaultPackageName(templateFile)+"-service_extend.go")
		return verifyFileContents(outFile, expected)
	}
	if err := runTest(t, opts, checks); err != nil {
		t.Errorf("Failed to run test: %v", err)
	}
}

func writeTempFile(t *testing.T, contents string) string {
	tempFile, err := ioutil.TempFile("", "temp")
	require.NoError(t, err, "Failed to create temp file")
	tempFile.Close()
	require.NoError(t, ioutil.WriteFile(tempFile.Name(), []byte(contents), 0666),
		"Write temp file failed")
	return tempFile.Name()
}

func verifyFileContents(filename, expected string) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	bytesStr := string(bytes)
	if bytesStr != expected {
		return fmt.Errorf("file contents mismatch. got:\n%vexpected:\n%v", bytesStr, expected)
	}

	return nil
}

func copyFile(src, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	writeF, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer writeF.Close()

	_, err = io.Copy(writeF, f)
	return err
}

// setupDirectory creates a temporary directory.
func setupDirectory(thriftFile string) (string, error) {
	tempDir, err := ioutil.TempDir("", "thrift-gen")
	if err != nil {
		return "", err
	}

	return tempDir, nil
}

func createAdditionalTestFile(thriftFile, tempDir string) error {
	f, err := os.Open(thriftFile)
	if err != nil {
		return err
	}

	var writer io.Writer
	rdr := bufio.NewReader(f)
	for {
		line, err := rdr.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
		}

		if strings.HasPrefix(line, "//Go code:") {
			fileName := strings.TrimSpace(strings.TrimPrefix(line, "//Go code:"))
			outFile := filepath.Join(tempDir, fileName)
			f, err := os.OpenFile(outFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
			if err != nil {
				return err
			}
			defer f.Close()
			writer = f
		} else if writer != nil {
			if strings.HasPrefix(line, "//") {
				writer.Write([]byte(strings.TrimPrefix(line, "//")))
			} else {
				return nil
			}
		}
	}
}

func checkDirectoryFiles(dir string, n int) error {
	dirContents, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	if len(dirContents) < n {
		return fmt.Errorf("expected to generate at least %v files, but found: %v", n, len(dirContents))
	}

	return nil
}

func runBuildTest(t *testing.T, thriftFile string) error {
	extraChecks := func(dir string) error {
		return checkDirectoryFiles(filepath.Join(dir, defaultPackageName(thriftFile)), 4)
	}

	opts := processOptions{InputFile: thriftFile}
	return runTest(t, opts, extraChecks)
}

func runTest(t *testing.T, opts processOptions, extraChecks func(string) error) error {
	tempDir, outputPkg := getOutputDir(t)

	// Generate code from the Thrift file.
	*packagePrefix = outputPkg + "/"
	opts.GenerateThrift = true
	opts.OutputDir = tempDir
	if err := processFile(opts); err != nil {
		return fmt.Errorf("processFile(%s) in %q failed: %v", opts.InputFile, tempDir, err)
	}

	// Create any extra Go files as specified in the Thrift file.
	if err := createAdditionalTestFile(opts.InputFile, tempDir); err != nil {
		return fmt.Errorf("failed creating additional test files for %s in %q: %v", opts.InputFile, tempDir, err)
	}

	// Run go build to ensure that the generated code builds.
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = tempDir
	// NOTE: we check output, since go build ./... returns 0 status code on failure:
	// https://github.com/golang/go/issues/11407
	output, err := cmd.CombinedOutput()
	var outputLines []string
	for _, s := range strings.Split(string(output), "\n") {
		if !(strings.HasPrefix(s, "go: downloading") || strings.TrimSpace(s) == "") {
			outputLines = append(outputLines, s)
		}
	}
	fmt.Println(err, error(nil), string(output), len(outputLines), outputLines)
	if err != nil || len(outputLines) > 0 {
		return fmt.Errorf("build in %q failed.\nError: %v Output:\n%v", tempDir, err, string(output))
	}

	// Run any extra checks the user may want.
	if err := extraChecks(tempDir); err != nil {
		return err
	}

	// Only delete the temp directory on success.
	os.RemoveAll(tempDir)
	return nil
}
