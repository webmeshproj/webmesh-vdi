
export default class {

  constructor (config) {
    this.config = config
    this.socket = null
  }

  _connect () {
    this.socket = new WebSocket(this.config.server.url)
    this.socket.binaryType = 'arraybuffer'
  }

  _addListener (f) {
    this.socket.addEventListener('message', f, false)
  }

  startRecording () {
    return
  }

  startPlayback () {
    if (!this.socket) {
      this._connect()
    }

    var mediaSource = new MediaSource()
    var buffer
    var queue = []

    var audio = document.createElement('audio');
    audio.src = window.URL.createObjectURL(mediaSource);

    mediaSource.addEventListener('sourceopen', function(e) {
      audio.play();

      buffer = mediaSource.addSourceBuffer('audio/webm; codecs="opus"')

      buffer.addEventListener('error', function(e) { console.log('error: ' + mediaSource.readyState) })
      buffer.addEventListener('abort', function(e) { console.log('abort: ' + mediaSource.readyState) })

      buffer.addEventListener('update', function() { // Note: Have tried 'updateend'
          if (queue.length > 0 && !buffer.updating) {
              buffer.appendBuffer(queue.shift())
          }
      })
    }, false)

    this._addListener((e) => {
      if (typeof e.data !== 'string') {
          if (buffer.updating || queue.length > 0) {
              queue.push(e.data)
          } else {
              buffer.appendBuffer(e.data)
          }
       }
    })

  }

  close () {
    if (this.socket) {
      this.socket.close()
    }
  }

}
