import Recorder from 'opus-recorder'
import Websock from '@novnc/novnc/core/websock.js'
import encoderPath from 'opus-recorder/dist/encoderWorker.min.js'

// AudioManager is an object for managing audio playback and recording
// to/from a desktop session.
export default class AudioManager {

  constructor (config) {
    this._config = config
    this._socket = null
    this._mediaRecorder = null
    this._context = null
  }

  // _connect will create a new Websocket connection. Use novnc/websockify-js 
  // for more control over the recv and send queues.
  _connect () {
    this._socket = new Websock()
    this._socket.open(this._config.server.url)
    this._socket.binaryType = 'arraybuffer'
    this._socket.on('close', () => { 
      this.stopRecording()
      if (this._config.onDisconnect) {
        this._config.onDisconnect()
      }
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

  // _startMicrophone (stream) {
  //   if (!this._socket) {
  //     this._connect()
  //   }
  //   this._context = new AudioContext()
  //   const source = this._context.createMediaStreamSource(stream)
  //   const processor = this._context.createScriptProcessor(256, 1, 1)

  //   source.connect(processor)
  //   processor.connect(this._context.destination)

  //   processor.onaudioprocess = (e) => {
  //     this._send(e.inputBuffer.getChannelData(0))
  //   }
  // }

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
        if (this._config.onError) {
          this._config.onError(serr)
        }
      })
  }

  // stopRecording stops the audio recorder process
  async stopRecording () {
    if (this._context) {
      // this._mediaRecorder.close()
      this._mediaRecorder.stop()
      this._mediaRecorder = null
      // await this._context.close()
      // this._context = null
    }
  }

  // // This implementation uses raw PCM data.
  // startRecording () {
  //   if (!navigator.getUserMedia) {
  //     navigator.getUserMedia = navigator.getUserMedia || navigator.webkitGetUserMedia ||
  //                 navigator.mozGetUserMedia || navigator.msGetUserMedia
  //   }
  //   if (!navigator.getUserMedia) {
  //     throw new Error('Audio recording is not supported in this browser')
  //   }
  //   // will prompt the user for microphone access if not already provided
  //   navigator.getUserMedia({ audio:true }, 
  //     (stream) => {
  //         this._startMicrophone(stream)
  //     },
  //     (err) => {
  //       throw err
  //     }
  //   )
  //   return
  // }

  // startPlayback starts the playback process
  startPlayback () {

    if (!this._socket) {
      this._connect()
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
