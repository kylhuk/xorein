package mobile

// BudgetClass describes battery budget envelopes.
type BudgetClass string

const (
	BudgetBackground  BudgetClass = "background"
	BudgetInteractive BudgetClass = "interactive"
	BudgetCritical    BudgetClass = "critical"
)

// BackgroundBudget selects a budget class deterministically.
func BackgroundBudget(cpuBudget int, batteryPercent int) BudgetClass {
	if batteryPercent < 20 || cpuBudget > 80 {
		return BudgetCritical
	}
	if cpuBudget > 40 {
		return BudgetInteractive
	}
	return BudgetBackground
}

// WakePolicy captures wake/suppression guidance.
type WakePolicy struct {
	AllowWake         bool
	SuppressionReason string
}

// EvaluateWakePolicy determines whether background wakeups should proceed.
func EvaluateWakePolicy(batteryPercent int, pendingHighPriority bool) WakePolicy {
	if pendingHighPriority && batteryPercent > 15 {
		return WakePolicy{AllowWake: true, SuppressionReason: "high-priority"}
	}
	if batteryPercent < 25 {
		return WakePolicy{AllowWake: false, SuppressionReason: "low battery"}
	}
	return WakePolicy{AllowWake: true, SuppressionReason: "budget available"}
}

// BatteryOptimizationDecision returns textual guidance for logging.
func BatteryOptimizationDecision(cpuBudget int, networkActive bool) string {
	if !networkActive {
		return "suppress syncs"
	}
	if cpuBudget > 70 {
		return "reduce render quality"
	}
	return "stay active"
}
