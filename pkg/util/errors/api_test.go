/*
Copyright 2020,2021 Avi Zimmerman

This file is part of kvdi.

kvdi is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kvdi is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with kvdi.  If not, see <https://www.gnu.org/licenses/>.
*/

package errors

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestAPIError(t *testing.T) {
	apiErr := ToAPIError(errors.New("fake error"), ServerError)
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
