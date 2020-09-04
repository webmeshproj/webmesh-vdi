export class DesktopAddressGetter {
  constructor (userStore, namespace, name) {
    this.userStore = userStore
    this.namespace = namespace
    this.name = name
  }

  _getToken () {
    return this.userStore.getters.token
  }

  _buildAddress (endpoint) {
    return `${window.location.origin.replace('http', 'ws')}/api/desktops/ws/${this.namespace}/${this.name}/${endpoint}?token=${this._getToken()}`
  }

  displayURL () {
    return this._buildAddress('display')
  }

  audioURL () {
    return this._buildAddress('audio')
  }

  statusURL () {
    return this._buildAddress('status')
  }

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

export function iframeRef (frameRef) {
  return frameRef.contentWindow
    ? frameRef.contentWindow.document
    : frameRef.contentDocument
}

// getErrorMessage turns a given error into a human readable string
export async function getErrorMessage (err) {
  console.log(err.code)
  console.log(err.message)
  console.log(err.stack)
  if (err.response !== undefined && err.response.data !== undefined) {
    if (err.response.data.error !== undefined) {
      return err.response.data.error
    }
    try {
      // This might be a json error from trying to download a file
      const text = await err.response.data.text()
      const errData = JSON.parse(text)
      if (errData.error !== undefined) {
        return errData.error
      }
    } catch {}
  }

  if (err.message) {
    return err.message
  }

  return err
}
