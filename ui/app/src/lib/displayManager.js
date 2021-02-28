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

import AudioManager from './audioManager.js'
import DesktopAddressGetter from './addresses.js'
import { Emitter, Events } from './events.js'
import { getDisplay } from './displays.js'

// DisplayManager handles display and audio connections to remote desktop sessions.
export default class DisplayManager extends Emitter {
    // Builds the DisplayManager instance. The userStore and sessionStore are Vuex
    // Store instances that reflect the currently logged in user and the current desktop
    // sessions respectively.
    constructor ({ userStore, sessionStore }) {
        super()
        // Vuex stores
        this._userStore = userStore
        this._sessionStore = sessionStore
        // Current session represents the session that the display is currently connected
        // to. This is not always the same as the activeSession as far as the sessionStore
        // is concerned.
        this._currentSession = this._getActiveSession()
        // A socket being used to query a desktop's boot status
        this._statusSocket = null
        // Status text to display to a user when a connection is pending
        this._statusText = ''
        // The display object managing the view canvas
        this._display = null
        // The audio player for streaming playback
        this._audioManager = null
        // Subscribe to changes to desktop sessions
        this._unsubscribeSessions = this._sessionStore.subscribe((mutation) => {
            this._handleSessionChange(mutation)
        })
    }

    // _getActiveSession returns the session currently marked as active in the session store.
    _getActiveSession () {
        return this._sessionStore.getters.activeSession
    }


    // _audioIsEnabled returns true if audio is currently enabled
    _audioIsEnabled () {
        return this._sessionStore.getters.audioEnabled
    }

    _recordingIsEnabled () {
        return this._sessionStore.getters.recordingEnabled
    }

    // _getSessionURLs returns an object that can easily retrieve URLs for the different
    // websocket endpoints.
    _getSessionURLs () {
        const activeSession = this._getActiveSession()
        return new DesktopAddressGetter(
            this._userStore,
            activeSession.namespace,
            activeSession.name
        )
    }

    // _handleSessionChange handles a change in the current session status in the session
    // vuex store.
    _handleSessionChange (mutation) {
        if (mutation.type === 'set_active_session') {
            this._handleActiveSessionChange()
        } else if (mutation.type === 'delete_session') {
            this._handleDeleteSession()
        } else if (mutation.type === 'toggle_audio') {
            this._handleToggleAudio()
        } else if (mutation.type === 'toggle_recording') {
            this._handleToggleRecording()
        }
    }

    // _handleActiveSessionChange is called when the active session is changed by the user.
    _handleActiveSessionChange () {
        const activeSession = this._getActiveSession()
        if (activeSession === undefined) {
            // The connected session was deleted and no other ones exist
            this._disconnect()
            return
        }
        if (this._currentSession === activeSession) {
            // Nothing actually changed
            return
        }
        // The user selected a new session
        this._disconnect()
        this._currentSession = this._getActiveSession()
        this._doStatusWebsocket()
    }

    // _handleDeleteSession is called when a user deletes a desktop session.
    _handleDeleteSession () {
        this._currentSession = this._getActiveSession()
    }

    // _handleToggleAudio is called when the user toggles audio.
    _handleToggleAudio () {
        if (!this._currentSession) { return }
        if (this._audioIsEnabled()) {
            this._enableAudio()
        } else {
            this._disableAudio()
        }
    }

    // _handleToggleRecording is called when the user toggles the microphone.
    _handleToggleRecording () {
        if (!this._currentSession) { return }
        if (this._recordingIsEnabled()) {
            this._enableRecording()
        } else {
            this._disableRecording()
        }
    }

    // _createAudioManager creates a new AudioManager object for this DisplayManager
    _createAudioManager () {
        this._audioManager = new AudioManager({
            addressGetter: this._getSessionURLs(),
            userStore: this._userStore
        })
        // Don't use bind here because the VNCViewer component is only expecting display
        // disconnected events.
        this._audioManager.on(Events.disconnected, () => { this._resetAudioStatus() })
        this._audioManager.on(Events.error, (err) => { this.emit(Events.error, err) })
    }

    // _enableAudio starts a new audio stream over the websocket endpoint.
    _enableAudio () {
        if (!this._audioManager) {
            this._createAudioManager()
        }
        console.log('Connecting to audio stream')
        this._audioManager.startPlayback()
    }

    // _enableRecording will stream microphone data to the audio input on the desktop session.
    _enableRecording () {
        // if there is no audioManager yet, call _enableAudio first, since it will also open
        // the websocket. I guess it should be possible to use microphone separate from playback.
        if (!this._audioManager) {
            this._enableAudio()
        }
        console.log('Starting microphone stream')
        this._audioManager.startRecording()
    }

    // _disableAudio will stop an audio stream if it is currently running.
    _disableAudio() {
        if (this._audioManager) {
            console.log('Stopping audio stream')
            try {
                this._audioManager.close()
            } finally {
                this._audioManager = null
            }
        }
    }

    // _disableRecording will disable microphone streaming.
    _disableRecording () {
        if (this._audioManager) {
            this._audioManager.stopRecording()
        }
    }

    // _statusIsReady returns true if the given desktop status message signals
    // that it is ready to serve display and audio connections.
    _statusIsReady (status) {
        return status.podPhase === 'Running' && status.running
    }

    // _doStatusWebsocket opens a websocket connection to the status endpoint for the
    // current desktop session. Once a message is received signaling the desktop is ready,
    // the socket is closed and a display connection is created.
    _doStatusWebsocket (retry) {

        const urls = this._getSessionURLs()
        const activeSession = this._getActiveSession()
        const socket = new WebSocket(urls.statusURL())
        console.log(activeSession)
        let msgCount = 0

        socket.onopen = () => {
            this.emit(Events.update, `Connecting to ${activeSession.namespace}/${activeSession.name}`)
        }
          
        socket.onmessage = (event) => {
            msgCount++

            console.log(`[status] Status update ${msgCount} received from server: ${event.data}`)
            const st = JSON.parse(event.data)

            // If there is an error on the pipe, send it to the user.
            if (st.error) {
                this._currentSession = this._getActiveSession()
                this.emit(Events.disconnected)
                this.emit(Events.error, new Error(st.error))
                return
            }

            // If the desktop is ready then create a connection, clear the status, and close this socket
            if (this._statusIsReady(st)) {
                console.log(`Desktop is ready, connecting`)
                this._createConnection()
                    .catch((err) => {
                        console.error(err)
                        console.log('Retrying connection with new token')
                        return this._userStore.dispatch('refreshToken')
                            .then(() => {
                                // only retry once
                                return this._createConnection()
                                    .catch((err) => { 
                                        this.emit(Events.disconnected)
                                        this.emit(Events.error, err)
                                     })
                            })
                    })
                if (socket.readyState === 1) {
                    socket.close()
                }
                this.emit(Events.update, 'Desktop is ready - Launching display')
                return
            }

            // Update the status text for the user
            let statusText = `Waiting for ${activeSession.namespace}/${activeSession.name}`
            if (msgCount > 6) {
                statusText += '\n\nThis is taking a while. The server might be pulling the'
                statusText += '\nimage for the first time, this is a large qemu disk image,'
                statusText += '\nor the control-plane is having trouble scheduling the desktop.'
            }
            this.emit(Events.update, statusText)
        }

        socket.onclose = (event) => {
            if (event.wasClean || event.code === 1000) {
                console.log(`[status] Connection closed cleanly, code=${event.code} reason=${event.reason}`)
            } else {
                if (event.code === 1006 && !retry) {
                    this._userStore.dispatch('refreshToken')
                        .then(() => {
                            this._doStatusWebsocket(true)
                        })
                        .catch((err) => {
                            this.emit(Events.error, err)
                            throw err
                        })
                    return
                }
                this.emit(Events.error, new Error(`Error getting session status: ${event.code} ${event.reason}`))
            }
        }
          
        socket.onerror = (err) => {
            if (retry) {
                this.emit(Events.disconnected)
                this.emit(Events.error, err)
            }
        }

        this._statusSocket = socket
    }

    // _createConnection will create a new display connection
    async _createConnection () {
        // get the websocket display address
        const urls = this._getSessionURLs()
        const displayURL = urls.displayURL()
        // get the viewport for the display
        const view = document.getElementById('view')
        if (view === null || view === undefined) {
            console.log('No view found in the window')
            return
        }

        const activeSession = this._getActiveSession()
        this._display = getDisplay(activeSession)
        this._display.bind(this)
        this._display.on(Events.disconnected, (ev) => { this._disconnectedFromDisplay(ev) })

        try {
            // create a display connection
            await this._display.connect(view, displayURL)
        } catch (err) {
            this._currentSession = this._getActiveSession()
            throw err
        }
    }


    // _disconnectedFromDisplay is called when the connection is dropped to a
    // display session.
    async _disconnectedFromDisplay (event) {
        if (!event || (event.detail && event.detail.clean)) {
            // The server disconnecting cleanly would mean expired session,
            // but this should probably be handled better.
            if (this._currentSession) {
                try {
                    // check if the desktop still exists, if we get an error back
                    // it was deleted.
                    await this._sessionStore.getters.sessionStatus(this._currentSession)
                } catch {
                    this._sessionStore.dispatch('deleteSessionOffline', this._currentSession)
                    this._currentSession = null
                    this.emit(Events.error, new Error("The desktop session has ended"))
                }
            }
        } else {
            console.log(`Something went wrong, connection is closed (${event}) - Reconnecting`)
            this._doStatusWebsocket()
        }

        // no matter what, make user recreate audio connection
        // TODO: know that the user was using audio and recreate
        // stream automatically if session is still active.
        if (this.audioPlayer) {
            this._disableAudio()
            this._resetAudioStatus()
        }

        this._currentSession = this._getActiveSession()
    }

    // _resetAudioStatus will reset the audio toggles in the Vuex store
    _resetAudioStatus () {
        this._sessionStore.dispatch('toggleAudio', false)
        this._sessionStore.dispatch('toggleRecording', false)
    }

    // _disconnect will close any connections currently open
    _disconnect () {
        if (this._display) {
            try {
                this._display.disconnect()
            } catch (err) {
                console.error(err)
            } finally {
                this._display = null
            }
        }
        if (this._statusSocket) {
            try {
                this._statusSocket.close()
            } catch (err) {
                console.error(err)
            } finally {
                this._statusSocket = null
            }
        }
    }

    // destroy is called when the viewport holding this display manager is destroyed.
    // It will unsubscribe from the Vuex store and close any currently open connections.
    destroy () {
        this._unsubscribeSessions()
        this._disconnect()
    }

    // connect will query the status of the active desktop session and then open
    // a new display.
    connect () {
        if (!this.hasActiveSession()) { return }
        this._doStatusWebsocket()
    }

    // hasActiveSession returns true if there is currently an active desktop session in 
    // the Vuex store.
    hasActiveSession () {
        const activeSession = this._getActiveSession()
        return activeSession !== undefined && activeSession !== null
    }

    // getCurrentSession returns the current session.
    getCurrentSession () {
        return this._currentSession
    }

    // sendClipboardData syncs the provied data to the clipboard inside the currently
    // active display connection.
    sendClipboardData (data) {
        if (!this._display) {
            return
        }
        this._display.call('sendClipboard', data)
    }

}
