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
	if doc.Root() != store {
		t.Fail()
	}
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
	book.CreateAttrFull("", "lang", "fr")
	attr = book.RemoveAttrByKeyFull("", "lang")
	if attr.Value != "fr" {
		t.Fail()
	}
	book.CreateAttr("lang", "de")
	attr = book.RemoveAttrByKey("lang")
	if attr.Value != "de" {
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

func TestCopy(t *testing.T) {
	s := `<store>
	<book lang="en">
		<title>Great Expectations</title>
		<author>Charles Dickens</author>
	</book>
</store>`

	doc1 := NewDocument()
	err := doc1.ReadFromString(s)
	if err != nil {
		t.Fail()
	}

	s1, err := doc1.WriteToString()
	if err != nil {
		t.Fail()
	}

	doc2 := doc1.Copy()
	s2, err := doc2.WriteToString()
	if err != nil {
		t.Fail()
	}

	if s1 != s2 {
		t.Error("Copied documents don't match")
	}

	e1 := doc1.FindElement("./store/book/title")
	e2 := doc2.FindElement("./store/book/title")
	if e1 == nil || e2 == nil {
		t.Error("Failed to find element")
	}
	if e1 == e2 {
		t.Error("Copied documents contain same element")
	}

	e1.Parent.RemoveElement(e1)
	s1, _ = doc1.WriteToString()
	s2, _ = doc2.WriteToString()
	if s1 == s2 {
		t.Error("Copied and modified documents should not match")
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
