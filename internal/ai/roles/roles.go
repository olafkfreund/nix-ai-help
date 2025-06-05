package roles

// RoleType defines the available roles for agents.
type RoleType string

const (
	RoleDiagnoser RoleType = "diagnoser"
	RoleExplainer RoleType = "explainer"
)

// RolePromptTemplate maps roles to their prompt templates.
var RolePromptTemplate = map[RoleType]string{
	RoleDiagnoser: "You are a NixOS diagnostics agent. Analyze the following input and provide a diagnosis:",
	RoleExplainer: "You are a NixOS explainer. Explain the following input in simple terms:",
}

// ValidateRole checks if a role is supported.
func ValidateRole(role string) bool {
	switch RoleType(role) {
	case RoleDiagnoser, RoleExplainer:
		return true
	default:
		return false
	}
}
