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

export function getWebsockifyAddr (namespace, name, token) {
  return `${window.location.origin.replace('http', 'ws')}/api/desktops/ws/${namespace}/${name}/display?token=${token}`
}

export function getXpraServerArgs (namespace, name, token) {
  const host = window.location.hostname
  const path = `/api/desktops/ws/${namespace}/${name}/display?token=${token}`

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

export function getWebsockifyAudioAddr (namespace, name, token) {
  return `${window.location.origin.replace('http', 'ws')}/api/desktops/ws/${namespace}/${name}/audio?token=${token}`
}

export function iframeRef (frameRef) {
  return frameRef.contentWindow
    ? frameRef.contentWindow.document
    : frameRef.contentDocument
}
