import RFB from '@novnc/novnc/core/rfb.js'
import AudioManager from './audioManager.js'

// DisplayManager handles display and audio connections to remote desktop sessions.
export default class DisplayManager {
    // Builds the DisplayManager instance. The userStore and sessionStore are Vuex
    // Store instances that reflect the currently logged in user and the current desktop
    // sessions respectively.
    constructor ({ userStore, sessionStore, onError, onStatusUpdate, onDisconnect, onConnect }) {
        // Vuex stores
        this._userStore = userStore
        this._sessionStore = sessionStore
        // Event listeners - I am sure there is a more correct way to do this
        this._errCb = onError
        this._disconnectCb = onDisconnect
        this._connectCb = onConnect
        this._statusCb = onStatusUpdate
        // Current session represents the session that the rfbClient is currently connected
        // to. This is not always the same as the activeSession as far as the sessionStore
        // is concerned.
        this._currentSession = this._getActiveSession()
        // A socket being used to query a desktop's boot status
        this._statusSocket = null
        // Status text to display to a user when a connection is pending
        this._statusText = ''
        // The RFB client for noVNC connections
        this._rfbClient = null
        // The audio player for streaming playback
        this._audioManager = null
        // Subscribe to changes to desktop sessions
        this._unsubscribeSessions = this._sessionStore.subscribe((mutation) => {
            this._handleSessionChange(mutation)
        })
    }

    // _callDisconnect will call the onDisconnect callback if configured.
    _callDisconnect () { if (this._disconnectCb) { this._disconnectCb() }}

    // _callConnect will call the onConnect callback if configured.
    _callConnect () { if (this._connectCb) { this._connectCb() } }

    // _callStatusUpdate will call the onStatusUpdate callback if configured.
    _callStatusUpdate (st) { if (this._statusCb) { this._statusCb(st) } }

    // _callError will call the onError callback if configured.
    _callError (err) { if (this._errCb) { this._errCb(err) } }

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
        const urls = this._getSessionURLs()
        const audioUrl = urls.audioURL()
        const playerCfg = {
            server: { url: audioUrl },
            onDisconnect: () => { this._resetAudioStatus() },
            onError: (err) => { this._callError(err) }
        }
        this._audioManager = new AudioManager(playerCfg)
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
    _doStatusWebsocket () {

        const urls = this._getSessionURLs()
        const activeSession = this._getActiveSession()
        const socket = new WebSocket(urls.statusURL())

        let msgCount = 0

        socket.onopen = (e) => {
            this._callStatusUpdate(`Starting desktop ${activeSession.namespace}/${activeSession.name}`)
        }
          
        socket.onmessage = (event) => {
            msgCount++

            console.log(`[status] Status update ${msgCount} received from server: ${event.data}`)
            const st = JSON.parse(event.data)

            // If there is an error on the pipe, send it to the user.
            if (st.error) {
                this._currentSession = this._getActiveSession()
                this._callDisconnect()
                this._callError(new Error(st.error))
                return
            }

            // If the desktop is ready then create a connection, clear the status, and close this socket
            if (this._statusIsReady(st)) {
                this._createConnection()
                if (socket.readyState === 1) {
                    socket.close()
                }
                this._callStatusUpdate('')
                return
            }

            // Update the status text for the user
            let statusText = `Waiting for ${activeSession.namespace}/${activeSession.name}`
            if (msgCount > 6) {
                statusText += '\n\nThis is taking a while. The server might be pulling the'
                statusText += '\nimage for the first time, or the control-plane is having'
                statusText += '\ntrouble scheduling the desktop instance.'
            }
            this._callStatusUpdate(statusText)
        }

        socket.onclose = (event) => {
            if (event.wasClean) {
                console.log(`[status] Connection closed cleanly, code=${event.code} reason=${event.reason}`)
            } else {
                this._callError(new Error(`Error getting session status: ${event.code} ${event.reason}`))
            }
        }
          
        socket.onerror = (err) => {
            this._callError(err)
            this._callDisconnect()
        }

        this._statusSocket = socket
    }

    // _createConnection will create a new RFB connection if the socketType
    // is xvnc. xpra sockets use the official client embedded in an iframe.
    async _createConnection () {
        if (this._currentSession.socketType !== 'xvnc') {
            // xpra sockets are handled via an iframe currently
            this._callConnect()
            return
        }
        // get the websocket display address
        const urls = this._getSessionURLs()
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
            this._callDisconnect()
            this._callError(err)
            this._currentSession = this._getActiveSession()
            return
        }
        this._callConnect()
    }

    // _createRFBConnection creates a new RFB connection.
    async _createRFBConnection (view, url) {
        if (this._rfbClient) { return }
        const rfb = new RFB(view, url)
        rfb.addEventListener('connect', () => { this._connectedToRFBServer() })
        rfb.addEventListener('disconnect', (ev) => { this._disconnectedFromRFBServer(ev) })
        rfb.resizeSession = true
        rfb.scaleViewport = true
        this._rfbClient = rfb
    }

    // _connectedToRFBServer is called when the RFB connection is established
    // with the desktop session.
    _connectedToRFBServer() {
        console.log('Connected to display server!')
        const activeSession = this._getActiveSession()
        if (activeSession.socketType === 'xvnc') {
            this._rfbClient.scaleViewport = true
            this._rfbClient.resizeSession = true
        }
    }

    // _disconnectedFromRFBServer is called when the connection is dropped to a
    // desktop session.
    async _disconnectedFromRFBServer (event) {
        if (this._rfbClient) {
            this._rfbClient = null
        }
        this._callDisconnect()

        if (event.detail.clean) {
            // The server disconnecting cleanly would mean expired session,
            // but this should probably be handled better.
            if (this._currentSession) {
                try {
                    // check if the desktop still exists, if we get an error back
                    // it was deleted.
                    await this._sessionStore.getters.sessionStatus(this._currentSession)
                } catch {
                    this._sessionStore.dispatch('deleteSession', this._currentSession)
                    this._currentSession = null
                    this._callError(new Error("The desktop session has ended"))
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
        if (this._rfbClient) {
            try {
                // _disconnectedFromRFBServer will call the disconnect callback
                this._rfbClient.disconnect()
            } catch (err) {
                console.log(err)
            } finally {
                this._rfbClient = null
            }
            return
        }
        if (this._statusSocket) {
            try {
                this._statusSocket.close()
            } catch (err) {
                console.log(err)
            } finally {
                this._statusSocket = null
            }
        }
        this._callDisconnect()
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

    // getConnectingStatus returns the status message for the connecting status
    getConnectingStatus () {
        return this._statusText
    }

    // syncClipboardData syncs the provied data to the clipboard inside the currently
    // active RFB connection.
    syncClipboardData (data) {
        if (!this._rfbClient) {
            return
        }
        const session = this._getActiveSession()
        if (session.socketType !== 'xvnc') {
            return
        }
        this._rfbClient.clipboardPasteFrom(data)
    }

    // xpraArgs returns the xpra args for the given desktop session
    xpraArgs () {
        const urls = this._getSessionURLs()
        return urls.xpraArgs()
    }
}

// DesktopAddressGetter is a convenience wrapper around retrieving connection
// URLs for a given desktop instance.
export class DesktopAddressGetter {
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

    // xpraArgs returns the arguments to pass to the xpra iframe for app-profile
    // desktop sessions.
    xpraArgs () {
      const host = window.location.hostname
      const path = `/api/desktops/ws/${this.namespace}/${this.name}/display?token=${this._getToken()}`
    
      let port = window.location.port
      let secure = false
    
      if (window.location.protocol.replace(/:$/g, '') === 'https') {
        secure = true
        if (port === '') {
          port = 443
        }
      } else {
        if (port === '') {
          port = 80
        }
      }
    
      return { host: host, path: path, port: port, secure: secure }
    }
  }