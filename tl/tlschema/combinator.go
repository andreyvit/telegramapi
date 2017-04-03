package tlschema

type Combinator struct {
	ID       uint32
	FullName string
}

type Type struct {
	constructors []*Combinator
}
