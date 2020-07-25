
export default class {

  constructor (config) {
    this.config = config
  }

  start () {    
    var socket = new WebSocket(this.config.server.url)
    socket.binaryType = 'arraybuffer'

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


    mediaSource.addEventListener('sourceopen', function(e) { console.log('sourceopen: ' + mediaSource.readyState) })
    mediaSource.addEventListener('sourceended', function(e) { console.log('sourceended: ' + mediaSource.readyState) })
    mediaSource.addEventListener('sourceclose', function(e) { console.log('sourceclose: ' + mediaSource.readyState) })
    mediaSource.addEventListener('error', function(e) { console.log('error: ' + mediaSource.readyState) })


    socket.addEventListener('message', function (e) {
      if (typeof e.data !== 'string') {
          if (buffer.updating || queue.length > 0) {
              queue.push(e.data)
          } else {
              buffer.appendBuffer(e.data)
          }
       }
    }, false)

    this.stop = function () {
      socket.close()
    }
  }

}
