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
