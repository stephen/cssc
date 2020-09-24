package ast

// KeyframeSelectorList is a list of selectors used by @keyframes blocks.
type KeyframeSelectorList struct {
	Span

	Selectors []KeyframeSelector
}

// Location implements Node.
func (n *KeyframeSelectorList) Location() *Span { return &n.Span }

func (KeyframeSelectorList) isPrelude() {}

var _ Prelude = &KeyframeSelectorList{}

// KeyframeSelector is a selector for rules in a @keyframes block.KeyframeSelector
// Valid values are a Percentage or to/from.
type KeyframeSelector interface {
	Node

	isKeyframeSelector()
}

func (Percentage) isKeyframeSelector() {}
func (Identifier) isKeyframeSelector() {}

var _ KeyframeSelector = &Percentage{}
var _ KeyframeSelector = &Identifier{}
