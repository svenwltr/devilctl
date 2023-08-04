package homie

type Device struct {
	Name           string
	Implementation string
	NodeIDs        []string
}

type Node struct {
	NodeID      string
	Name        string
	Type        string
	PropertyIDs []string
}

type Property struct {
	NodeID     string
	PropertyID string

	Name     string
	DataType string

	Format   string
	Settable bool
	Retained bool
	Unit     string
}
