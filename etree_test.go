// Copyright 2013 Brett Vickers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etree

import (
	"testing"
)

func TestDocument(t *testing.T) {

	// Create a document
	doc := NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	doc.CreateProcInst("xml-stylesheet", `type="text/xsl" href="style.xsl"`)
	store := doc.CreateElement("store")
	store.CreateAttrFull("xmlns", "t", "urn:books-com:titles")
	store.CreateDirective("Directive")
	store.CreateComment("This is a comment")
	book := store.CreateElement("book")
	book.CreateAttrFull("", "lang", "fr")
	lang := book.CreateAttr("lang", "en")
	title := book.CreateElementFull("t", "title")
	title.SetText("Nicholas Nickleby")
	title.SetText("Great Expectations")
	author := book.CreateElement("author")
	author.CreateCharData("Charles Dickens")
	doc.IndentTabs()

	// Serialize the document to a string
	s, err := doc.WriteToString()
	if err != nil {
		t.Fail()
	}

	// Make sure the serialized XML matches expectation.
	expected := `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type="text/xsl" href="style.xsl"?>
<store xmlns:t="urn:books-com:titles">
	<!Directive>
	<!--This is a comment-->
	<book lang="en">
		<t:title>Great Expectations</t:title>
		<author>Charles Dickens</author>
	</book>
</store>
`
	if expected != s {
		t.Error("etree: serialized XML doesn't match expectation.")
	}

	// Test the structure of the XML
	if len(store.ChildElements()) != 1 || len(store.Child) != 7 {
		t.Fail()
	}
	if len(book.ChildElements()) != 2 || len(book.Attr) != 1 || len(book.Child) != 5 {
		t.Fail()
	}
	if len(title.ChildElements()) != 0 || len(title.Child) != 1 || len(title.Attr) != 0 {
		t.Fail()
	}
	if len(author.ChildElements()) != 0 || len(author.Child) != 1 || len(author.Attr) != 0 {
		t.Fail()
	}
	if book.Parent != store || store.Parent != &doc.Element || doc.Parent != nil {
		t.Fail()
	}
	if title.Parent != book || author.Parent != book {
		t.Fail()
	}

	// Perform some basic queries on the document
	elements := doc.SelectElementsFull("", "store")
	if len(elements) != 1 || elements[0] != store {
		t.Fail()
	}
	element := doc.SelectElementFull("", "store")
	if element != store {
		t.Fail()
	}
	elements = store.SelectElementsFull("", "book")
	if len(elements) != 1 || elements[0] != book {
		t.Fail()
	}
	element = store.SelectElementFull("", "book")
	if element != book {
		t.Fail()
	}
	attr := book.SelectAttrFull("", "lang")
	if attr == nil || attr.Key != "lang" || attr.Value != "en" {
		t.Fail()
	}
	if book.SelectAttrValueFull("", "lang", "unknown") != "en" {
		t.Fail()
	}
	if book.SelectAttrValueFull("t", "missing", "unknown") != "unknown" {
		t.Fail()
	}
	attr = book.RemoveAttr(lang)
	if attr != lang {
		t.Fail()
	}
	attr = book.SelectAttr("lang")
	if attr != nil {
		t.Fail()
	}
	element = book.SelectElementFull("t", "title")
	if element != title || element.Text() != "Great Expectations" || len(element.Attr) != 0 {
		t.Fail()
	}
	element = book.SelectElement("title")
	if element != title {
		t.Fail()
	}
	element = book.SelectElementFull("", "title")
	if element != nil {
		t.Fail()
	}
	element = book.RemoveElement(title)
	if element != title {
		t.Fail()
	}
	element = book.SelectElement("title")
	if element != nil {
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
