package drp

import (
	"encoding/json"
	"fmt"

	"github.com/go-test/deep"
)

func diffObjects(exp, fnd interface{}, t string) error {
	b1, _ := json.MarshalIndent(exp, "", "  ")
	b2, _ := json.MarshalIndent(fnd, "", "  ")
	if string(b1) != string(b2) {
		return fmt.Errorf("json diff: %s: %v\n%v\n", t, string(b1), string(b2))

	}
	if diff := deep.Equal(exp, fnd); diff != nil {
		return fmt.Errorf("%s doesn't match: %v", t, diff)
	}
	return nil
}
