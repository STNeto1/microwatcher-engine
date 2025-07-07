package iter

func Map[TInput any, TOutput any](iter []TInput, fn func(TInput) TOutput) []TOutput {
	ret := make([]TOutput, len(iter))

	for i, v := range iter {
		ret[i] = fn(v)
	}

	return ret
}
