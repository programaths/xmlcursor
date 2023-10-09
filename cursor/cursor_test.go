package cursor

import (
	"com.programaths.xmlcursor/parser"
	"encoding/xml"
	"strings"
	"testing"
)

func TestParseXmlAndInsert(t *testing.T) {
	parse, err := parser.Parse(strings.NewReader("<a><b></b></a>"))
	if err != nil {
		t.FailNow()
	}
	cursor := NewCursor(parse)
	if cursor.index != -1 {
		t.Fail()
	}
	token := cursor.ToNextToken()

	if token.IsStart() {
		if token.Name().Local != "a" {
			t.Fail()
		}
		cursor.ToNextToken()
		cursor.ToNextToken()
		cursor.BeginElement(xml.Name{Local: "c"})
	} else {
		t.Fail()
	}

	cursor.ToStartDoc()
	xmlText := cursor.Xml()
	println(xmlText)
}

func TestParseMove(t *testing.T) {
	parse, err := parser.Parse(strings.NewReader("<a><b></b></a>"))
	if err != nil {
		t.FailNow()
	}
	cursor := NewCursor(parse)
	if cursor.index != -1 {
		t.Fail()
	}
	token := cursor.ToNextToken()

	if token.IsStart() {
		if token.Name().Local != "a" {
			t.Fail()
		}
		cursor.ToEndToken()
		cursor.BeginElement(xml.Name{Local: "c"})
		cursor.InsertElement(xml.Name{Local: "d"})
		cursor.InsertElement(xml.Name{Local: "e"})
		cursor.ToNextToken()
		cursor.BeginElement(xml.Name{Local: "f"})
	} else {
		t.Fail()
	}

	cursor.ToStartDoc()
	xmlText := cursor.Xml()
	println(xmlText)
}

func TestCursor_MoveAround(t *testing.T) {
	parse, err := parser.Parse(strings.NewReader("<a>uuu<b>test<c></c>hhh<d></d>qqq<e></e><f></f>ccc</b>ppp</a>"))
	if err != nil {
		t.FailNow()
	}
	cursor := NewCursor(parse)
	cursor.ToStartDoc()
	cursor.ToFirstChild()
	if cursor.CurrentToken().Name().Local != "b" {
		t.Fail()
	}
	cursor.ToFirstChildByName(xml.Name{
		Space: "",
		Local: "e",
	})
	if cursor.CurrentToken().Name().Local != "e" {
		t.Fail()
	}
}

func TestCursor_SetBookmark(t *testing.T) {
	parse, err := parser.Parse(strings.NewReader("<a>uuu<b>ok<c></c>hhh<!-- test --><d></d>qqq<e></e><f></f>ccc</b>ppp</a>"))
	if err != nil {
		t.FailNow()
	}
	cursor := NewCursor(parse)
	//NewCursor(parse)
	cursor.ToStartDoc()
	cursor.ToFirstChild()
	if cursor.CurrentToken().Name().Local != "b" {
		t.Fail()
	}
	cursor.ToFirstChildByName(xml.Name{
		Space: "",
		Local: "e",
	})
	if cursor.CurrentToken().Name().Local != "e" {
		t.Fail()
	}
	cursor.SetBookmark("test")
	cursor.ToStartDoc()
	if cursor.CurrentToken().Name().Local == "e" {
		t.Fail()
	}
	cursor.ToNextToken()
	cursor.InsertElementWithText(xml.Name{
		Space: "",
		Local: "test",
	}, "some text")
	cursor.ToFirstBookmark("test")
	if cursor.CurrentToken().Name().Local != "e" {
		t.Fail()
	}
	cursor.ToStartDoc()
	println(cursor.Xml())
}

func TestTokenTypes(t *testing.T) {
	parse, err := parser.Parse(strings.NewReader(`<?xml version="1.0" encoding="UTF-8" ?><root foo="42">text<!-- comment--><![CDATA[
	<test> of c data
	]]></root>`))
	if err != nil {
		t.FailNow()
	}
	cursor := NewCursor(parse)
	cursor.ToStartDoc()
	if !cursor.CurrentToken().IsProcessingInstruction() {
		t.Fail()
	}
	if !cursor.ToNextToken().IsStart() {
		t.Fail()
	}
	if !cursor.ToNextToken().IsText() {
		t.Fail()
	}
	if !cursor.ToNextToken().IsComment() {
		t.Fail()
	}
	if !cursor.ToNextToken().IsText() {
		t.Fail()
	}
	if !cursor.ToNextToken().IsEnd() {
		t.Fail()
	}
}

func BenchmarkManipulation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parse, err := parser.Parse(strings.NewReader("<a><b></b></a>"))
		if err != nil {
			b.FailNow()
		}
		cursor := NewCursor(parse)
		cursor.ToNextToken()

		cursor.ToEndToken()
		cursor.BeginElement(xml.Name{Local: "c"})

		cursor.ToStartDoc()
	}
}
