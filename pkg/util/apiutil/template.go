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

package apiutil

import (
	"bytes"
	"encoding/json"
	"html/template"

	desktopvsv1 "github.com/tinyzimmer/kvdi/apis/desktops/v1"
	"github.com/tinyzimmer/kvdi/pkg/types"
)

// RenderTemplate renders the given template with the provided session data.
func RenderTemplate(tmpl *desktopvsv1.Template, data *types.JWTClaims) error {
	body, err := json.Marshal(tmpl)
	if err != nil {
		return err
	}
	t, err := template.New("").Parse(string(body))
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, map[string]interface{}{
		"Session": data,
	}); err != nil {
		return err
	}
	return json.Unmarshal(buf.Bytes(), tmpl)
}
