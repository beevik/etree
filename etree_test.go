// Copyright 2013 Brett Vickers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etree

import (
	"testing"
)

var testXml string = `
<?xml version="1.0" encoding="UTF-8"?>
<bookstore>

  <!Directive>

  <book category="COOKING">
    <title lang="en">Everyday Italian</title>
    <author>Giada De Laurentiis</author>
    <year>2005</year>
    <price>30.00</price>
  </book>

  <book category="CHILDREN">
    <title lang="en">Harry Potter</title>
    <author>J K. Rowling</author>
    <year>2005</year>
    <price>29.99</price>
  </book>

  <book category="WEB">
    <title lang="en">XQuery Kick Start</title>
    <author>James McGovern</author>
    <author>Per Bothner</author>
    <author>Kurt Cagle</author>
    <author>James Linn</author>
    <author>Vaidyanathan Nagarajan</author>
    <year>2003</year>
    <price>49.99</price>
  </book>

  <!-- Final book -->
  <book category="WEB">
    <title lang="en">Learning XML</title>
    <author>Erik T. Ray</author>
    <year>2003</year>
    <price>39.95</price>
  </book>

</bookstore>
`

func TestCreateDocument(t *testing.T) {
	doc := NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	doc.CreateProcInst("xml-stylesheet", `type="text/xsl" href="style.xsl"`)
	root := doc.CreateElement("bookstore")
	root.CreateDirective("Directive")
	root.CreateComment("This is a comment")
	book := root.CreateElement("book")
	book.CreateAttr("lang", "fr")
	book.CreateAttr("lang", "en")
	title := book.CreateElement("title")
	title.SetText("Nicholas Nickleby")
	title.SetText("Great Expectations")
	author := book.CreateElement("author")
	author.SetText("Charles Dickens")
	doc.Indent(4)
	s, err := doc.WriteToString()
	if err != nil {
		t.Fail()
	}
	expected := `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type="text/xsl" href="style.xsl"?>
<bookstore>
    <!Directive>
    <!--This is a comment-->
    <book lang="en">
        <title>Great Expectations</title>
        <author>Charles Dickens</author>
    </book>
</bookstore>
`
	if expected != s {
		t.Fail()
	}
}

func compareElements(a []*Element, b []*Element) bool {
	if len(a) != len(b) {
		return true
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return true
		}
	}
	return false
}

func TestPath(t *testing.T) {
	doc := NewDocument()

	if err := doc.ReadFromString(testXml); err != nil {
		t.Fail()
	}

	elements := doc.FindElements("*")
	if len(elements) != 1 || elements[0].Tag != "bookstore" {
		t.Fail()
	}

	elements = doc.FindElements("//book")
	if len(elements) != 4 {
		t.Fail()
	}
	for _, e := range elements {
		if e.Tag != "book" || len(e.Attr) != 1 || e.SelectAttrValue("category", "") == "" {
			t.Fail()
		}
	}

	if compareElements(doc.FindElements("//book//"), doc.FindElements("//book//*")) {
		t.Fail()
	}
	if compareElements(doc.FindElements(".//book"), doc.FindElements("//book")) {
		t.Fail()
	}

	elements = doc.FindElements("./bookstore/book[2]")
	if len(elements) != 1 || elements[0].Tag != "book" {
		t.Fail()
	}

	elements = doc.FindElements(".//book[@category='WEB']")
	if len(elements) != 2 {
		t.Fail()
	}
	for _, e := range elements {
		if e.Tag != "book" || e.SelectAttrValue("category", "") != "WEB" {
			t.Fail()
		}
		if e.SelectAttrValue("missing", "xyz") != "xyz" {
			t.Fail()
		}
	}

	elements = doc.FindElements("./bookstore/book/title/..")
	if len(elements) != 4 {
		t.Fail()
	}
	for _, e := range elements {
		if e.Tag != "book" {
			t.Fail()
		}
	}

	element := doc.FindElement("./bookstore/book[4]/title")
	if element.Text() != "Learning XML" {
		t.Fail()
	}

	if doc.FindElement("./bookstore/book[0]") != doc.FindElement("./bookstore/book[1]") {
		t.Fail()
	}

}

func TestPath2(t *testing.T) {
	doc := NewDocument()

	if err := doc.ReadFromString(testXml); err != nil {
		t.Fail()
	}

	defer func() {
		if e := recover(); e != nil {
			if e != errPath {
				t.Fail()
			}
		}
	}()

	doc.FindElement("/bookstore")
	t.Fail()
}
