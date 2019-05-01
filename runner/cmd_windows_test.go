package runner

import "testing"

func TestIsNoTaskRunning(t *testing.T) {
	tests := map[string]struct {
		output string
		want   bool
	}{
		"should return true when have no task running": {output: "INFO: No tasks are running which match the specified criteria.\r\n", want: true},
		"should return false when have task running":   {output: "", want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := isNoTaskRunning(tc.output)

			if got != tc.want {
				t.Fatalf("expected: %v, got: %v", tc.want, got)
			}
		})
	}
}
