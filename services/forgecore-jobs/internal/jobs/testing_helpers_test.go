package jobs_test

import "encoding/json"

// mustMarshal marshals a payload or panics — for test setup only.
func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
