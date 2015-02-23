// Copyright 2015 Brett Vickers.
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
	store.CreateAttr("xmlns:t", "urn:books-com:titles")
	store.CreateDirective("Directive")
	store.CreateComment("This is a comment")
	book := store.CreateElement("book")
	book.CreateAttr("lang", "fr")
	book.CreateAttr("lang", "en")
	title := book.CreateElement("t:title")
	title.SetText("Nicholas Nickleby")
	title.SetText("Great Expectations")
	author := book.CreateElement("author")
	author.CreateCharData("Charles Dickens")
	doc.IndentTabs()

	// Serialize the document to a string
	s, err := doc.WriteToString()
	if err != nil {
		t.Error("etree: failed to serialize document")
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
		t.Error("etree: serialization incorrect")
	}

	// Test the structure of the XML
	if doc.Root() != store {
		t.Error("etree: root mismatch")
	}
	if len(store.ChildElements()) != 1 || len(store.Child) != 7 {
		t.Error("etree: incorrect tree structure")
	}
	if len(book.ChildElements()) != 2 || len(book.Attr) != 1 || len(book.Child) != 5 {
		t.Error("etree: incorrect tree structure")
	}
	if len(title.ChildElements()) != 0 || len(title.Child) != 1 || len(title.Attr) != 0 {
		t.Error("etree: incorrect tree structure")
	}
	if len(author.ChildElements()) != 0 || len(author.Child) != 1 || len(author.Attr) != 0 {
		t.Error("etree: incorrect tree structure")
	}
	if book.Parent != store || store.Parent != &doc.Element || doc.Parent != nil {
		t.Error("etree: incorrect tree structure")
	}
	if title.Parent != book || author.Parent != book {
		t.Error("etree: incorrect tree structure")
	}

	// Perform some basic queries on the document
	elements := doc.SelectElements("store")
	if len(elements) != 1 || elements[0] != store {
		t.Error("etree: incorrect SelectElements result")
	}
	element := doc.SelectElement("store")
	if element != store {
		t.Error("etree: incorrect SelectElement result")
	}
	elements = store.SelectElements("book")
	if len(elements) != 1 || elements[0] != book {
		t.Error("etree: incorrect SelectElements result")
	}
	element = store.SelectElement("book")
	if element != book {
		t.Error("etree: incorrect SelectElement result")
	}
	attr := book.SelectAttr("lang")
	if attr == nil || attr.Key != "lang" || attr.Value != "en" {
		t.Error("etree: incorrect SelectAttr result")
	}
	if book.SelectAttrValue("lang", "unknown") != "en" {
		t.Error("etree: incorrect SelectAttrValue result")
	}
	if book.SelectAttrValue("t:missing", "unknown") != "unknown" {
		t.Error("etree: incorrect SelectAttrValue result")
	}
	attr = book.RemoveAttr("lang")
	if attr.Value != "en" {
		t.Error("etree: incorrect RemoveAttr result")
	}
	book.CreateAttr("lang", "de")
	attr = book.RemoveAttr("lang")
	if attr.Value != "de" {
		t.Error("etree: incorrect RemoveAttr result")
	}
	element = book.SelectElement("t:title")
	if element != title || element.Text() != "Great Expectations" || len(element.Attr) != 0 {
		t.Error("etree: incorrect SelectElement result")
	}
	element = book.SelectElement("title")
	if element != title {
		t.Error("etree: incorrect SelectElement result")
	}
	element = book.SelectElement("p:title")
	if element != nil {
		t.Error("etree: incorrect SelectElement result")
	}
	element = book.RemoveElement(title)
	if element != title {
		t.Error("etree: incorrect RemoveElement result")
	}
	element = book.SelectElement("title")
	if element != nil {
		t.Error("etree: incorrect SelectElement result")
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
		t.Error("etree: incorrect ReadFromString result")
	}

	s1, err := doc1.WriteToString()
	if err != nil {
		t.Error("etree: incorrect WriteToString result")
	}

	doc2 := doc1.Copy()
	s2, err := doc2.WriteToString()
	if err != nil {
		t.Error("etree: incorrect Copy result")
	}

	if s1 != s2 {
		t.Error("etree: mismatched Copy result")
	}

	e1 := doc1.FindElement("./store/book/title")
	e2 := doc2.FindElement("./store/book/title")
	if e1 == nil || e2 == nil {
		t.Error("etree: incorrect FindElement result")
	}
	if e1 == e2 {
		t.Error("etree: incorrect FindElement result")
	}

	e1.Parent.RemoveElement(e1)
	s1, _ = doc1.WriteToString()
	s2, _ = doc2.WriteToString()
	if s1 == s2 {
		t.Error("etree: incorrect result after RemoveElement")
	}
}
