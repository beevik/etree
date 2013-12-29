// Copyright 2013 Brett Vickers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etree

import (
	"testing"
)

type test struct {
	path   string
	result interface{}
}

var tests = []test{

	// basic queries
	{"./bookstore/book/title", []string{"Everyday Italian", "Harry Potter", "XQuery Kick Start", "Learning XML"}},
	{"./bookstore/book/author", []string{"Giada De Laurentiis", "J K. Rowling", "James McGovern", "Per Bothner", "Kurt Cagle", "James Linn", "Vaidyanathan Nagarajan", "Erik T. Ray"}},
	{"./bookstore/book/year", []string{"2005", "2005", "2003", "2003"}},
	{"./bookstore/book/p:price", []string{"30.00", "29.99", "49.99", "39.95"}},
	{"./bookstore/book/isbn", nil},

	// descendant queries
	{"//title", []string{"Everyday Italian", "Harry Potter", "XQuery Kick Start", "Learning XML"}},
	{"//book/title", []string{"Everyday Italian", "Harry Potter", "XQuery Kick Start", "Learning XML"}},
	{".//title", []string{"Everyday Italian", "Harry Potter", "XQuery Kick Start", "Learning XML"}},
	{".//bookstore//title", []string{"Everyday Italian", "Harry Potter", "XQuery Kick Start", "Learning XML"}},
	{".//book/title", []string{"Everyday Italian", "Harry Potter", "XQuery Kick Start", "Learning XML"}},
	{".//p:price/.", []string{"30.00", "29.99", "49.99", "39.95"}},
	{".//price", nil},

	// positional queries
	{"./bookstore/book[1]/title", "Everyday Italian"},
	{"./bookstore/book[4]/title", "Learning XML"},
	{"./bookstore/book[5]/title", nil},
	{"./bookstore/book[3]/author[0]", "James McGovern"},
	{"./bookstore/book[3]/author[1]", "James McGovern"},
	{"./bookstore/book[3]/author[3]/./.", "Kurt Cagle"},
	{"./bookstore/book[3]/author[6]", nil},
	{"./bookstore/book[-1]/title", "Learning XML"},
	{"./bookstore/book[-4]/title", "Everyday Italian"},
	{"./bookstore/book[-5]/title", nil},

	// text queries
	{"./bookstore/book[author='James McGovern']/title", "XQuery Kick Start"},
	{"./bookstore/book[author='Per Bothner']/title", "XQuery Kick Start"},
	{"./bookstore/book[author='Kurt Cagle']/title", "XQuery Kick Start"},
	{"./bookstore/book[author='James Linn']/title", "XQuery Kick Start"},
	{"./bookstore/book[author='Vaidyanathan Nagarajan']/title", "XQuery Kick Start"},
	{"//book[p:price='29.99']/title", "Harry Potter"},
	{"//book[price='29.99']/title", nil},

	// attribute queries
	{"./bookstore/book[@category='WEB']/title", []string{"XQuery Kick Start", "Learning XML"}},
	{"./bookstore/book[@category='COOKING']/title[@lang='en']", "Everyday Italian"},
	{"./bookstore/book/title[@lang='en'][@sku='150']", "Harry Potter"},
	{"./bookstore/book/title[@lang='fr']", nil},

	// parent queries
	{"./bookstore/book[@category='COOKING']/title/../../book[4]/title", "Learning XML"},
}

func TestGoodPaths(t *testing.T) {
	doc := NewDocument()
	err := doc.ReadFromString(testXml)
	if err != nil {
		t.Error(err)
	}
	for _, tt := range tests {
		path, err := CompilePath(tt.path)
		if err != nil {
			fail(t, tt)
			return
		}

		elements := doc.FindElementsPath(path)
		switch s := tt.result.(type) {
		case nil:
			if len(elements) != 0 {
				fail(t, tt)
			}
		case string:
			if len(elements) != 1 || elements[0].Text() != s {
				fail(t, tt)
			}
			element := doc.FindElementPath(path)
			if element == nil || element.Text() != s {
				fail(t, tt)
			}
		case []string:
			if len(elements) != len(s) {
				fail(t, tt)
				return
			}
			for i := 0; i < len(elements); i++ {
				if elements[i].Text() != s[i] {
					fail(t, tt)
					break
				}
			}
			element := doc.FindElementPath(path)
			if element == nil || element.Text() != s[0] {
				fail(t, tt)
			}
		}

	}
}

func fail(t *testing.T, tt test) {
	t.Errorf("etree: failed test '%s'\n", tt.path)
}

var badPaths = []string{
	"/bookstore",
	"./bookstore/book[]",
	"./bookstore/book[@category='WEB'",
	"./bookstore/book[@category='WEB]",
	"./bookstore/book[author]a",
}

func TestBadPaths(t *testing.T) {
	for _, s := range badPaths {
		_, err := CompilePath(s)
		if err == nil {
			t.Errorf("etree: bad path '%s' failed to cause error\n", s)
		}
	}
}
