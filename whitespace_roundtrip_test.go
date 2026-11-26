package etree

import "testing"

func TestAttributeWhitespaceRoundTrip(t *testing.T) {
	doc := NewDocument()
	if err := doc.ReadFromString("<a x=\"1&#10;2\" y=\"&#xD;\" z=\"&#9;\"/>"); err != nil {
		t.Fatal(err)
	}
	s, err := doc.WriteToString()
	if err != nil {
		t.Fatal(err)
	}
	want := "<a x=\"1&#xA;2\" y=\"&#xD;\" z=\"&#x9;\"/>"
	if s != want {
		t.Fatalf("got %q, want %q", s, want)
	}
	// A conforming parser normalizes literal attribute whitespace to spaces,
	// so the serialized character references must survive a read-back.
	roundTrip := NewDocument()
	if err := roundTrip.ReadFromString(s); err != nil {
		t.Fatal(err)
	}
	e := roundTrip.Root()
	cases := []struct{ key, value string }{
		{"x", "1\n2"}, {"y", "\r"}, {"z", "\t"},
	}
	for _, kv := range cases {
		key, value := kv.key, kv.value
		attr := e.SelectAttr(key)
		if attr == nil || attr.Value != value {
			t.Errorf("attribute %q: got %v, want %q", key, attr, value)
		}
	}
}
