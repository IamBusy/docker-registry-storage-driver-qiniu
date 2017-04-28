package qiniu

import (
	"testing"
	"fmt"
)

func TestFromParameters(t *testing.T) {
	testData := map[string]interface{}{

	}
	_,err := FromParameters(testData)
	fmt.Print(err)
}
