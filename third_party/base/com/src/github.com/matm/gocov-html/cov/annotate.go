// Copyright (c) 2012 The Gocov Authors.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package cov

import (
	"fmt"
	"github.com/axw/gocov"
	"go/token"
	"html"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

const (
	hitPrefix  = "    "
	missPrefix = "MISS"
)

type fileInfo struct {
	file *token.File
	data string
}

type annotator struct {
	fset  *token.FileSet
	files map[string]fileInfo
}

func newAnnotator() *annotator {
	a := &annotator{}
	a.fset = token.NewFileSet()
	a.files = make(map[string]fileInfo)
	return a
}

var annotatorGbl = newAnnotator()

func annotateFunctionSource(w io.Writer, fn *gocov.Function, stmtCovSkipped *int) {
	if fn == nil {
		panic("nil function to annotate")
	}
	a := annotatorGbl
	a.printFunctionSource(w, fn, stmtCovSkipped)
}

func (a *annotator) getFile(fnFile string) (f *token.File, src string, err error) {

	if fi, ok := a.files[fnFile]; ok {
		return fi.file, fi.data, nil
	}

	info, err := os.Stat(fnFile)
	if err != nil {
		return
	}

	f = a.fset.AddFile(fnFile, a.fset.Base(), int(info.Size()))
	data, err := ioutil.ReadFile(fnFile)
	if err != nil {
		return
	}

	f.SetLinesForContent(data)
	src = string(data)

	a.files[fnFile] = fileInfo{f, src}
	return
}

func (a *annotator) printFunctionSource(w io.Writer, fn *gocov.Function, stmtCovSkipped *int) error {
	file, data, err := a.getFile(fn.File)
	if err != nil {
		return err
	}

	// bugfix: 不修改 fn.Statements，以便本函数可以重复调用
	statements := make([]*gocov.Statement, len(fn.Statements))
	copy(statements, fn.Statements)

	finit := stmtCovSkipped != nil
	linenoToSkip := -1
	lineno := file.Line(file.Pos(fn.Start))
	lines := strings.Split(data[fn.Start:fn.End], "\n")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "<div class=\"funcname\" id=\"fn_%s\">func %s</div>", fn.Name, fn.Name)
	fmt.Fprintf(w, "<div class=\"info\"><a href=\"#s_fn_%s\">Back</a><p>In <code>%s</code>:</p></div>",
		fn.Name, fn.File)
	fmt.Fprintln(w, "<table class=\"listing\">")
	for i, line := range lines {
		lineno := lineno + i
		statementFound := false
		hit := false
		for j := 0; j < len(statements); j++ {
			start := file.Line(file.Pos(statements[j].Start))
			if start == lineno {
				statementFound = true
				if !hit && statements[j].Reached > 0 {
					hit = true
				}
				statements = append(statements[:j], statements[j+1:]...)
			}
		}
		hitmiss := hitPrefix
		if statementFound && !hit {
			hitmiss = missPrefix
		}
		tr := "<tr"
		if hitmiss == missPrefix {
			tr += ` class="miss">`
			if finit {
				if strings.Index(line, "// gocov:") != -1 || lineno == linenoToSkip {
					*stmtCovSkipped++
					linenoToSkip = lineno + 1
				}
			}
		} else {
			tr += ">"
		}
		fmt.Fprintf(w, "%s<td>%d</td><td><code><pre>%s</pre></code></td></tr>", tr, lineno,
			html.EscapeString(strings.Replace(line, "\t", "        ", -1)))
	}
	fmt.Fprintln(w, "</table>")
	return nil
}
