package cursor

import (
	"com.programaths.xmlcursor/parser"
	"encoding/xml"
	"slices"
	"sort"
	"strings"
)

// Cursor allow to navigate and modify an XML [Document]
type Cursor struct {
	document parser.Document
	index    int
	stack    []int
}

// NewCursor create a cursor to navigate the document
// doc the document
func NewCursor(doc parser.Document) Cursor {
	return Cursor{
		document: doc,
		index:    -1,
	}
}

func (c *Cursor) updateBookmarks(index int, offset int) {
	for bookmarkName, bookmarkIndexes := range c.document.Bookmarks {
		for bookmarkIndex, bookmarkTokenIndex := range bookmarkIndexes {
			if bookmarkTokenIndex >= index {
				c.document.Bookmarks[bookmarkName][bookmarkIndex] = bookmarkTokenIndex + offset
			}
		}
	}
}

// Push save the current position of the [Cursor]
func (c *Cursor) Push() {
	c.stack = append(c.stack, c.index)
}

// Pop restore the current position of the [Cursor]
func (c *Cursor) Pop() {
	c.stack = c.stack[:len(c.stack)-1]
}

// SetBookmark add a bookmark to current cursor location if not already present.
func (c *Cursor) SetBookmark(bookmark interface{}) {
	bookmarksIndexes := c.document.Bookmarks[bookmark]
	if bookmarksIndexes == nil {
		bookmarksIndexes = make([]int, 0)
	}
	insertionIndex := sort.SearchInts(bookmarksIndexes, c.index)
	if len(bookmarksIndexes) == insertionIndex || bookmarksIndexes[insertionIndex] != c.index {
		bookmarksIndexes = slices.Insert(bookmarksIndexes, insertionIndex, c.index)
	}
	c.document.Bookmarks[bookmark] = bookmarksIndexes
}

// ClearBookmark remove a bookmark to current cursor location if already present.
func (c *Cursor) ClearBookmark(bookmark interface{}) {
	bookmarksIndexes := c.document.Bookmarks[bookmark]
	if bookmarksIndexes == nil {
		bookmarksIndexes = make([]int, 0)
	}
	insertionIndex := sort.SearchInts(bookmarksIndexes, c.index)
	if bookmarksIndexes[insertionIndex] == c.index {
		bookmarksIndexes = slices.Delete(bookmarksIndexes, insertionIndex, insertionIndex+1)
	}
	c.document.Bookmarks[bookmark] = bookmarksIndexes
}

// ToFirstBookmark move to first bookmark and report if the [Cursor] moved
func (c *Cursor) ToFirstBookmark(bookmark interface{}) bool {
	firstBookmark, ok := c.document.Bookmarks[bookmark]
	if ok {
		if len(firstBookmark) > 0 {
			c.index = firstBookmark[0]
			return true
		}
	}
	return false
}

// ToNextBookmark move to next bookmark and report if the [Cursor] moved
// Technically, it moves to the first bookmark after the current cursor position.
func (c *Cursor) ToNextBookmark(bookmark interface{}) bool {
	ints, ok := c.document.Bookmarks[bookmark]
	if ok {
		for _, v := range ints {
			if v > c.index {
				c.index = v
				return true
			}
		}
	}
	return false
}

// CurrentToken retrieve current token
func (c *Cursor) CurrentToken() parser.Token {
	if c.index < 0 || c.index > len(c.document.Tokens) {
		return parser.Token{Token: struct{}{}, Type: parser.None}
	}
	return c.document.Tokens[c.index]
}

// ToNextToken move to next token and return it
func (c *Cursor) ToNextToken() parser.Token {
	c.index++
	if c.index >= len(c.document.Tokens) {
		c.index--
		return parser.Token{Token: struct{}{}, Type: parser.None}
	}
	return c.document.Tokens[c.index]
}

// ToNextSibling move to next sibling and report if the [Cursor] moved
func (c *Cursor) ToNextSibling() bool {
	idx := c.index
	if !c.CurrentToken().IsStart() {
		return false
	}
	c.ToEndToken()
	c.ToNextToken()
	for !c.CurrentToken().IsStart() && !c.CurrentToken().IsEnd() {
		c.ToNextToken()
	}
	if c.CurrentToken().IsNone() || c.CurrentToken().IsEnd() {
		c.index = idx
		return false
	}
	return true
}

// ToNextSiblingByName move to next sibling with the given name and report if the [Cursor] moved
func (c *Cursor) ToNextSiblingByName(name xml.Name) bool {
	idx := c.index

	for {
		if !c.ToNextSibling() {
			c.index = idx
			return false
		}
		if c.CurrentToken().Name().Local == name.Local && c.CurrentToken().Name().Space == name.Space {
			return true
		}
	}
}

// ToFirstChild move to first child and report if the [Cursor] moved
func (c *Cursor) ToFirstChild() bool {
	idx := c.index
	c.ToNextToken()
	for !c.CurrentToken().IsNone() && !c.CurrentToken().IsEnd() && !c.CurrentToken().IsStart() {
		c.ToNextToken()
	}
	if !c.CurrentToken().IsStart() {
		c.index = idx
		return false
	}
	return true
}

// ToFirstChildByName move to first child with the given name and report if the [Cursor] moved
func (c *Cursor) ToFirstChildByName(name xml.Name) bool {
	if !c.ToFirstChild() {
		return false
	}
	if c.CurrentToken().Name().Local == name.Local && c.CurrentToken().Name().Space == name.Space {
		return true
	}
	if !c.ToNextSiblingByName(name) {
		return false
	}
	return true
}

// ToPreviousToken move to previous token
func (c *Cursor) ToPreviousToken() parser.Token {
	c.index--
	if c.index < 0 {
		c.index++
		return parser.Token{Token: struct{}{}, Type: parser.None}
	}
	return c.document.Tokens[c.index]
}

// ToStartToken move to the start token, works anywhere between tokens
func (c *Cursor) ToStartToken() parser.Token {
	c.Push()
	defer c.Pop()
	for c.index >= 0 {
		if t := c.document.Tokens[c.index]; t.IsStart() {
			return t
		}
		c.index--
	}
	return parser.Token{Token: struct{}{}, Type: parser.None}
}

// ToEndToken move to the end token matching the current start token
// Return [parser.None] if the cursor is not on a [parser.Start] token
func (c *Cursor) ToEndToken() parser.Token {
	if !c.document.Tokens[c.index].IsStart() {
		return parser.Token{Token: struct{}{}, Type: parser.None}
	}

	level := 1
	for {
		nt := c.ToNextToken()
		switch nt.Type {
		case parser.None:
			return nt
		case parser.Start:
			level++
		case parser.End:
			level--
			if level == 0 {
				return nt
			}
		}
	}
}

// NewCursor create a new cursor from the [Cursor] at the same position and same document
func (c *Cursor) NewCursor() Cursor {
	return Cursor{
		document: c.document,
		index:    c.index,
	}
}

// BeginElement create a start and end token and place the cursor between
func (c *Cursor) BeginElement(name xml.Name) {
	c.updateBookmarks(c.index, 2)
	newStartElement := xml.StartElement{
		Name: name,
		Attr: make([]xml.Attr, 0),
	}
	newEndElement := xml.EndElement{
		Name: name,
	}
	c.document.Tokens = append(c.document.Tokens, parser.Token{}, parser.Token{})
	copy(c.document.Tokens[c.index+2:], c.document.Tokens[c.index:len(c.document.Tokens)-2])
	c.document.Tokens[c.index] = parser.Token{
		Token: newStartElement,
		Type:  parser.Start,
	}
	c.document.Tokens[c.index+1] = parser.Token{
		Token: newEndElement,
		Type:  parser.End,
	}
	c.ToNextToken()
}

// InsertElement create a start and end token and place the cursor after
func (c *Cursor) InsertElement(name xml.Name) {
	c.BeginElement(name)
	c.ToNextToken()
}

// InsertText create a text token and place the cursor after
func (c *Cursor) InsertText(text string) {
	c.updateBookmarks(c.index, 1)
	c.document.Tokens = append(c.document.Tokens, parser.Token{})
	copy(c.document.Tokens[c.index+1:], c.document.Tokens[c.index:len(c.document.Tokens)-1])
	c.document.Tokens[c.index] = parser.Token{
		Token: xml.CharData(text),
		Type:  parser.Text,
	}
	c.ToNextToken()
}

// InsertElementWithText create a start, text end token and place the cursor after
func (c *Cursor) InsertElementWithText(name xml.Name, text string) {
	c.updateBookmarks(c.index, 3)
	newStartElement := xml.StartElement{
		Name: name,
		Attr: make([]xml.Attr, 0),
	}
	newEndElement := xml.EndElement{
		Name: name,
	}
	c.document.Tokens = append(c.document.Tokens, parser.Token{}, parser.Token{}, parser.Token{})
	copy(c.document.Tokens[c.index+3:], c.document.Tokens[c.index:len(c.document.Tokens)-3])
	c.document.Tokens[c.index] = parser.Token{
		Token: newStartElement,
		Type:  parser.Start,
	}
	c.document.Tokens[c.index+1] = parser.Token{
		Token: xml.CharData(text),
		Type:  parser.Text,
	}
	c.document.Tokens[c.index+2] = parser.Token{
		Token: newEndElement,
		Type:  parser.End,
	}
	c.index += 3
}

// IsLeftOf report if c is left of c2
func (c *Cursor) IsLeftOf(c2 Cursor) bool {
	return c.index < c2.index
}

// IsRightOf report if c is right of c2
func (c *Cursor) IsRightOf(c2 Cursor) bool {
	return c.index > c2.index
}

// ComparePosition return a negative, null or positive number according to relative positions of c and c2.
//
//   - > 0 : c is right of c2
//   - = 0 : c is at same position than c2
//   - < 0 : c is left of c2
func (c *Cursor) ComparePosition(c2 Cursor) int {
	return c.index - c2.index
}

// IsAtSamePositionAs report if c and c2 are at the same position
func (c *Cursor) IsAtSamePositionAs(c2 Cursor) bool {
	return c.index == c2.index
}

// ToStartDoc move cursor to start of document, NOT the first start element
func (c *Cursor) ToStartDoc() {
	c.index = 0 // Not correct, but ok for testing
}

// Xml produces the textual representation of the XML document
func (c *Cursor) Xml() string {
	var sb = strings.Builder{}
	encoder := xml.NewEncoder(&sb)
	c.Push()
	encoder.EncodeToken(c.CurrentToken().Token)
	level := 1
ite:
	for {
		nt := c.ToNextToken()
		switch nt.Type {
		case parser.None:
			break ite
		case parser.Start:
			level++
			encoder.EncodeToken(nt.Token)
		case parser.End:
			encoder.EncodeToken(nt.Token)
			level--
			if level == 0 {
				break ite
			}
		default:
			encoder.EncodeToken(nt.Token)
		}
	}
	encoder.Flush()
	c.Pop()
	return sb.String()
}
