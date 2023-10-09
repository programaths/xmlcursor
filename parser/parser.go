package parser

import (
	"encoding/xml"
	"io"
)

type TokenType int

const (
	Start = TokenType(iota)
	End
	Text
	Comment
	ProcInst
	Directive
	None
)

// Token wrapping [xml.Token]
type Token struct {
	xml.Token
	Type TokenType
}

// Document can be manipulated with cursors
type Document struct {
	Tokens    []Token
	Bookmarks map[interface{}][]int
}

// IsStart indicates if the token is a start element
func (t Token) IsStart() bool {
	return t.Type == Start
}

// IsEnd indicates if the token is an end element
func (t Token) IsEnd() bool {
	return t.Type == End
}

// IsText indicates if the token is text/char data
func (t Token) IsText() bool {
	return t.Type == Text
}

// IsComment indicates if the token is a comment
func (t Token) IsComment() bool {
	return t.Type == Comment
}

// IsProcessingInstruction indicates if the token is a processing instruction
func (t Token) IsProcessingInstruction() bool {
	return t.Type == ProcInst
}

func (t Token) IsDirective() bool {
	return t.Type == Directive
}

// IsNone indicates if the token is a nothing, happens when you leave the [Document]
func (t Token) IsNone() bool {
	return t.Type == None
}

// Name retrieves the name of the token or nothing
func (t Token) Name() xml.Name {
	switch v := t.Token.(type) {
	case xml.StartElement:
		return v.Name
	case xml.EndElement:
		return v.Name
	}
	return xml.Name{}
}

// Text retrieves the text of the token or nothing
func (t Token) Text() string {
	switch v := t.Token.(type) {
	case xml.CharData:
		return string(v)
	case xml.Comment:
		return string(v)
	case xml.Directive:
		return string(v)
	}
	return ""
}

// Parse read input from the reader to create a [Document]
func Parse(r io.Reader) (Document,error) {

	var tokens = make([]Token, 0)
	decoder := xml.NewDecoder(r)
	token, err := decoder.Token()
	for err == nil {
		switch v := token.(type) {
		case xml.StartElement:
			tokens = append(tokens, Token{
				Token: v.Copy(),
				Type:  Start,
			})
		case xml.CharData:
			tokens = append(tokens, Token{
				Token: v.Copy(),
				Type:  Text,
			})
		case xml.Comment:
			tokens = append(tokens, Token{
				Token: v.Copy(),
				Type:  Comment,
			})
		case xml.ProcInst:
			tokens = append(tokens, Token{
				Token: v.Copy(),
				Type:  ProcInst,
			})
		case xml.Directive:
			tokens = append(tokens, Token{
				Token: v.Copy(),
				Type:  Directive,
			})
		case xml.EndElement:
			tokens = append(tokens, Token{
				Token: v,
				Type:  End,
			})
		}
		token, err = decoder.Token()
	}
	if err != nil && err != io.EOF {
		return Document{}, err
	}
	var doc = Document{
		Tokens:    tokens,
		Bookmarks: make(map[interface{}][]int),
	}
	return doc,nil
}
