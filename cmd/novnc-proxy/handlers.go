package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/audio"
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

	audioBuffer := audio.NewBuffer(log, userID)

	if err := audioBuffer.Start(audio.CodecOpus); err != nil {
		log.Error(err, "Error setting up audio buffer")
		return
	}

	defer func() {
		if err := audioBuffer.Close(); err != nil {
			log.Error(err, "Error closing audio buffer")
		}
	}()

	go func() {

		if _, err := io.Copy(wsconn, audioBuffer); err != nil {
			log.Error(err, "Error while copying from audio stream to websocket connection")
		}
	}()

	if err := audioBuffer.Wait(); err != nil {
		if err := audioBuffer.Error(); err != nil {
			log.Error(err, "Error while streaming audio")
			if _, err := wsconn.Write(append([]byte(err.Error()), []byte("\n")...)); err != nil {
				log.Error(err, "Failed to write error to websocket client")
			}
		}
	}

	if err := wsconn.Close(); err != nil {
		log.Error(err, "Error closing websocket connection")
	}

	log.Info("Finishing proxying audio stream")
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

func downloadFileHandler(w http.ResponseWriter, r *http.Request) {
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
		apiutil.ReturnAPIError(errors.New("Directory download is not yet supported"), w)
		return
	}

	// Open the file
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
	ctType := http.DetectContentType(hdr)

	// Get the file size
	fSize := strconv.FormatInt(finfo.Size(), 10) // Get file size as a string

	w.Header().Set("Content-Disposition", "attachment; filename="+finfo.Name())
	w.Header().Set("Content-Type", ctType)
	w.Header().Set("Content-Length", fSize)

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
