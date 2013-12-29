// Copyright 2013 Brett Vickers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etree

import (
	"testing"
)

var testXml string = `
<?xml version="1.0" encoding="UTF-8"?>
<bookstore xmlns:p="books-com:prices">

  <!Directive>

  <book category="COOKING">
    <title lang="en">Everyday Italian</title>
    <author>Giada De Laurentiis</author>
    <year>2005</year>
    <p:price>30.00</p:price>
  </book>

  <book category="CHILDREN">
    <title lang="en" sku="150">Harry Potter</title>
    <author>J K. Rowling</author>
    <year>2005</year>
    <p:price>29.99</p:price>
  </book>

  <book category="WEB">
    <title lang="en">XQuery Kick Start</title>
    <author>James McGovern</author>
    <author>Per Bothner</author>
    <author>Kurt Cagle</author>
    <author>James Linn</author>
    <author>Vaidyanathan Nagarajan</author>
    <year>2003</year>
    <p:price>49.99</p:price>
  </book>

  <!-- Final book -->
  <book category="WEB">
    <title lang="en">Learning XML</title>
    <author>Erik T. Ray</author>
    <year>2003</year>
    <p:price>39.95</p:price>
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
