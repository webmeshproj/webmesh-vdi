import RFB from '@novnc/novnc/core/rfb'
import WSAudioPlayer from '../lib/wsaudio.js'
import { DesktopAddressGetter } from '../lib/common.js'
import { parseConfigFileTextToJson } from 'app/node_modules/typescript/lib/typescript.js'

// DisplayManager handles display and audio connections to remote desktop sessions.
export default class DisplayManager {
    // Builds the DisplayManager instance. The userStore and sessionStore are Vuex
    // Store instances that reflect the currently logged in user and the current desktop
    // sessions respectively. The errCb is a function to call when internal errors occur.
    constructor ({ userStore, sessionStore, onError, onStatusUpdate, onDisconnect, onConnect }) {
        this.userStore = userStore
        this.sessionStore = sessionStore
        this.errCb = onError
        this.disconnectCb = onDisconnect
        this.connectCb = onConnect
        this.statusCb = onStatusUpdate
        // Current session represents the session that the rfbClient is currently connected
        // to. This is not always the same as the activeSession as far as the sessionStore
        // is concerned.
        this._currentSession = this._getActiveSession()
        this._statusText = ''
        this.rfbClient = null
        this.audioPlayer = null
        this.unsubscribeSessions = this.sessionStore.subscribe((mutation) => {
            this._handleSessionChange(mutation)
        })
    }

    // _getActiveSession returns the session currently marked as active in the session store.
    _getActiveSession () {
        return this.sessionStore.getters.activeSession
    }

    // _audioIsEnabled returns true if audio is currently enabled
    _audioIsEnabled () {
        return this.sessionStore.getters.audioEnabled
    }

    // _getSessionURLs returns an object that can easily retrieve URLs for the different
    // websocket endpoints.
    _getSessionURLs () {
        const activeSession = this._getActiveSession()
        return new DesktopAddressGetter(
            this.userStore,
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
        if (this._audioIsEnabled()) {
            this._enableAudio()
        } else {
            this._disableAudio()
        }
    }

    // _enableAudio starts a new audio stream over the websocket endpoint.
    _enableAudio () {
        console.log('Enabling audio playback')
        const urls = this._getSessionURLs()
        const audioUrl = urls.audioURL()
        console.log(`Connecting to audio stream at ${audioUrl}`)
        const playerCfg = { server: { url: audioUrl } }
        this.audioPlayer = new WSAudioPlayer(playerCfg)
        this.audioPlayer.startPlayback()
    }

    // _disableAudio will stop an audio stream if it is currently running.
    _disableAudio() {
        if (this.audioPlayer !== null) {
            console.log('Stopping audio stream')
            this.audioPlayer.close()
            this.audioPlayer = null
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
    _doStatusWebsocket () {

        const urls = this._getSessionURLs()
        const activeSession = this._getActiveSession()
        const socket = new WebSocket(urls.statusURL())

        let msgCount = 0

        socket.onopen = (e) => {
            this.statusCb(`Starting desktop ${activeSession.namespace}/${activeSession.name}`)
        }
          
        socket.onmessage = (event) => {
            msgCount++

            console.log(`[status] Status update ${msgCount} received from server: ${event.data}`)
            const st = JSON.parse(event.data)
            if (st.error) {
                this._currentSession = this._getActiveSession()
                this.disconnectCb()
                this.errCb(new Error(st.error))
                return
            }
            if (this._statusIsReady(st)) {
                this._createConnection()
                if (socket.readyState === 1) {
                    socket.close()
                }
                this.statusCb('')
            }

            let statusText = `Waiting for ${activeSession.namespace}/${activeSession.name}`
            if (msgCount > 5) {
                statusText += '\n\nThis is taking a while, the server might be pulling the image for the first time'
            }
            this.statusCb(statusText)
        }

        socket.onclose = (event) => {
            if (event.wasClean) {
                console.log(`[status] Connection closed cleanly, code=${event.code} reason=${event.reason}`)
            } else {
                this.errCb(new Error(`Error getting session status: ${event.code} ${event.reason}`))
            }
        }
          
        socket.onerror = (err) => {
            this.errCb(err)
            this.disconnectCb()
        }

    }

    // _createConnection will create a new RFB connection if the socketType
    // is xvnc. xpra sockets use the official client embedded in an iframe.
    async _createConnection () {
        if (this._currentSession.socketType !== 'xvnc') {
            this.connectCb()
            return
        }
        const urls = this._getSessionURLs()
        // get the websocket address with the token included as a query argument
        const displayURL = urls.displayURL()
        // get the view port for the display
        const view = document.getElementById('view')
        if (view === null || view === undefined) {
            return
        }
        try {
            // create a vnc connection
            await this._createRFBConnection(view, displayURL)
        } catch (err) {
            this.disconnectCb()
            this._currentSession = null
            this.errCb(err)
            return
        }
        this.connectCb()
    }

    // _createRFBConnection creates a new RFB connection.
    async _createRFBConnection (view, url) {
        const rfb = new RFB(view, url)
        rfb.addEventListener('connect', () => { this._connectedToRFBServer() })
        rfb.addEventListener('disconnect', (ev) => { this._disconnectedFromRFBServer(ev) })
        rfb.resizeSession = true
        this.rfbClient = rfb
    }

    // _connectedToRFBServer is called when the RFB connection is established
    // with the desktop session.
    _connectedToRFBServer() {
        const activeSession = this._getActiveSession()
        if (activeSession.socketType === 'xvnc') {
            this.rfbClient.scaleViewport = true
            this.rfbClient.resizeSession = true
        }
    }

    // _disconnectedFromRFBServer is called when the connection is dropped to a
    // desktop session.
    async _disconnectedFromRFBServer (event) {
        if (event.detail.clean) {
            // The server disconnecting cleanly would mean expired session,
            // but this should probably be handled better.
            if (this._currentSession) {
                try {
                    // check if the desktop still exists, if we get an error back
                    // it was deleted.
                    await this.sessionStore.getters.sessionStatus(this._currentSession)
                } catch {
                    this.sessionStore.dispatch('deleteSession', this._currentSession)
                    this._currentSession = null
                    this.errCb(new Error("The desktop session has ended"))
                }
            }
            console.log('Disconnected')
        } else {
            console.log('Something went wrong, connection is closed')
            this._doStatusWebsocket()
        }

        // no matter what, make user recreate audio connection
        // TODO: know that the user was using audio and recreate
        // stream automatically if session is still active.
        if (this.audioPlayer !== null) {
            this.audioPlayer.stop()
            this.audioPlayer = null
            this.sessionStore.dispatch('toggleAudio', false)
        }

        this._currentSession = this._getActiveSession()
        this.disconnectCb()
    }

    // _disconnect will close any connections currently open
    _disconnect () {
        if (this.rfbClient) {
            try {
                this.rfbClient.disconnect()
            } catch (err) {
                console.log(err)
            } finally {
                this.rfbClient = null
            }
        }
    }

    // destroy is called when the viewport holding this display manager is destroyed.
    // It will unsubscribe from the Vuex store and close any currently open connections.
    destroy () {
        this.unsubscribeSessions()
        this._disconnect()
    }

    // connect will query the status of the active desktop session and then open
    // a new display.
    connect () {
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

    // getConnectingStatus returns the status message for the connecting status
    getConnectingStatus () {
        return this._statusText
    }

    // syncClipboardData syncs the provied data to the clipboard inside the currently
    // active RFB connection.
    syncClipboardData (data) {
        if (!this.rfbClient) {
            return
        }
        const session = this._getActiveSession()
        if (session.socketType !== 'xvnc') {
            return
        }
        this.rfbClient.clipboardPasteFrom(data)
    }

    // xpraArgs returns the xpra args for the given desktop session
    xpraArgs () {
        const urls = this._getSessionURLs()
        return urls.xpraArgs()
    }
}