package voice

import "fmt"

// NoiseReducer identifies the chosen noise suppression implementation.
type NoiseReducer string

const (
	RNNoise NoiseReducer = "rnnoise"
	DTLN    NoiseReducer = "dtln"
)

// Decision records the selection made for noise suppression.
type Decision struct {
	Selected NoiseReducer
	Reason   string
}

// SelectNoiseReducer applies the policy to pick RNNoise when the environment is stable.
func SelectNoiseReducer(studio bool, fallbackErr error) Decision {
	if studio {
		return Decision{Selected: RNNoise, Reason: "stable environment"}
	}
	if fallbackErr != nil {
		return Decision{Selected: DTLN, Reason: classifyFallback(fallbackErr)}
	}
	return Decision{Selected: DTLN, Reason: "baseline fallback"}
}

func classifyFallback(err error) string {
	if err == nil {
		return "no error"
	}
	return fmt.Sprintf("fallback triggered: %s", err.Error())
}
