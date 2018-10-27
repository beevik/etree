// Copyright 2015 Brett Vickers.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etree

import (
	"encoding/xml"
	"io"
	"testing"
)

func checkEq(t *testing.T, got, want string) {
	if got != want {
		t.Errorf("etree: unexpected result.\nGot:\n%s\nWanted:\n%s\n", got, want)
	}
}

func TestCharAndTextElements(t *testing.T) {
	doc := NewDocument()
	root := doc.CreateElement("root")
	root.CreateCharData("This ")
	root.CreateTextData("is ")
	e1 := NewCharData("a ")
	e2 := NewTextData("text ")
	root.AddChild(e1)
	root.AddChild(e2)
	root.CreateCharData("Element!!")
	doc.IndentTabs()

	s, err := doc.WriteToString()
	if err != nil {
		t.Error("etree: failed to serialize document")
	}

	// Make sure the serialized XML matches expectation.
	expected := `<root>This <![CDATA[is ]]>a <![CDATA[text ]]>Element!!</root>
`
	checkEq(t, s, expected)

	// Check we can parse the output
	err = doc.ReadFromString(s)
	if err != nil {
		t.Fatal("etree: incorrect ReadFromString result")
	}
	if doc.Root().Text() != "This is a text Element!!" {
		// The Golang XML decoder merges all the Text data into a single text
		t.Error("etree: invalid structure")
	}
}

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
	review := book.CreateElement("review")
	review.CreateTextData("<<< Will be replaced")
	review.SetText(">>> Excellent book")
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
		<review><![CDATA[>>> Excellent book]]></review>
	</book>
</store>
`
	checkEq(t, s, expected)

	// Test the structure of the XML
	if doc.Root() != store {
		t.Error("etree: root mismatch")
	}
	if len(store.ChildElements()) != 1 || len(store.Child) != 7 {
		t.Error("etree: incorrect tree structure")
	}
	if len(book.ChildElements()) != 3 || len(book.Attr) != 1 || len(book.Child) != 7 {
		t.Error("etree: incorrect tree structure")
	}
	if len(title.ChildElements()) != 0 || len(title.Child) != 1 || len(title.Attr) != 0 {
		t.Error("etree: incorrect tree structure")
	}
	if len(author.ChildElements()) != 0 || len(author.Child) != 1 || len(author.Attr) != 0 {
		t.Error("etree: incorrect tree structure")
	}
	if len(review.ChildElements()) != 0 || len(review.Child) != 1 || len(review.Attr) != 0 {
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
	element = book.SelectElement("review")
	if element != review || element.Text() != ">>> Excellent book" || len(element.Attr) != 0 {
		t.Error("etree: incorrect SelectElement result")
	}
}

func TestDocumentRead_NonUTF8Encodings(t *testing.T) {
	s := `<?xml version="1.0" encoding="ISO-8859-1"?>
	<store>
	<book lang="en">
		<title>Great Expectations</title>
		<author>Charles Dickens</author>
	</book>
</store>`

	doc := NewDocument()
	doc.ReadSettings.CharsetReader = func(label string, input io.Reader) (io.Reader, error) {
		return input, nil
	}
	err := doc.ReadFromString(s)
	if err != nil {
		t.Fatal("etree: incorrect ReadFromString result")
	}
}

func TestDocumentRead_Permissive(t *testing.T) {
	s := "<select disabled></select>"

	doc := NewDocument()
	err := doc.ReadFromString(s)
	if err == nil {
		t.Fatal("etree: incorrect ReadFromString result")
	}

	doc.ReadSettings.Permissive = true
	err = doc.ReadFromString(s)
	if err != nil {
		t.Fatal("etree: incorrect ReadFromString result")
	}
}

func TestDocumentRead_HTMLEntities(t *testing.T) {
	s := `<store>
	<book lang="en">
		<title>&rarr;&nbsp;Great Expectations</title>
		<author>Charles Dickens</author>
	</book>
</store>`

	doc := NewDocument()
	err := doc.ReadFromString(s)
	if err == nil {
		t.Fatal("etree: incorrect ReadFromString result")
	}

	doc.ReadSettings.Entity = xml.HTMLEntity
	err = doc.ReadFromString(s)
	if err != nil {
		t.Fatal("etree: incorrect ReadFromString result")
	}
}

func TestEscapeCodes(t *testing.T) {
	cases := []struct {
		input         string
		normal        string
		attrCanonical string
		textCanonical string
	}{
		{
			"&<>'\"\t\n\r",
			"<e a=\"&amp;&lt;&gt;&apos;&quot;\t\n\r\">&amp;&lt;&gt;&apos;&quot;\t\n\r</e>",
			"<e a=\"&amp;&lt;>'&quot;&#x9;&#xA;&#xD;\">&amp;&lt;&gt;&apos;&quot;\t\n\r</e>",
			"<e a=\"&amp;&lt;&gt;&apos;&quot;\t\n\r\">&amp;&lt;&gt;'\"\t\n&#xD;</e>",
		},
		{
			"\x00\x1f\x08\x09\x0a\x0d",
			"<e a=\"���\t\n\r\">���\t\n\r</e>",
			"<e a=\"���&#x9;&#xA;&#xD;\">���\t\n\r</e>",
			"<e a=\"���\t\n\r\">���\t\n&#xD;</e>",
		},
	}
	for _, c := range cases {
		doc := NewDocument()

		e := doc.CreateElement("e")
		e.SetText(c.input)
		e.CreateAttr("a", c.input)

		doc.WriteSettings.CanonicalText = false
		doc.WriteSettings.CanonicalAttrVal = false
		s, err := doc.WriteToString()
		if err != nil {
			t.Error("etree: Escape test produced inocrrect result.")
		}
		checkEq(t, s, c.normal)

		doc.WriteSettings.CanonicalText = false
		doc.WriteSettings.CanonicalAttrVal = true
		s, err = doc.WriteToString()
		if err != nil {
			t.Error("etree: Escape test produced inocrrect result.")
		}
		checkEq(t, s, c.attrCanonical)

		doc.WriteSettings.CanonicalText = true
		doc.WriteSettings.CanonicalAttrVal = false
		s, err = doc.WriteToString()
		if err != nil {
			t.Error("etree: Escape test produced inocrrect result.")
		}
		checkEq(t, s, c.textCanonical)
	}
}

func TestCanonical(t *testing.T) {
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
	jon.SetText("\r<'\">&\u0004\u0005\u001f�")

	sally := people.CreateElement("Person")
	sally.CreateAttr("name", "Sally")
	sally.CreateAttr("escape", "\r\n\t<'\">&")

	doc.Indent(2)
	s, err := doc.WriteToString()
	if err != nil {
		t.Error("etree: WriteSettings WriteTo produced incorrect result.")
	}

	expected := BOM + `<?xml-stylesheet type="text/xsl" href="style.xsl"?>
<People>
  <!--These are all known people-->
  <Person name="Jon O'Reilly">&#xD;&lt;'"&gt;&amp;����</Person>
  <Person name="Sally" escape="&#xD;&#xA;&#x9;&lt;'&quot;>&amp;"></Person>
</People>
`
	checkEq(t, s, expected)
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
		t.Fatal("etree: incorrect ReadFromString result")
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

func TestGetPath(t *testing.T) {
	testdoc := `<a>
 <b1>
  <c1>
   <d1/>
   <d1a/>
  </c1>
 </b1>
 <b2>
  <c2>
   <d2/>
  </c2>
 </b2>
</a>`

	doc := NewDocument()
	err := doc.ReadFromString(testdoc)
	if err != nil {
		t.Fatalf("etree ReadFromString: %v\n", err)
	}

	cases := []struct {
		from    string
		to      string
		relpath string
		topath  string
	}{
		{"a", ".", "..", "/"},
		{".", "a", "./a", "/a"},
		{"a/b1/c1/d1", ".", "../../../..", "/"},
		{".", "a/b1/c1/d1", "./a/b1/c1/d1", "/a/b1/c1/d1"},
		{"a", "a", ".", "/a"},
		{"a/b1", "a/b1/c1", "./c1", "/a/b1/c1"},
		{"a/b1/c1", "a/b1", "..", "/a/b1"},
		{"a/b1/c1", "a/b1/c1", ".", "/a/b1/c1"},
		{"a", "a/b1", "./b1", "/a/b1"},
		{"a/b1", "a", "..", "/a"},
		{"a", "a/b1/c1", "./b1/c1", "/a/b1/c1"},
		{"a/b1/c1", "a", "../..", "/a"},
		{"a/b1/c1/d1", "a", "../../..", "/a"},
		{"a", "a/b1/c1/d1", "./b1/c1/d1", "/a/b1/c1/d1"},
		{"a/b1", "a/b2", "../b2", "/a/b2"},
		{"a/b2", "a/b1", "../b1", "/a/b1"},
		{"a/b1/c1/d1", "a/b2/c2/d2", "../../../b2/c2/d2", "/a/b2/c2/d2"},
		{"a/b2/c2/d2", "a/b1/c1/d1", "../../../b1/c1/d1", "/a/b1/c1/d1"},
		{"a/b1/c1/d1", "a/b1/c1/d1a", "../d1a", "/a/b1/c1/d1a"},
	}

	for _, c := range cases {
		fe := doc.FindElement(c.from)
		te := doc.FindElement(c.to)

		rp := te.GetRelativePath(fe)
		if rp != c.relpath {
			t.Errorf("GetRelativePath from '%s' to '%s'. Expected '%s', got '%s'.\n", c.from, c.to, c.relpath, rp)
		}

		p := te.GetPath()
		if p != c.topath {
			t.Errorf("GetPath for '%s'. Expected '%s', got '%s'.\n", c.to, c.topath, p)
		}
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
		t.Fatal("etree ReadFromString: " + err.Error())
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
	checkEq(t, s1, expected1)

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
	checkEq(t, s2, expected2)

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
	checkEq(t, s3, expected3)

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
	checkEq(t, s4, expected4)
}

func TestCdata(t *testing.T) {
	var tests = []struct {
		in, out string
	}{
		{`<tag>1234567</tag>`, "1234567"},
		{`<tag><![CDATA[1234567]]></tag>`, "1234567"},
		{`<tag>1<![CDATA[2]]>3<![CDATA[4]]>5<![CDATA[6]]>7</tag>`, "1234567"},
		{`<tag>1<![CDATA[2]]>3<inner>4</inner>5<![CDATA[6]]>7</tag>`, "123"},
		{`<tag>1<inner>4</inner>5<![CDATA[6]]>7</tag>`, "1"},
		{`<tag><![CDATA[1]]><inner>4</inner>5<![CDATA[6]]>7</tag>`, "1"},
	}

	for _, test := range tests {
		doc := NewDocument()
		err := doc.ReadFromString(test.in)
		if err != nil {
			t.Fatal("etree ReadFromString: " + err.Error())
		}

		tag := doc.FindElement("tag")
		if tag.Text() != test.out {
			t.Fatalf("etree invalid cdata. Expected: %v. Got: %v\n", test.out, tag.Text())
		}
	}
}

func TestAddChild(t *testing.T) {
	testdoc := `<book lang="en">
  <t:title>Great Expectations</t:title>
  <author>Charles Dickens</author>
</book>
`
	doc1 := NewDocument()
	err := doc1.ReadFromString(testdoc)
	if err != nil {
		t.Fatal("etree ReadFromString: " + err.Error())
	}

	doc2 := NewDocument()
	root := doc2.CreateElement("root")

	for _, e := range doc1.FindElements("//book/*") {
		root.AddChild(e)
	}

	expected1 := `<book lang="en"/>
`
	doc1.Indent(2)
	s1, _ := doc1.WriteToString()
	checkEq(t, s1, expected1)

	expected2 := `<root>
  <t:title>Great Expectations</t:title>
  <author>Charles Dickens</author>
</root>
`
	doc2.Indent(2)
	s2, _ := doc2.WriteToString()
	checkEq(t, s2, expected2)
}

func TestSetRoot(t *testing.T) {
	testdoc := `<?test a="wow"?>
<book>
  <title>Great Expectations</title>
  <author>Charles Dickens</author>
</book>
`
	doc := NewDocument()
	err := doc.ReadFromString(testdoc)
	if err != nil {
		t.Fatal("etree ReadFromString: " + err.Error())
	}

	origroot := doc.Root()
	if origroot.Parent() != &doc.Element {
		t.Error("Root incorrect")
	}

	newroot := NewElement("root")
	doc.SetRoot(newroot)

	if doc.Root() != newroot {
		t.Error("doc.Root() != newroot")
	}
	if origroot.Parent() != nil {
		t.Error("origroot.Parent() != nil")
	}

	expected1 := `<?test a="wow"?>
<root/>
`
	doc.Indent(2)
	s1, _ := doc.WriteToString()
	checkEq(t, s1, expected1)

	doc.SetRoot(origroot)
	doc.Indent(2)
	expected2 := testdoc
	s2, _ := doc.WriteToString()
	checkEq(t, s2, expected2)

	doc2 := NewDocument()
	doc2.CreateProcInst("test", `a="wow"`)
	doc2.SetRoot(NewElement("root"))
	doc2.Indent(2)
	expected3 := expected1
	s3, _ := doc2.WriteToString()
	checkEq(t, s3, expected3)

	doc2.SetRoot(doc.Root())
	doc2.Indent(2)
	expected4 := testdoc
	s4, _ := doc2.WriteToString()
	checkEq(t, s4, expected4)

	expected5 := `<?test a="wow"?>
`
	doc.Indent(2)
	s5, _ := doc.WriteToString()
	checkEq(t, s5, expected5)
}

func TestSortAttrs(t *testing.T) {
	testdoc := `<el foo='5' Foo='2' aaa='4' สวัสดี='7' AAA='1' a01='3' z='6' a:ZZZ='9' a:AAA='8'/>`
	doc := NewDocument()
	err := doc.ReadFromString(testdoc)
	if err != nil {
		t.Fatal("etree ReadFromString: " + err.Error())
	}

	doc.Root().SortAttrs()
	doc.Indent(2)
	out, _ := doc.WriteToString()
	checkEq(t, out, `<el AAA="1" Foo="2" a01="3" aaa="4" foo="5" z="6" สวัสดี="7" a:AAA="8" a:ZZZ="9"/>`+"\n")
}

func TestCharsetReaderEncoding(t *testing.T) {
	cases := []string{
		`<?xml version="1.0" encoding="ISO-8859-1"?><foo></foo>`,
		`<?xml version="1.0" encoding="UTF-8"?><foo></foo>`,
		`<?xml version="1.0" encoding="US-ASCII"?><foo></foo>`,
	}

	for _, c := range cases {
		doc := NewDocument()
		if err := doc.ReadFromBytes([]byte(c)); err != nil {
			t.Error(err)
		}
	}
}
