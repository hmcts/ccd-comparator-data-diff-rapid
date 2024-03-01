package jsonx

type NodeAny map[string]any

func (n NodeAny) IsEmpty() bool {
	return len(n) == 0
}

type ArrayAny []any

func (a ArrayAny) IsEmpty() bool {
	return len(a) == 0
}

type NodeChange map[string]ArrayChange

func (n NodeChange) IsEmpty() bool {
	return len(n) == 0
}

type ArrayChange []Change

func (a ArrayChange) IsEmpty() bool {
	return len(a) == 0
}

type Change struct {
	Value   string `json:"value"`
	Deleted bool   `json:"deleted,omitempty"`
	Added   bool   `json:"added,omitempty"`
}

func (a Change) Compare(b Change) bool {
	return a.Value == b.Value
}

func (a Change) ChangeType() string {
	if a.Added {
		return "added"
	} else {
		return "deleted"
	}
}

func (a Change) HasCrossMatch(b Change) bool {
	return (a.Added && b.Deleted) || (a.Deleted == b.Added)
}
