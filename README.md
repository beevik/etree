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
    <?xml version="1.0" encoding="UTF-8"?>
    <?xml-stylesheet type="text/xsl" href="style.xsl"?>
    <People>
      <!--These are all known people-->
      <Person name="Jon"/>
      <Person name="Sally"/>
    </People>
