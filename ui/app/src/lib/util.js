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

export function iframeRef (frameRef) {
  return frameRef.contentWindow
    ? frameRef.contentWindow.document
    : frameRef.contentDocument
}

// getErrorMessage turns a given error into a human readable string
export async function getErrorMessage (err) {
  console.log(err.code)
  console.log(err.message)
  console.log(err.stack)
  if (err.response !== undefined && err.response.data !== undefined) {
    if (err.response.data.error !== undefined) {
      return err.response.data.error
    }
    try {
      // This might be a json error from trying to download a file
      const text = await err.response.data.text()
      const errData = JSON.parse(text)
      if (errData.error !== undefined) {
        return errData.error
      }
    } catch {}
  }

  if (err.message) {
    return err.message
  }

  return err
}
