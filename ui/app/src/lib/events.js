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

export const Events = Object.freeze({'connected': 1, 'disconnected': 2, 'update': 3, 'error': 4})

export class Emitter {
    constructor() {
        this._callbacks = {}
    }

    // on adds an event listener to the extending object
    on(event, f) {
        if (this._callbacks[event]) {
            this._callbacks[event].push(f)
            return
        }
        this._callbacks[event] = [f]
    }

    // emit broadcasts the given event and details on the object
    emit(event, details) {
        if (this._callbacks[event]) {
            this._callbacks[event].forEach((f) => { f(details) })
        }
    }

    // bind will copy the event listeners of the provided object
    // to this one.
    bind(obj) {
        if (!obj._callbacks) { return }
        for (const [event, cbs] of Object.entries(obj._callbacks)) {
            cbs.forEach((cb) => { this.on(event, cb) })
        }
    }
}