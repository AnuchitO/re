package runner

import "testing"

func TestIsNoTaskRunning(t *testing.T) {
	t.Run("should return true when have no task running", func(t *testing.T) {
		output := "INFO: No tasks are running which match the specified criteria.\r\n"

		yes := isNoTaskRunning(output)

		if !yes {
			t.Error("expect should be return true but it not.")
		}
	})

	t.Run("should return false when have task running", func(t *testing.T) {
		output := ""

		yes := isNoTaskRunning(output)

		if yes {
			t.Error("expect should be return false but it not.")
		}
	})
}
