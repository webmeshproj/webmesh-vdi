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

import Recorder from 'opus-recorder'
import Websock from '@novnc/novnc/core/websock.js'
import { Emitter, Events } from './events.js'
import encoderPath from 'opus-recorder/dist/encoderWorker.min.js'

// AudioManager is an object for managing audio playback and recording
// to/from a desktop session.
export default class AudioManager extends Emitter {

  constructor ({ addressGetter, userStore }) {
    super()
    this._addressGetter = addressGetter
    this._userStore = userStore

    this._socket = null
    this._mediaRecorder = null
  }

  // _connect will create a new Websocket connection. Use novnc/websockify-js 
  // for more control over the recv and send queues.
  async _connect (retry) {
    this._socket = new Websock()
    this._socket.open(this._addressGetter.audioURL())
    this._socket.binaryType = 'arraybuffer'
    this._socket.on('close', (event) => {
      if (!event.wasClean && (event.code === 1006 && !retry)) {
        this._userStore.refreshToken()
          .then(() => {
            this._connect(true)
          })
          .catch((err) => {
            this.emit(Events.error, err)
            throw err
          })
        return
      }
      this.stopRecording()
      if (!event.wasClean || (event.code !== 1000 && event.code !== 1005)) {
        this.emit(Events.error, new Error(`Unexpected message from websocket: ${event.code} ${event.reason}`))
      }
      this.emit(Events.disconnected)
      this._socket = null
    })
  }

  // _disconnect checks if there is a socket and closes it.
  _disconnect () {
    if (this._socket) {
      try {
        this._socket.close()
      } finally {
        this._socket = null
      }
    }
  }

  // _send will send the given data over the websocket connection.
  _send (data) {
    if (this._socket) {
      this._socket.send(data)
    }
  }

  // startRecording starts the recording process.
  startRecording () {
    // build a config for the OpusRecorder
    const config = {
      encoderPath: encoderPath,
      encoderApplication: 2048,
      streamPages: true  // Receive every frame in real time
    }

    // create the media recorder
    this._mediaRecorder = new Recorder(config)
    this._mediaRecorder.ondataavailable = (data) => {
      this._send(data)
    }

    // start the media recorder.
    this._mediaRecorder.start()
      .then(() => { console.log('Started audio recorder')})
      .catch((err) => {
        const serr = new Error(`Failed to start audio recording: ${err}`)
        this.emit(Events.error, serr)
      })
  }

  // stopRecording stops the audio recorder process
  async stopRecording () {
    if (this._mediaRecorder) {
      this._mediaRecorder.stop()
      this._mediaRecorder = null
    }
  }

  // startPlayback starts the playback process
  async startPlayback () {

    if (!this._socket) {
      await this._connect()
    }

    // Create a new MediaSource and tie it to a fake audio object
    var mediaSource = new MediaSource()
    var buffer
    var queue = []
    var audio = document.createElement('audio');
    audio.src = window.URL.createObjectURL(mediaSource);

    mediaSource.addEventListener('sourceopen', function(e) {
      // Start the audio player
      audio.play();

      // Currently assumes proxy instance is sending webm/opus data, this should be configurable
      // and discoverable. 
      buffer = mediaSource.addSourceBuffer('audio/webm; codecs="opus"')

      // get some verbosity in the console on errors
      buffer.addEventListener('error', function(e) { console.log('error: ' + mediaSource.readyState) })
      buffer.addEventListener('abort', function(e) { console.log('abort: ' + mediaSource.readyState) })

      buffer.addEventListener('update', function() { // Note: Have tried 'updateend'
          // If there is an item in the queue and the buffer is done updating
          // pass the next audio segment.
          if (queue.length > 0 && !buffer.updating) {
              buffer.appendBuffer(queue.shift())
          }
      })
    }, false)

    this._socket.on('message', () => {
      // Pull all data out of the receive buffer and convert to Uint8Array
      const data = this._socket.rQshiftBytes(this._socket.rQlen)
      const buf = new Uint8Array(data)
      // If the buffer is updating or the queue is non-empty, queue the audio segment
      if (buffer.updating || queue.length > 0) {
          queue.push(buf)
      } else {
          // Otherwise, place the audio segment directly in the buffer
          buffer.appendBuffer(buf)
      }
    })

  }

  // close closes the websocket connection.
  close () {
    this._disconnect()
  }

}
