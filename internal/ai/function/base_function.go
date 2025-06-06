package function

// Re-export base function types and functions from functionbase
import "nix-ai-help/internal/ai/functionbase"

type FunctionInterface = functionbase.FunctionInterface
type BaseFunction = functionbase.BaseFunction

var NewBaseFunction = functionbase.NewBaseFunction
var StringParam = functionbase.StringParam
var StringParamWithEnum = functionbase.StringParamWithEnum
var BoolParam = functionbase.BoolParam
var ObjectParam = functionbase.ObjectParam
var SuccessResult = functionbase.SuccessResult
var ErrorResult = functionbase.ErrorResult
