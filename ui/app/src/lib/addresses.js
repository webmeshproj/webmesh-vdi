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

// DesktopAddressGetter is a convenience wrapper around retrieving connection
// URLs for a given desktop instance.
export default class DesktopAddressGetter {
    // constructor takes the Vuex user session store (for token retrieval) and
    // the namespace and name of the desktop instance.
    constructor (userStore, namespace, name) {
      this.userStore = userStore
      this.namespace = namespace
      this.name = name
    }
  
    // _getToken returns the current authentication token.
    _getToken () {
      return this.userStore.getters.token
    }
  
    // _buildAddress builds a websocket address for the given desktop function (endpoint).
    _buildAddress (endpoint) {
      return `${window.location.origin.replace('http', 'ws')}/api/desktops/ws/${this.namespace}/${this.name}/${endpoint}?token=${this._getToken()}`
    }
  
    // displayURL returns the websocket address for display connections.
    displayURL () {
      return this._buildAddress('display')
    }
  
    // audioURL returns the websocket address for audio connections.
    audioURL () {
      return this._buildAddress('audio')
    }
  
    // statusURL returns the websocket address for querying desktop status.
    statusURL () {
      return this._buildAddress('status')
    }

    logsFollowURL (container) {
        return this._buildAddress(`logs/${container}`)
    }
  
    logsURL (container) {
        return `/api/desktops/${this.namespace}/${this.name}/logs/${container}`
    }

  }