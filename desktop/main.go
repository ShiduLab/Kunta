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
    "time"
)

//go:embed web web/* web/assets/*
var embedded embed.FS

func firstExisting(paths ...string) string {
    for _, p := range paths {
        if p == "" {
            continue
        }
        if info, err := os.Stat(p); err == nil && !info.IsDir() {
            return p
        }
    }
    return ""
}

func browserExecutable() string {
    pf := os.Getenv("ProgramFiles")
    pf86 := os.Getenv("ProgramFiles(x86)")
    lad := os.Getenv("LOCALAPPDATA")
    return firstExisting(
        filepath.Join(pf86, "Microsoft", "Edge", "Application", "msedge.exe"),
        filepath.Join(pf, "Microsoft", "Edge", "Application", "msedge.exe"),
        filepath.Join(lad, "Microsoft", "Edge", "Application", "msedge.exe"),
        filepath.Join(pf, "Google", "Chrome", "Application", "chrome.exe"),
        filepath.Join(pf86, "Google", "Chrome", "Application", "chrome.exe"),
        filepath.Join(lad, "Google", "Chrome", "Application", "chrome.exe"),
    )
}

func launchDefault(url string) error {
    return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

func main() {
    web, err := fs.Sub(embedded, "web")
    if err != nil {
        return
    }

    listener, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil {
        return
    }

    mux := http.NewServeMux()
    mux.Handle("/", http.FileServer(http.FS(web)))
    server := &http.Server{
        Handler:           mux,
        ReadHeaderTimeout: 5 * time.Second,
    }
    go func() { _ = server.Serve(listener) }()

    url := fmt.Sprintf("http://%s/", listener.Addr().String())
    browser := browserExecutable()
    if browser == "" {
        _ = launchDefault(url)
        select {}
    }

    profile := filepath.Join(os.Getenv("LOCALAPPDATA"), "Kunta", "WebRuntime")
    _ = os.MkdirAll(profile, 0o755)

    cmd := exec.Command(browser,
        "--app="+url,
        "--user-data-dir="+profile,
        "--no-first-run",
        "--disable-default-apps",
    )
    _ = cmd.Run()
    _ = server.Close()
}
