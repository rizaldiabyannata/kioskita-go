package core

// BusinessTemplate merepresentasikan struktur satu file template JSON.
type BusinessTemplate struct {
	DisplayName string            `json:"displayName"`
	Schema      []AttributeSchema `json:"schema"`
}
