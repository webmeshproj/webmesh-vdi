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

import RFB from '@novnc/novnc'
import { Emitter, Events } from './events.js'

// A base implementation for a display to be extended by objects using different protocols.
class Display extends Emitter {
    // A generic connector that calls to the child implementation, emitting any errors
    async connect(view, displayUrl) {
        try {
            await this._connect(view, displayUrl)
        } catch (err) {
            this.emit(Events.error, err)
            throw err
        }
    }

    // A generic disconnector that calls to the child implementation, emitting any errors
    async disconnect() {
        try {
            await this._disconnect()
        } catch (err) {
            this.emit(Events.error, err)
            throw err
        }
    }

    // Convenience wrapper for calling a method on a child implementation, emitting any errors.
    call(method, data) {
        if (typeof (this[method]) === 'function') {
            this[method](data)
        }
    }
}

// A display object that handles the canvas with a feed from an RFB connection
export class VNCDisplay extends Display {
    async _connect(view, displayUrl) {
        if (this._rfbClient) { 
            console.log('An RFB client already appears to be connected, returning')
            return 
        }
        console.log('Creating RFB connection')
        this._rfbClient = new RFB(view, displayUrl)
        this._rfbClient.addEventListener('connect', (ev) => { this._connectedToRFBServer(ev) })
        this._rfbClient.addEventListener('disconnect', (ev) => { this._disconnectedFromRFBServer(ev) })
        this._rfbClient.addEventListener('clipboard', (ev) => { this._handleRecvClipboard(ev) })
        this._rfbClient.resizeSession = true
        this._rfbClient.scaleViewport = true
    }

    async _disconnect() {
        if (this._rfbClient) { await this._rfbClient.disconnect() }
    }

    // _connectedToRFBServer is called when the RFB connection is established
    // with the desktop session.
    _connectedToRFBServer (ev) {
        console.log('Connected to RFB server')        
        const canvas = document.querySelector('canvas')
        canvas.addEventListener('keydown', (e) => {
            if (e.ctrlKey && e.location === 2) { // secondary ctrl locks pointer
                console.log('Locking pointer to view dom')
                // canvas.requestPointerLock() // https://github.com/novnc/noVNC/pull/1520
            }
        })
        this.emit(Events.connected, ev)
    }

    // _disconnectedFromRFBServer is called when the connection is dropped to a
    // desktop session.
    async _disconnectedFromRFBServer (event) {
        console.log('Disconnected from RFB server')
        if (this._rfbClient) {
            this._rfbClient = null
        }
        this.emit(Events.disconnected, event)
    }

    // _handleRecvClipboard is called when the RFB connection sends clipboard data
    // from the server.
    async _handleRecvClipboard (ev) {
        if (!ev.detail.text) {
            console.log(`Received invalid clipboard event: ${ev}`)
            return
        }
        try {
            await navigator.clipboard.writeText(ev.detail.text)
            console.log('Synced remote clipboard contents to local')
        } catch (err) {
            this.emit(Events.error, err)
            throw err
        }
    }

    // sendClipboard sends the given data to the clipboard in the instance
    sendClipboard (data) {
        console.log('Sending clipboard contents to RFB server')
        if (!this._rfbClient) {
            return
        }
        this._rfbClient.clipboardPasteFrom(data)
    }

    // setQualityLevel sets the quality level on an RFB connection
    setQualityLevel (lvl) {
        if (this._rfbClient) {
            this._rfbClient.qualityLevel = lvl
        }
    }
}

// export class SpiceDisplay extends Display {
//     async _connect(view) {

//     }

//     async _disconnect() {

//     }
// }