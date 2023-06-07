package v1

// DatabaseAccountOnDelete is the options that can be set for onDelete.
// +kubebuilder:validation:Enum=retain;delete
type DatabaseAccountOnDelete string

func (d DatabaseAccountOnDelete) String() string {
	return string(d)
}

const (
	// OnDeleteRetain retain the database and user.
	OnDeleteRetain DatabaseAccountOnDelete = "retain"

	// OnDeleteDelete delete the created database and user.
	OnDeleteDelete DatabaseAccountOnDelete = "delete"
)

// DatabaseAccountCreateStage is the stage the account creation is up to.
// +kubebuilder:validation:Enum=Init;UserCreate;DatabaseCreate;Error;Ready
type DatabaseAccountCreateStage string

func (d DatabaseAccountCreateStage) String() string {
	return string(d)
}

const (
	// UnknownStage is the first stage of creating the account.
	UnknownStage DatabaseAccountCreateStage = ""

	// InitStage is the first stage of creating the account.
	InitStage DatabaseAccountCreateStage = "Init"

	// UserCreateStage is the step where the account creation has been started.
	UserCreateStage DatabaseAccountCreateStage = "UserCreate"

	// DatabaseCreateStage is the step where the account creation has been started.
	DatabaseCreateStage DatabaseAccountCreateStage = "DatabaseCreate"

	// ErrorStage is when the account has failed and won't be completed without changes.
	ErrorStage DatabaseAccountCreateStage = "Error"

	// ReadyStage is when the account is ready to be used.
	ReadyStage DatabaseAccountCreateStage = "Ready"
)
