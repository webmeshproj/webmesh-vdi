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

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/audio"
	"github.com/tinyzimmer/kvdi/pkg/audio/pa"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

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

	log.Info("Connection to vnc server established")

	log.Info("Setting up pulse-audio devices")

	paDevices := pa.NewDeviceManager(log.WithName("pa_devices"), strconv.Itoa(userID))

	if err := paDevices.AddSink("kvdi", "kvdi-playback"); err != nil {
		log.Error(err, "Failed to add kvdi-playback device, audio playback may not work as expected")
	}

	if err := paDevices.AddSource("virtmic", "kvdi-microphone", filepath.Join(v1.DesktopRunDir, "mic.fifo"), "s16le", 1, 16000); err != nil {
		log.Error(err, "Failed to add virtmic device, microphone support may not work as expected")
	}

	if err := paDevices.SetDefaultSource("virtmic"); err != nil {
		log.Error(err, "Failed to set virtmic to default source, microphone support may not work as expected")
	}

	defer paDevices.Destroy()

	log.Info("Starting display proxy")

	wsconn.PayloadType = websocket.BinaryFrame

	// wrap the connection so we can log metrics
	watcher := apiutil.NewWebsocketWatcher(wsconn)

	stChan := logWatcherMetrics("display", watcher)
	defer func() { stChan <- struct{}{} }()

	// Copy client connection to the server
	go func() {
		if _, err := io.Copy(vncConn, watcher); err != nil {
			log.Error(err, "Error while copying stream from websocket connection to display socket")
		}
		stChan <- struct{}{}
	}()

	// Copy server connection to the client
	go func() {
		if _, err := io.Copy(watcher, vncConn); err != nil {
			log.Error(err, "Error while copying stream from display socket to websocket connection")
		}
		stChan <- struct{}{}
	}()

	// need a better way to block here
	select {}
}

func wsAudioHandler(wsconn *websocket.Conn) {
	log.Info(fmt.Sprintf("Received audio proxy request, setting up pulseaudio/g-streamer"))

	wsconn.PayloadType = websocket.BinaryFrame

	// Create a new audio buffer
	audioBuffer := audio.NewBuffer(&audio.BufferOpts{
		Logger:           log,
		PulseServer:      fmt.Sprintf("/run/user/%d/pulse/native", userID),
		PulseMonitorName: "kvdi.monitor",
		PulseMicName:     "virtmic",
		PulseMicPath:     filepath.Join(v1.DesktopRunDir, "mic.fifo"),
	})

	// Start the audio buffer
	if err := audioBuffer.Start(); err != nil {
		log.Error(err, "Error setting up audio buffer")
		return
	}

	watcher := apiutil.NewWebsocketWatcher(wsconn)
	stChan := logWatcherMetrics("audio", watcher)
	defer func() { stChan <- struct{}{} }()

	// Make sure GST processes and watchers are dead when the handler returns
	defer func() {
		if !audioBuffer.IsClosed() {
			if err := audioBuffer.Close(); err != nil {
				log.Error(err, "Error closing audio buffer")
			}
		}
		stChan <- struct{}{}
	}()

	// Copy audo playback data to the connection
	go func() {
		if _, err := io.Copy(watcher, audioBuffer); err != nil {
			if !errors.IsBrokenPipeError(err) {
				log.Error(err, "Error while copying from audio stream to websocket connection")
			}
		}
		if !audioBuffer.IsClosed() {
			if err := audioBuffer.Close(); err != nil {
				log.Error(err, "Error closing audio buffer")
			}
		}
	}()

	// Copy any received recording data to the buffer
	go func() {
		if _, err := io.Copy(audioBuffer, watcher); err != nil {
			if !errors.IsBrokenPipeError(err) {
				log.Error(err, "Error while copying from websocket connection to audio buffer")
			}
		}
		if !audioBuffer.IsClosed() {
			if err := audioBuffer.Close(); err != nil {
				log.Error(err, "Error closing audio buffer")
			}
		}
	}()

	// Wait for the audiobuffer to exit
	if err := audioBuffer.Wait(); err != nil {
		log.Info(err.Error())
		if errs := audioBuffer.Errors(); errs != nil {
			log.Error(err, "Errors occured while streaming audio")
			for _, e := range errs {
				log.Info(e.Error())
			}
		}
	}

	// Close the websocket connection
	if err := watcher.Close(); err != nil {
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
