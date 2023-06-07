package v1

// PostgreSQLResourceName is a kubernetes validation for a PostgreSQL resource name.
// +optional
// +kubebuilder:validation:Pattern:="^[a-zA-Z_][a-zA-Z0-9_]+$"
// +kubebuilder:validation:MaxLength:=61
type PostgreSQLResourceName string

func (d PostgreSQLResourceName) String() string {
	return string(d)
}
