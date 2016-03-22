// Copyright 2015 Brett Vickers.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etree

import "testing"

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
		t.Errorf("etree: serialization incorrect\ngot:\n%s\nwanted:\n%s\n",
			s, expected)
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
	if book.parent != store || store.parent != &doc.Element || doc.parent != nil {
		t.Error("etree: incorrect tree structure")
	}
	if title.parent != book || author.parent != book {
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
	element = book.RemoveChild(title).(*Element)
	if element != title {
		t.Error("etree: incorrect RemoveElement result")
	}
	element = book.SelectElement("title")
	if element != nil {
		t.Error("etree: incorrect SelectElement result")
	}
}

func TestWriteSettings(t *testing.T) {
	BOM := "\xef\xbb\xbf"

	doc := NewDocument()
	doc.WriteSettings.CanonicalEndTags = true
	doc.WriteSettings.CanonicalText = true
	doc.WriteSettings.CanonicalAttrVal = true
	doc.CreateCharData(BOM)
	doc.CreateProcInst("xml-stylesheet", `type="text/xsl" href="style.xsl"`)

	people := doc.CreateElement("People")
	people.CreateComment("These are all known people")

	jon := people.CreateElement("Person")
	jon.CreateAttr("name", "Jon O'Reilly")
	jon.SetText("<'\">&")

	sally := people.CreateElement("Person")
	sally.CreateAttr("name", "Sally")
	sally.CreateAttr("escape", "<'\">&")

	doc.Indent(2)
	s, err := doc.WriteToString()
	if err != nil {
		t.Error("etree: WriteSettings WriteTo produced incorrect result.")
	}

	expected := BOM + `<?xml-stylesheet type="text/xsl" href="style.xsl"?>
<People>
  <!--These are all known people-->
  <Person name="Jon O'Reilly">&lt;'"&gt;&amp;</Person>
  <Person name="Sally" escape="&lt;'&quot;>&amp;"></Person>
</People>
`

	if s != expected {
		t.Error("etree: WriteSettings WriteTo produced unexpected result.")
		t.Error("wanted:\n" + expected)
		t.Error("got:\n" + s)
	}
}

func TestCopy(t *testing.T) {
	s := `<store>
	<book lang="en">
		<title>Great Expectations</title>
		<author>Charles Dickens</author>
	</book>
</store>`

	doc := NewDocument()
	err := doc.ReadFromString(s)
	if err != nil {
		t.Error("etree: incorrect ReadFromString result")
	}

	s1, err := doc.WriteToString()
	if err != nil {
		t.Error("etree: incorrect WriteToString result")
	}

	doc2 := doc.Copy()
	s2, err := doc2.WriteToString()
	if err != nil {
		t.Error("etree: incorrect Copy result")
	}

	if s1 != s2 {
		t.Error("etree: mismatched Copy result")
		t.Error("wanted:\n" + s1)
		t.Error("got:\n" + s2)
	}

	e1 := doc.FindElement("./store/book/title")
	e2 := doc2.FindElement("./store/book/title")
	if e1 == nil || e2 == nil {
		t.Error("etree: incorrect FindElement result")
	}
	if e1 == e2 {
		t.Error("etree: incorrect FindElement result")
	}

	e1.parent.RemoveChild(e1)
	s1, _ = doc.WriteToString()
	s2, _ = doc2.WriteToString()
	if s1 == s2 {
		t.Error("etree: incorrect result after RemoveElement")
	}
}

func TestInsertChild(t *testing.T) {
	testdoc := `<book lang="en">
  <t:title>Great Expectations</t:title>
  <author>Charles Dickens</author>
</book>
`

	doc := NewDocument()
	err := doc.ReadFromString(testdoc)
	if err != nil {
		t.Error("etree ReadFromString: " + err.Error())
	}

	year := NewElement("year")
	year.SetText("1861")

	book := doc.FindElement("//book")
	book.InsertChild(book.SelectElement("t:title"), year)

	expected1 := `<book lang="en">
  <year>1861</year>
  <t:title>Great Expectations</t:title>
  <author>Charles Dickens</author>
</book>
`
	doc.Indent(2)
	s1, _ := doc.WriteToString()
	if s1 != expected1 {
		t.Errorf("etree: serialization incorrect\ngot:\n%s\nwanted:\n%s\n", s1, expected1)
	}

	book.RemoveChild(year)
	book.InsertChild(book.SelectElement("author"), year)

	expected2 := `<book lang="en">
  <t:title>Great Expectations</t:title>
  <year>1861</year>
  <author>Charles Dickens</author>
</book>
`
	doc.Indent(2)
	s2, _ := doc.WriteToString()
	if s2 != expected2 {
		t.Errorf("etree: serialization incorrect\ngot:\n%s\nwanted:\n%s\n", s2, expected2)
	}

	book.RemoveChild(year)
	book.InsertChild(book.SelectElement("UNKNOWN"), year)

	expected3 := `<book lang="en">
  <t:title>Great Expectations</t:title>
  <author>Charles Dickens</author>
  <year>1861</year>
</book>
`
	doc.Indent(2)
	s3, _ := doc.WriteToString()
	if s3 != expected3 {
		t.Errorf("etree: serialization incorrect\ngot:\n%s\nwanted:\n%s\n", s3, expected3)
	}

	book.RemoveChild(year)
	book.InsertChild(nil, year)

	expected4 := `<book lang="en">
  <t:title>Great Expectations</t:title>
  <author>Charles Dickens</author>
  <year>1861</year>
</book>
`
	doc.Indent(2)
	s4, _ := doc.WriteToString()
	if s4 != expected4 {
		t.Errorf("etree: serialization incorrect\ngot:\n%s\nwanted:\n%s\n", s4, expected4)
	}
}

func TestAddChild(t *testing.T) {
	testdoc := `<book lang="en">
  <t:title>Great Expectations</t:title>
  <author>Charles Dickens</author>
`
	doc1 := NewDocument()
	err := doc1.ReadFromString(testdoc)
	if err != nil {
		t.Error("etree ReadFromString: " + err.Error())
	}

	doc2 := NewDocument()
	root := doc2.CreateElement("root")

	for _, e := range doc1.FindElements("//book/*") {
		root.AddChild(e)
	}

	expected1 := `<book lang="en"/>
`
	expected2 := `<root>
  <t:title>Great Expectations</t:title>
  <author>Charles Dickens</author>
</root>
`
	doc1.Indent(2)
	s1, _ := doc1.WriteToString()

	if s1 != expected1 {
		t.Errorf("etree: serialization incorrect\ngot:\n%s\nwanted:\n%s\n", s1, expected1)
	}

	doc2.Indent(2)
	s2, _ := doc2.WriteToString()
	if s2 != expected2 {
		t.Errorf("etree: serialization incorrect\ngot:\n%s\nwanted:\n%s\n", s2, expected2)
	}
}
