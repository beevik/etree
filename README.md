etree
=====

A go package that builds an XML element tree and serializes it.

See http://godoc.org/github.com/beevik/etree for documentation.

###Example: Creating an XML document

The following example creates an XML document using the etree package and outputs it to stdout:
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

Output
```xml
<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type="text/xsl" href="style.xsl"?>
<People>
  <!--These are all known people-->
  <Person name="Jon"/>
  <Person name="Sally"/>
</People>
```

###Document used by remaining examples:

For the remaining examples, we will be using the following `bookstore.xml` document as the source.
```xml
<bookstore>

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

  <book category="WEB">
    <title lang="en">Learning XML</title>
    <author>Erik T. Ray</author>
    <year>2003</year>
    <price>39.95</price>
  </book>

</bookstore>
```

###Example: Reading an XML file

This example loads the bookstore sample XML from a file called `bookmark.xml`:
```go
doc := etree.NewDocument()
if err := doc.ReadFromFile("bookstore.xml"); err != nil {
    panic(err)
}
```

###Example: Processing elements and attributes

This example processes the bookstore XML document with some simple element tree queries:
```go
root := doc.SelectElement("bookstore")
fmt.Println("ROOT element:", root.Tag)

for _, book := range root.SelectElements("book") {
    fmt.Println("CHILD element:", book.Tag)
    title := book.SelectElement("title")
    if title != nil {
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

This example processes the bookstore XML document using Path queries:
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
