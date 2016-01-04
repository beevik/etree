// Copyright 2015 Brett Vickers.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etree_test

import (
	"os"

	"github.com/beevik/etree"
)

// Create an etree Document, add XML entities to it, and serialize it
// to stdout.
func ExampleDocument_creating() {
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	doc.CreateProcInst("xml-stylesheet", `type="text/xsl" href="style.xsl"`)

	people := doc.CreateElement("People")
	people.CreateComment("These are all known people")

	jon := people.CreateElement("Person")
	jon.CreateAttr("name", "Jon O'Reilly")

	sally := people.CreateElement("Person")
	sally.CreateAttr("name", "Sally")

	doc.Indent(2)
	doc.WriteTo(os.Stdout)
	// Output:
	// <?xml version="1.0" encoding="UTF-8"?>
	// <?xml-stylesheet type="text/xsl" href="style.xsl"?>
	// <People>
	//   <!--These are all known people-->
	//   <Person name="Jon O&apos;Reilly"/>
	//   <Person name="Sally"/>
	// </People>
}

// Create an etree Document, add XML entities to it, and serialize it
// to stdout.
func ExampleDocument_creatingWithNonDefaultWriteSettings() {
	doc := etree.NewDocument()
	doc.WriteSettings.EnableCanonicalEndTags = true
	doc.WriteSettings.EnableCanonicalText = true
	doc.WriteSettings.EnableCanonicalAttrVal = true
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
	doc.WriteTo(os.Stdout)
	// Output:
	// <?xml-stylesheet type="text/xsl" href="style.xsl"?>
	// <People>
	//   <!--These are all known people-->
	//   <Person name="Jon O'Reilly">&lt;'"&gt;&amp;</Person>
	//   <Person name="Sally" escape="&lt;'&quot;>&amp;"></Person>
	// </People>
}

func ExampleDocument_reading() {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile("document.xml"); err != nil {
		panic(err)
	}
}

func ExamplePath() {
	xml := `<bookstore><book><title>Great Expectations</title>
      <author>Charles Dickens</author></book><book><title>Ulysses</title>
      <author>James Joyce</author></book></bookstore>`

	doc := etree.NewDocument()
	doc.ReadFromString(xml)
	for _, e := range doc.FindElements(".//book[author='Charles Dickens']") {
		book := etree.CreateDocument(e)
		book.Indent(2)
		book.WriteTo(os.Stdout)
	}
	// Output:
	// <book>
	//   <title>Great Expectations</title>
	//   <author>Charles Dickens</author>
	// </book>
}
