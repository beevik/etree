etree
=====

The etree package is a lightweight, pure go package that expresses XML in
the form of an element tree.  Its design was inspired by the Python
[ElementTree](http://docs.python.org/2/library/xml.etree.elementtree.html)
module. Some of the package's features include:

* Represents XML documents as trees of elements for easy traversal.
* Imports, serializes, modifies or creates XML documents from scratch.
* Writes and reads XML to/from files, byte slices, strings and io interfaces.
* Performs simple or complex searches with lightweight XPath-like query APIs.
* Auto-indents XML using spaces or tabs for better readability.
* Implemented in pure go; depends only on standard go libraries.
* Built on top of the go [encoding/xml](http://golang.org/pkg/encoding/xml) package.

See http://godoc.org/github.com/beevik/etree for the godoc-formatted API
documentation.

###Example: Creating an XML document

The following example creates an XML document from scratch using the etree
package and outputs its indented contents to stdout.
```go
doc := etree.NewDocument()
doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
doc.CreateProcInst("xml-stylesheet", `type="text/xsl" href="style.xsl"`)

people := doc.CreateElement("People")
people.CreateComment("These are all known people")

jon := people.CreateElement("Person")
jon.CreateAttr("name", "Jon")

sally := people.CreateElement("Person")
sally.CreateAttr("name", "Sally")

doc.Indent(2)
doc.WriteTo(os.Stdout)
```

Output:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type="text/xsl" href="style.xsl"?>
<People>
  <!--These are all known people-->
  <Person name="Jon"/>
  <Person name="Sally"/>
</People>
```

###Document used by remaining examples

All remaining examples will use the following `bookstore.xml` document as
their source.
```xml
<bookstore xmlns:p="urn:schemas-books-com:prices">

  <book category="COOKING">
    <title lang="en">Everyday Italian</title>
    <author>Giada De Laurentiis</author>
    <year>2005</year>
    <p:price>30.00</p:price>
  </book>

  <book category="CHILDREN">
    <title lang="en">Harry Potter</title>
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

  <book category="WEB">
    <title lang="en">Learning XML</title>
    <author>Erik T. Ray</author>
    <year>2003</year>
    <p:price>39.95</p:price>
  </book>

</bookstore>
```

###Example: Reading an XML file

This example loads XML from a file called `bookstore.xml`.
```go
doc := etree.NewDocument()
if err := doc.ReadFromFile("bookstore.xml"); err != nil {
    panic(err)
}
```

###Example: Processing elements and attributes

This example illustrates some ways to access elements and attributes
using simple etree queries.
```go
root := doc.SelectElement("bookstore")
fmt.Println("ROOT element:", root.Tag)

for _, book := range root.SelectElements("book") {
    fmt.Println("CHILD element:", book.Tag)
    if title := book.SelectElement("title"); title != nil {
        lang := title.SelectAttrValue("lang", "unknown")
        fmt.Printf("  TITLE: %s (%s)\n", title.Text(), lang)
    }
    for _, attr := range book.Attr {
        fmt.Printf("  ATTR: %s=%s\n", attr.Key, attr.Value)
    }
}
```
Output:
```
ROOT element: bookstore
CHILD element: book
  TITLE: Everyday Italian (en)
  ATTR: category=COOKING
CHILD element: book
  TITLE: Harry Potter (en)
  ATTR: category=CHILDREN
CHILD element: book
  TITLE: XQuery Kick Start (en)
  ATTR: category=WEB
CHILD element: book
  TITLE: Learning XML (en)
  ATTR: category=WEB
```

###Example: Path queries

In this example, we select all book titles that fall into the category
of 'WEB'.  The book elements are searched for recursively and may
appear at any level of the tree.
```go
for _, t := range doc.FindElements("//book[@category='WEB']/title") {
    fmt.Println("Title:", t.Text())
}
```

Output:
```
Title: XQuery Kick Start
Title: Learning XML
```

This example finds the first book element under the bookstore element
and outputs the tag and text of all of its child elements.
```go
for _, e := range doc.FindElements("./bookstore/book[1]/*") {
    fmt.Printf("%s: %s\n", e.Tag, e.Text())
}
```

Output:
```
title: Everyday Italian
author: Giada De Laurentiis
year: 2005
price: 30.00
```

This example finds all books with a price of 49.99 and outputs their titles.
Note that this example uses the FindElementsPath function, which takes as an
argument a pre-compiled path object.  Use this API when you plan to search
using the same path more than once.
```go
path := etree.MustCompilePath("./bookstore/book[p:price='49.99']/title")
for _, e := range doc.FindElementsPath(path) {
    fmt.Println(e.Text())
}
```

Output:
```
XQuery Kick Start
```


###Contributing

This project accepts contributions. Just fork the repo and submit a pull request!

