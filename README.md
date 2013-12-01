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

Here is the output of the program:
```
    <?xml version="1.0" encoding="UTF-8"?>
    <?xml-stylesheet type="text/xsl" href="style.xsl"?>
    <People>
      <!--These are all known people-->
      <Person name="Jon"/>
      <Person name="Sally"/>
    </People>
```

###Sample document source

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

###Example: Processing elements and attributes

This example processes the bookstore XML document with some simple element tree queries:
```go
doc := etree.NewDocument()
if err := doc.ReadFromFile("test.xml"); err != nil {
    panic(err)
}

root := doc.SelectElement("bookstore")
fmt.Println("ROOT element:", root.Tag)

for _, c := range root.ChildElements() {
    fmt.Println("CHILD element:", c.Tag)
    for _, a := range c.Attr {
        fmt.Printf("  ATTR: %s=%s\n", a.Key, a.Value)
    }
}
```
Output:
```
ROOT element: bookstore
CHILD element: book
  ATTR: category=COOKING
CHILD element: book
  ATTR: category=CHILDREN
CHILD element: book
  ATTR: category=WEB
CHILD element: book
  ATTR: category=WEB
```