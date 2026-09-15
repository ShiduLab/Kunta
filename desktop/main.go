//go:build windows

package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

//go:embed web web/* web/assets/*
var embedded embed.FS

var lastHeartbeat atomic.Int64
var pageConnected atomic.Bool

func messageBox(title, text string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("MessageBoxW")
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	textPtr, _ := syscall.UTF16PtrFromString(text)
	_, _, _ = proc.Call(
		0,
		uintptr(unsafe.Pointer(textPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		0x00000010,
	)
}

func findOpera() string {
	roots := []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Opera"),
		filepath.Join(os.Getenv("ProgramFiles"), "Opera"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Opera"),
	}

	var found []string
	for _, root := range roots {
		if root == "" {
			continue
		}
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d == nil || d.IsDir() {
				return nil
			}
			if strings.EqualFold(d.Name(), "opera.exe") {
				found = append(found, path)
			}
			return nil
		})
	}

	if len(found) == 0 {
		return ""
	}

	sort.Slice(found, func(i, j int) bool {
		ii, ierr := os.Stat(found[i])
		jj, jerr := os.Stat(found[j])
		if ierr == nil && jerr == nil && !ii.ModTime().Equal(jj.ModTime()) {
			return ii.ModTime().After(jj.ModTime())
		}
		return found[i] > found[j]
	})
	return found[0]
}

func serveIndex(w http.ResponseWriter, web fs.FS) {
	data, err := fs.ReadFile(web, "index.html")
	if err != nil {
		http.Error(w, "index.html non disponibile", http.StatusInternalServerError)
		return
	}

	heartbeat := `<script>
(function(){
  function ping(){
    fetch('/__kunta_alive',{method:'POST',cache:'no-store',keepalive:true}).catch(function(){});
  }
  ping();
  setInterval(ping,2000);
})();
</script>`

	html := string(data)
	if strings.Contains(html, "</body>") {
		html = strings.Replace(html, "</body>", heartbeat+"\n</body>", 1)
	} else {
		html += heartbeat
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(html))
}

func main() {
	web, err := fs.Sub(embedded, "web")
	if err != nil {
		messageBox("Kunta", "Impossibile aprire le risorse interne di Kunta.")
		return
	}

	opera := findOpera()
	if opera == "" {
		messageBox("Kunta", "Opera non è stato trovato. Kunta Desktop usa esclusivamente Opera.")
		return
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		messageBox("Kunta", "Impossibile avviare il server locale di Kunta.")
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/__kunta_alive", func(w http.ResponseWriter, r *http.Request) {
		pageConnected.Store(true)
		lastHeartbeat.Store(time.Now().UnixNano())
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			serveIndex(w, web)
			return
		}
		http.FileServer(http.FS(web)).ServeHTTP(w, r)
	})

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() { _ = server.Serve(listener) }()

	url := fmt.Sprintf("http://%s/", listener.Addr().String())
	profile := filepath.Join(os.Getenv("LOCALAPPDATA"), "Kunta", "OperaProfile")
	_ = os.MkdirAll(profile, 0o755)

	cmd := exec.Command(
		opera,
		"--app="+url,
		"--user-data-dir="+profile,
		"--no-first-run",
		"--disable-default-apps",
	)
	if err := cmd.Start(); err != nil {
		_ = server.Close()
		messageBox("Kunta", "Opera non può essere avviato.")
		return
	}

	startupDeadline := time.Now().Add(30 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !pageConnected.Load() {
			if time.Now().After(startupDeadline) {
				_ = server.Close()
				messageBox("Kunta", "Opera non ha aperto Kunta entro 30 secondi.")
				return
			}
			continue
		}

		last := time.Unix(0, lastHeartbeat.Load())
		if time.Since(last) > 8*time.Second {
			_ = server.Close()
			return
		}
	}
}
