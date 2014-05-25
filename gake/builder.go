// Copyright 2014 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"
)

// Build uses the tool "go build" to compile the make files.
// Returns the working directory and the error, if any.
func Build(pkg *makePackage) (workDir string, err error) {
	workDir, err = ioutil.TempDir("", "gake-")
	if err != nil {
		return
	}

	// Copy all files to the temporary directory.
	for _, f := range pkg.Files {
		src, err := ioutil.ReadFile(f.Name)
		if err != nil {
			return "", err
		}
		err = ioutil.WriteFile(filepath.Join(workDir, filepath.Base(f.Name)), src, 0644)
		if err != nil {
			return "", err
		}
	}

	// Write the 'makemain.go' file.
	f, err := os.Create(filepath.Join(workDir, "makemain.go"))
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err = makemainTmpl.Execute(f, pkg); err != nil {
		return "", err
	}

	// Build

	if err = os.Chdir(workDir); err != nil {
		return "", err
	}

	dstFile := "foo"
	if runtime.GOOS == "windows" {
		dstFile += ".exe"
	}
	cmd := exec.Command("go", "build", "--tags", "gake", "-o", dstFile)
	//cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return "", err
	}

	return
}

var makemainTmpl = template.Must(template.New("main").Parse(`
package main

import (
	"regexp"

	"github.com/kless/osutil/gake/making"
)

var makes = []making.InternalMake{
{{range $_, $f := .Files}}{{range $f.MakeFuncs}}
	{"{{.Name}}", {{.Name}}},{{end}}{{end}}
}

var matchPat string
var matchRe *regexp.Regexp

func matchString(pat, str string) (result bool, err error) {
	if matchRe == nil || matchPat != pat {
		matchPat = pat
		matchRe, err = regexp.Compile(matchPat)
		if err != nil {
			return
		}
	}
	return matchRe.MatchString(str), nil
}

func main() {
	making.Main(matchString, makes)
}
`))