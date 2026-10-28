package etree

import "testing"

func TestReplacedTextTokensAreDetached(t *testing.T) {
	for _, method := range []string{"SetText", "SetCData", "SetTail"} {
		for _, replacement := range []string{"replacement", ""} {
			for _, content := range []string{"one<![CDATA[two]]>", "one<![CDATA[two]]>three"} {
				t.Run(method+"/"+replacement+"/"+content, func(t *testing.T) {
					doc := NewDocument()
					prefix := ""
					if method == "SetTail" {
						prefix = "<previous/>"
					}
					if err := doc.ReadFromString("<root>" + prefix + content + "<next/></root>"); err != nil {
						t.Fatal(err)
					}
					root := doc.Root()
					start := 0
					if method == "SetTail" {
						start = 1
					}
					text := append([]Token(nil), root.Child[start:len(root.Child)-1]...)
					next := root.SelectElement("next")
					switch method {
					case "SetText":
						root.SetText(replacement)
					case "SetCData":
						root.SetCData(replacement)
					case "SetTail":
						root.SelectElement("previous").SetTail(replacement)
					}
					removed := text
					if replacement != "" {
						if root.Child[start] != text[0] {
							t.Fatal("replacement did not retain the first text token")
						}
						if cd := text[0].(*CharData); cd.Data != replacement || cd.IsCData() != (method == "SetCData") {
							t.Fatal("retained text has the wrong data or CDATA status")
						}
						removed = text[1:]
					}
					for _, token := range removed {
						if token.Parent() != nil || token.Index() != -1 {
							t.Errorf("removed token still belongs to a parent at index %d", token.Index())
						}
						if token.(*CharData).Remove() {
							t.Error("removing an already replaced token changed the tree")
						}
						other := NewElement("other")
						other.AddChild(token)
						if token.Parent() != other || len(other.Child) != 1 {
							t.Fatal("removed token could not be attached to a new parent")
						}
					}
					if root.SelectElement("next") != next || next.Parent() != root {
						t.Fatal("using a removed text token removed the live sibling")
					}
					for i, token := range root.Child {
						if token.Parent() != root || token.Index() != i {
							t.Fatalf("live child %d has stale ownership or index", i)
						}
					}
				})
			}
		}
	}
}
