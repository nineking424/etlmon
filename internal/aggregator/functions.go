package aggregator

// Avg calculates the average of a slice of float64
func Avg(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// Max returns the maximum value from a slice of float64
func Max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

// Min returns the minimum value from a slice of float64
func Min(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

// Last returns the last value from a slice of float64
func Last(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return values[len(values)-1]
}

// AggregationFunc is a type for aggregation functions
type AggregationFunc func([]float64) float64

// AggregationFuncs maps aggregation type names to functions
var AggregationFuncs = map[string]AggregationFunc{
	"avg":  Avg,
	"max":  Max,
	"min":  Min,
	"last": Last,
}
