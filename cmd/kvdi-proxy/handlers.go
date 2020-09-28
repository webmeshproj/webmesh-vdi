package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/audio"
	"github.com/tinyzimmer/kvdi/pkg/audio/pa"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"

	"golang.org/x/net/websocket"
)

var (
	monitorDeviceName    = "kvdi"
	monitorDeviceMonitor = "kvdi.monitor"
	monitorDescription   = "kvdi-playback"
	micDeviceName        = "virtmixc"
	micDeviceDescription = "kvdi-microphone"
	micDevicePath        = filepath.Join(v1.DesktopRunDir, "mic.fifo")
	micDeviceFormat      = "s16le"
	micDeviceChannels    = 1
	micDeviceSampleRate  = 16000
)

func wsHandshake(*websocket.Config, *http.Request) error { return nil }

func getPulseServer() string { return fmt.Sprintf("/run/user/%d/pulse/native", userID) }

func setupPulseAudio(manager *pa.DeviceManager) error {
	if err := manager.WaitForReady(time.Second * 2); err != nil {
		return err
	}
	if _, err := manager.AddSink(monitorDeviceName, monitorDescription); err != nil {
		return err
	}

	if _, err := manager.AddSource(&pa.SourceOpts{
		Name:         micDeviceName,
		Description:  micDeviceDescription,
		FifoPath:     micDevicePath,
		SampleFormat: micDeviceFormat,
		SampleRate:   micDeviceSampleRate,
		Channels:     micDeviceChannels,
	}); err != nil {
		return err
	}

	if err := manager.SetDefaultSource(micDeviceName); err != nil {
		return err
	}
	return nil
}

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

	paDevices, err := pa.NewDeviceManager(&pa.DeviceManagerOpts{
		PulseServer: getPulseServer(),
	})
	if err != nil {
		log.Error(err, "Failed to create new PA device manager, audio will be disabled")
	}

	if paDevices != nil {
		if err := setupPulseAudio(paDevices); err != nil {
			if derr := paDevices.Destroy(); derr != nil {
				log.Error(derr, "Failed to cleanup device manager")
			}
			log.Error(err, "Failure while setting up pulse audio, audio will be disabled")
		} else {
			defer func() {
				if derr := paDevices.Destroy(); derr != nil {
					log.Error(derr, "Failed to cleanup device manager")
				}
			}()
		}
	}

	log.Info("Starting display proxy")

	wsconn.PayloadType = websocket.BinaryFrame

	// wrap the connection so we can log metrics
	watcher := apiutil.NewWebsocketWatcher(wsconn)

	stChan := logWatcherMetrics("display", watcher)
	defer func() { stChan <- struct{}{} }()

	ctx, cancel := context.WithCancel(context.Background())

	// Copy client connection to the server
	go func() {
		if _, err := io.Copy(vncConn, watcher); err != nil {
			log.Error(err, "Error while copying stream from websocket connection to display socket")
		}
		cancel()
	}()

	// Copy server connection to the client
	go func() {
		if _, err := io.Copy(watcher, vncConn); err != nil {
			log.Error(err, "Error while copying stream from display socket to websocket connection")
		}
		cancel()
	}()

	// block until the context is finished
	for range ctx.Done() {
	}
}

func wsAudioHandler(wsconn *websocket.Conn) {
	log.Info(fmt.Sprintf("Received audio proxy request, setting up pulseaudio/g-streamer"))

	wsconn.PayloadType = websocket.BinaryFrame

	// Create a new audio buffer
	audioBuffer := audio.NewBuffer(&audio.BufferOpts{
		Logger:           log,
		PulseServer:      getPulseServer(),
		PulseMonitorName: monitorDeviceMonitor,
		PulseMicName:     micDeviceName,
		PulseMicPath:     micDevicePath,
	})

	// Start the audio buffer
	if err := audioBuffer.Start(); err != nil {
		log.Error(err, "Error setting up audio buffer")
		return
	}

	watcher := apiutil.NewWebsocketWatcher(wsconn)
	stChan := logWatcherMetrics("audio", watcher)
	defer func() { stChan <- struct{}{} }()

	// Copy audio playback data to the connection
	go func() {
		if _, err := io.Copy(watcher, audioBuffer); err != nil {
			if !errors.IsBrokenPipeError(err) {
				log.Error(err, "Error while copying from audio stream to websocket connection")
			}
		}
		audioBuffer.Close()
	}()

	// Copy any received recording data to the buffer
	go func() {
		if _, err := io.Copy(audioBuffer, watcher); err != nil {
			if !errors.IsBrokenPipeError(err) {
				log.Error(err, "Error while copying from websocket connection to audio buffer")
			}
		}
	}()

	// Wait for the audiobuffer to exit
	audioBuffer.Wait()

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
