package roles

// RoleType defines the available roles for agents.
type RoleType string

const (
	RoleDiagnoser RoleType = "diagnoser"
	RoleExplainer RoleType = "explainer"
	RoleDiagnose  RoleType = "diagnose"
)

// RolePromptTemplate maps roles to their prompt templates.
var RolePromptTemplate = map[RoleType]string{
	RoleDiagnoser: "You are a NixOS diagnostics agent. Analyze the following input and provide a diagnosis:",
	RoleExplainer: "You are a NixOS explainer. Explain the following input in simple terms:",
	RoleDiagnose:  "You are the NixAI diagnose agent. Use all available context (logs, configs, user input) to identify and explain the root cause of the user's NixOS problem. Provide actionable steps for resolution:",
}

// ValidateRole checks if a role is supported.
func ValidateRole(role string) bool {
	switch RoleType(role) {
	case RoleDiagnoser, RoleExplainer, RoleDiagnose:
		return true
	default:
		return false
	}
}
