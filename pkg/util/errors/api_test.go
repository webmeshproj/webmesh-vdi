package errors

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestAPIError(t *testing.T) {
	apiErr := ToAPIError(errors.New("fake error"))
	apiJSON := apiErr.JSON()

	res := make(map[string]string)
	if err := json.Unmarshal(apiJSON, &res); err != nil {
		t.Fatal(err)
	}
	if resErr, ok := res["error"]; !ok {
		t.Error("Expoected 'error' field in error json")
	} else if resErr != apiErr.Error() {
		t.Error("Error changed during marshaling")
	}
}
