package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/audio"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
)

func wsHandshake(*websocket.Config, *http.Request) error { return nil }

func websockifyHandler(wsconn *websocket.Conn) {
	log.Info(fmt.Sprintf("Received display proxy request, connecting to %s", vncAddr))
	vncConn, err := net.Dial(vncConnectProto, vncConnectAddr)

	if err != nil {
		log.Error(err, "Failed to connect to display server")
		wsconn.Close()
		return
	}

	log.Info("Connection established, proxying display")

	wsconn.PayloadType = websocket.BinaryFrame

	// Copy client connection to the server
	go func() {
		if _, err := io.Copy(vncConn, wsconn); err != nil {
			log.Error(err, "Error while copying stream from websocket connection to display socket")
		}
	}()

	// Copy server connection to the client
	go func() {
		if _, err := io.Copy(wsconn, vncConn); err != nil {
			log.Error(err, "Error while copying stream from display socket to websocket connection")
		}
	}()

	select {}
}

func wsAudioHandler(wsconn *websocket.Conn) {
	log.Info(fmt.Sprintf("Received validated proxy request, connecting to audio stream"))

	wsconn.PayloadType = websocket.BinaryFrame

	audioBuffer := audio.NewBuffer(log, strconv.Itoa(userID))

	// Start the audio buffer
	if err := audioBuffer.Start(audio.CodecOpus); err != nil {
		log.Error(err, "Error setting up audio buffer")
		return
	}

	// Make double sure GST processes are dead when the handler returns
	defer func() {
		if !audioBuffer.IsClosed() {
			if err := audioBuffer.Close(); err != nil {
				log.Error(err, "Error closing audio buffer")
			}
		}
	}()

	// Copy audo playback data to the connection
	go func() {
		if _, err := io.Copy(wsconn, audioBuffer); err != nil {
			if !audioBuffer.IsClosed() {
				if cerr := audioBuffer.Close(); cerr != nil {
					log.Error(cerr, "Error closing audio buffer")
				}
			}
			if !errors.IsBrokenPipeError(err) {
				log.Error(err, "Error while copying from audio stream to websocket connection")
			}
		}
	}()

	// Copy any received recording data to the buffer
	go func() {
		if _, err := io.Copy(audioBuffer, wsconn); err != nil {
			if !audioBuffer.IsClosed() {
				if cerr := audioBuffer.Close(); cerr != nil {
					log.Error(cerr, "Error closing audio buffer")
				}
			}
			if !errors.IsBrokenPipeError(err) {
				log.Error(err, "Error while copying from websocket connection to audio buffer")
			}
		}
	}()

	// Wait for the audiobuffer to exit
	if err := audioBuffer.Wait(); err != nil {
		if errs := audioBuffer.Errors(); errs != nil {
			log.Error(err, "Errors occured while streaming audio")
			for _, e := range errs {
				log.Info(e.Error())
			}
		}
	}

	// Close the websocket connection
	if err := wsconn.Close(); err != nil {
		if !errors.IsBrokenPipeError(err) {
			log.Error(err, "Error closing websocket connection")
		}
	}

	log.Info("Audio stream proxy ended")
}

func statFileHandler(w http.ResponseWriter, r *http.Request) {
	path, err := getLocalPathFromRequest(r)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	finfo, err := os.Stat(path)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	resp := &v1.StatDesktopFileResponse{
		Stat: &v1.FileStat{
			Name:        finfo.Name(),
			IsDirectory: finfo.IsDir(),
		},
	}
	if finfo.IsDir() {
		resp.Stat.Contents = make([]*v1.FileStat, 0)
		files, err := ioutil.ReadDir(path)
		if err != nil {
			apiutil.ReturnAPIError(err, w)
			return
		}
		for _, file := range files {
			resp.Stat.Contents = append(resp.Stat.Contents, &v1.FileStat{
				Name:        file.Name(),
				IsDirectory: file.IsDir(),
				Size:        file.Size(),
			})
		}

	} else {
		resp.Stat.Size = finfo.Size()
	}

	apiutil.WriteJSON(resp, w)
}

func serveFile(w http.ResponseWriter, path string) {
	// Stat the file
	finfo, err := os.Stat(path)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	f, err := os.Open(path)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer f.Close()

	// Get the file header
	hdr := make([]byte, 512)
	if _, err := f.Read(hdr); err != nil {
		apiutil.ReturnAPIError(errors.New("Failed to read header from file"), w)
		return
	}

	// Get content type of file
	contentType := http.DetectContentType(hdr)

	// Get the file size
	fileSizeStr := strconv.FormatInt(finfo.Size(), 10) // Get file size as a string

	w.Header().Set("Content-Length", fileSizeStr)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename="+finfo.Name())
	w.Header().Set("X-Suggested-Filename", finfo.Name())
	w.Header().Set("X-Decompressed-Content-Length", fileSizeStr)

	w.WriteHeader(http.StatusOK)
	// Seek back to the start of the file (since we read the header already)
	if _, err := f.Seek(0, 0); err != nil {
		apiutil.ReturnAPIError(errors.New("Failed to seek to beginning of file"), w)
		return
	}

	// Copy the file contents to the response
	if _, err := io.Copy(w, f); err != nil {
		log.Error(err, "Failed to copy file contents to response buffer")
	}
}

func downloadDir(w http.ResponseWriter, path string) {
	tarball, err := common.TarDirectoryToTempFile(path)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	serveFile(w, tarball)
}

func downloadFileHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	path, err := getLocalPathFromRequest(r)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	// Stat the file
	finfo, err := os.Stat(path)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	if finfo.IsDir() {
		downloadDir(w, path)
		return
	}

	serveFile(w, path)
}

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	uploadDir := filepath.Join(v1.DesktopHomeMntPath, "Uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	if err := os.Chown(uploadDir, userID, userID); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer file.Close()
	dstFile := filepath.Join(uploadDir, handler.Filename)

	f, err := os.OpenFile(dstFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer f.Close()

	if _, err := io.Copy(f, file); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	if err := os.Chown(dstFile, userID, userID); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	apiutil.WriteOK(w)
}

func getLocalPathFromRequest(r *http.Request) (path string, err error) {
	if _, err := os.Stat(v1.DesktopHomeMntPath); err != nil {
		return "", errors.New("File transfer is disabled for this desktop session")
	}

	// build out the path prefix to strip from the request URL
	pathPrefix := apiutil.GetGorillaPath(r)
	pathPrefix = strings.Replace(pathPrefix, "{name}", mux.Vars(r)["name"], 1)
	pathPrefix = strings.Replace(pathPrefix, "{namespace}", mux.Vars(r)["namespace"], 1)

	fPath := filepath.Join(v1.DesktopHomeMntPath, strings.TrimPrefix(r.URL.Path, pathPrefix))
	absPath, err := filepath.Abs(fPath)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(absPath, v1.DesktopHomeMntPath) {
		// requestor tried to traverse outside the user's home directory (into proxy root fs)
		return "", fmt.Errorf("%s is outside the user's home directory", fPath)
	}
	return absPath, nil
}
