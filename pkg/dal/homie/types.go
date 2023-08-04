package homie

type Nodes map[string]Node

type Node struct {
	Name string
	Type string

	Properties Properties
}

type Properties map[string]Property

type Property struct {
	Name     string
	DataType string

	Format   string
	Settable bool
	Retained bool
	Unit     string

	Value any
}
