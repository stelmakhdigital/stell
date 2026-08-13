package llm

import (
	"errors"
	"testing"
)

func TestIsRetryableNetErr(t *testing.T) {
	if !isRetryableNetErr(errors.New(`write tcp 1.2.3.4:1->5.6.7.8:8000: write: broken pipe`)) {
		t.Fatal("broken pipe should retry")
	}
	if isRetryableNetErr(errors.New("context deadline exceeded")) {
		t.Fatal("timeout should not retry")
	}
}
