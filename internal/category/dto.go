package category

type CreateInput struct {
	Name string `json:"name"`
	Type Type   `json:"type"`
}

type UpdateInput struct {
	Name *string `json:"name,omitempty"`
	Type *Type   `json:"type,omitempty"`
}

type ListFilter struct {
	Type *Type
}
