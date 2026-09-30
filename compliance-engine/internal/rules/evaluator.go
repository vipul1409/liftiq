package rules

// MetricValues maps metric name → latest telemetry value.
// Any metric absent from the map produces a Result with Status Unknown.
type MetricValues map[string]float64

// EvaluateAll runs all 20 ASME rules against the provided metric values and
// returns exactly 20 results in the same order as ASME20.
// Rules whose metric is absent from values are marked Unknown.
func EvaluateAll(values MetricValues) []Result {
	results := make([]Result, 0, len(ASME20))
	for _, rule := range ASME20 {
		v, ok := values[rule.Metric]
		if !ok {
			results = append(results, Result{
				RuleID:      rule.ID,
				Description: rule.Description,
				ASMERef:     rule.ASMERef,
				Metric:      rule.Metric,
				Threshold:   rule.Threshold,
				Unit:        rule.Unit,
				Comparison:  rule.Comparison,
				Subsystem:   rule.Subsystem,
				Status:      Unknown,
				Message:     "no current telemetry reading for this metric (stale or missing)",
			})
			continue
		}
		results = append(results, rule.Evaluate(v))
	}
	return results
}
