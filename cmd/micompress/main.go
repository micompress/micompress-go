// Command micompress compresses images, PDFs, video, audio, documents and archives
// via micompress.com.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	version = "0.1.2"
	base    = "https://micompress.com"
	ua      = "micompress-go/" + version + " (+https://micompress.com)"
	// Explicit channel id for server-side attribution (see extend/README.md).
	clientID   = "go/" + version
	accountURL = "https://micompress.com/account"
	plansURL   = "https://micompress.com/plans"
)

var extCategory = map[string]string{
	"jpg": "image", "jpeg": "image", "png": "image", "webp": "image", "gif": "image",
	"bmp": "image", "tiff": "image", "tif": "image", "avif": "image", "svg": "image",
	"heic": "image", "heif": "image", "ico": "image", "psd": "image", "dng": "image",
	"raw": "image", "cr2": "image", "arw": "image", "nef": "image",
	"mp4": "video", "mov": "video", "avi": "video", "mkv": "video", "webm": "video",
	"mpeg": "video", "mpg": "video", "m4v": "video", "wmv": "video", "flv": "video", "3gp": "video",
	"mp3": "audio", "wav": "audio", "flac": "audio", "ogg": "audio", "m4a": "audio",
	"aac": "audio", "wma": "audio", "opus": "audio",
	"pdf": "pdf",
	"csv": "document", "json": "document", "xml": "document", "md": "document",
	"odt": "document", "ods": "document", "docx": "document", "doc": "document",
	"ppt": "document", "pptx": "document", "xls": "document", "xlsx": "document",
	"epub": "document", "html": "document", "htm": "document", "hwp": "document",
	"txt": "document", "rtf": "document",
	"zip": "archive", "7z": "archive", "tar": "archive", "gz": "archive", "tgz": "archive", "rar": "archive",
}

var categoryPath = map[string]string{
	"image": "image-compressors", "video": "video-compressors", "audio": "audio-compressors",
	"document": "document-compressors", "pdf": "pdf-compressors", "archive": "archive-compressors",
}

type meta struct {
	Success      bool   `json:"success"`
	Code         string `json:"code"`
	Error        string `json:"error"`
	DownloadURL  string `json:"downloadUrl"`
	InputSize    int64  `json:"inputSize"`
	OutputSize   int64  `json:"outputSize"`
	SavedPercent int64  `json:"savedPercent"`
}

func endpointFor(name string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
	cat, ok := extCategory[ext]
	if !ok {
		return ""
	}
	return categoryPath[cat]
}

func human(n int64) string {
	units := []string{"B", "KB", "MB", "GB"}
	v := float64(n)
	i := 0
	for v >= 1024 && i < len(units)-1 {
		v /= 1024
		i++
	}
	if i == 0 {
		return fmt.Sprintf("%d%s", int64(v+0.5), units[i])
	}
	return fmt.Sprintf("%.1f%s", v, units[i])
}

type options struct {
	preset, apiBase, suffix, output, token, apiKey string
	quiet                                          bool
}

// describeError turns a failed API response into a helpful message
// (adds hints for auth/quota error codes).
func describeError(status int, m meta) error {
	msg := m.Error
	if msg == "" {
		msg = fmt.Sprintf("HTTP %d", status)
	}
	// Only append a link when the server text doesn't already include it.
	switch m.Code {
	case "LOGIN_REQUIRED":
		return fmt.Errorf("file exceeds the anonymous size limit. Pass --api-key <key> or set MICOMPRESS_API_KEY (free key at %s)", accountURL)
	case "INVALID_API_KEY":
		if !strings.Contains(msg, accountURL) {
			return fmt.Errorf("%s Check your key at %s", msg, accountURL)
		}
	case "API_QUOTA_REACHED", "USAGE_LIMIT_REACHED":
		if !strings.Contains(msg, plansURL) {
			return fmt.Errorf("%s See plans and limits at %s", msg, plansURL)
		}
	}
	return fmt.Errorf("%s", msg)
}

func compress(client *http.Client, path string, o options) ([]byte, meta, error) {
	var m meta
	name := filepath.Base(path)
	endpoint := endpointFor(name)
	if endpoint == "" {
		return nil, m, fmt.Errorf("unsupported file type")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, m, err
	}
	defer f.Close()

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fw, err := w.CreateFormFile("file", name)
	if err != nil {
		return nil, m, err
	}
	if _, err := io.Copy(fw, f); err != nil {
		return nil, m, err
	}
	_ = w.WriteField("preset", o.preset)
	w.Close()

	url := fmt.Sprintf("%s/%s/compress", strings.TrimRight(o.apiBase, "/"), endpoint)
	req, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		return nil, m, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", ua)
	req.Header.Set("X-MiCompress-Client", clientID)
	if o.apiKey != "" {
		req.Header.Set("X-API-Key", o.apiKey)
	}
	if o.token != "" {
		req.Header.Set("Authorization", "Bearer "+o.token)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, m, err
	}
	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		return nil, m, fmt.Errorf("unexpected response (HTTP %d)", res.StatusCode)
	}
	if !m.Success || m.DownloadURL == "" {
		return nil, m, describeError(res.StatusCode, m)
	}

	dreq, _ := http.NewRequest(http.MethodGet, strings.TrimRight(o.apiBase, "/")+m.DownloadURL, nil)
	dreq.Header.Set("User-Agent", ua)
	if o.token != "" {
		dreq.Header.Set("Authorization", "Bearer "+o.token)
	}
	dres, err := client.Do(dreq)
	if err != nil {
		return nil, m, err
	}
	defer dres.Body.Close()
	if dres.StatusCode != http.StatusOK {
		return nil, m, fmt.Errorf("download failed (HTTP %d)", dres.StatusCode)
	}
	data, err := io.ReadAll(dres.Body)
	return data, m, err
}

func outPath(input string, o options) string {
	name := filepath.Base(input)
	if o.output != "" {
		_ = os.MkdirAll(o.output, 0o755)
		return filepath.Join(o.output, name)
	}
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	return filepath.Join(filepath.Dir(input), stem+o.suffix+ext)
}

func main() {
	o := options{preset: "recommended", apiBase: base, suffix: "-min"}
	var files []string

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-v", "--version":
			fmt.Println(version)
			return
		case "-h", "--help":
			fmt.Println("micompress <file...> [options]")
			fmt.Println("  -o, --output DIR     Output directory")
			fmt.Println("  -p, --preset NAME    lossless|recommended|extreme")
			fmt.Println("      --api-key KEY    Personal API key from micompress.com/account")
			fmt.Println("                       (or set MICOMPRESS_API_KEY)")
			fmt.Println("      --token JWT      Legacy login token (still supported)")
			fmt.Println("      --suffix STR     Suffix for output files (default -min)")
			fmt.Println("      --api URL        API base URL")
			fmt.Println("  -q, --quiet          Only print errors")
			fmt.Println("Plans and limits: " + plansURL)
			fmt.Println("Powered by micompress.com")
			return
		case "-q", "--quiet":
			o.quiet = true
		case "-o", "--output":
			i++
			if i < len(args) {
				o.output = args[i]
			}
		case "-p", "--preset":
			i++
			if i < len(args) {
				o.preset = args[i]
			}
		case "--api":
			i++
			if i < len(args) {
				o.apiBase = args[i]
			}
		case "--token":
			i++
			if i < len(args) {
				o.token = args[i]
			}
		case "--api-key":
			i++
			if i < len(args) {
				o.apiKey = args[i]
			}
		case "--suffix":
			i++
			if i < len(args) {
				o.suffix = args[i]
			}
		default:
			files = append(files, args[i])
		}
	}

	// --api-key wins over the MICOMPRESS_API_KEY environment variable.
	if o.apiKey == "" {
		o.apiKey = os.Getenv("MICOMPRESS_API_KEY")
	}
	o.apiKey = strings.TrimSpace(o.apiKey)

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: micompress <file...> [options]")
		os.Exit(1)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	failed := 0
	for _, file := range files {
		data, m, err := compress(client, file, o)
		if err != nil {
			failed++
			fmt.Fprintf(os.Stderr, "✗ %s: %v\n", file, err)
			continue
		}
		dest := outPath(file, o)
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			failed++
			fmt.Fprintf(os.Stderr, "✗ %s: %v\n", file, err)
			continue
		}
		if !o.quiet {
			fmt.Printf("✔ %s → %s  %s → %s (-%d%%)\n", file, dest, human(m.InputSize), human(m.OutputSize), m.SavedPercent)
		}
	}
	os.Exit(map[bool]int{true: 0, false: 1}[failed == 0])
}
