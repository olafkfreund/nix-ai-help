package function

// Re-export types from functionbase to maintain compatibility
import "nix-ai-help/internal/ai/functionbase"

type FunctionParameter = functionbase.FunctionParameter
type FunctionSchema = functionbase.FunctionSchema
type FunctionExample = functionbase.FunctionExample
type FunctionCall = functionbase.FunctionCall
type FunctionResult = functionbase.FunctionResult
type FunctionExecution = functionbase.FunctionExecution
type Progress = functionbase.Progress
type ProgressCallback = functionbase.ProgressCallback
type FunctionOptions = functionbase.FunctionOptions
type ValidationError = functionbase.ValidationError
type FunctionError = functionbase.FunctionError

// Re-export utility functions
var ToJSON = functionbase.ToJSON
var FromJSON = functionbase.FromJSON
