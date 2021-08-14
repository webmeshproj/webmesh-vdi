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

	desktopvsv1 "github.com/kvdi/kvdi/apis/desktops/v1"
)

// RenderTemplate renders the given template with the provided data.
func RenderTemplate(tmpl *desktopvsv1.Template, data interface{}) error {
	body, err := json.Marshal(tmpl)
	if err != nil {
		return err
	}
	t, err := template.New("").Parse(string(body))
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return err
	}
	return json.Unmarshal(buf.Bytes(), tmpl)
}
