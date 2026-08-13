package render

import (
	"bytes"
	"encoding/base64"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const (
	envImages    = "STELL_TUI_IMAGES"
	maxImageB    = 512 << 10
	maxImagePx   = 2048 * 2048
	maxImageSide = 4096
	chunkKitty   = 4096
	maxImgMemo   = 16
)

var (
	mdImage  = regexp.MustCompile(`!\[[^\]]*]\(([^)]+)\)`)
	pathLike = regexp.MustCompile(`(?i)(?:^|\s)((?:\.?/)?[\w./-]+\.(?:png|jpe?g|gif))\b`)

	imgMu   sync.Mutex
	imgMemo = map[string]string{}
)

// Protocol is an inline-image encoding.
type Protocol string

const (
	ProtocolNone  Protocol = ""
	ProtocolKitty Protocol = "kitty"
	ProtocolITerm Protocol = "iterm"
)

// DetectImageProtocol inspects env. Empty means placeholder-only (safe on xterm).
func DetectImageProtocol(getenv func(string) string) Protocol {
	if getenv == nil {
		getenv = os.Getenv
	}
	if v := getenv(envImages); v == "0" || v == "false" {
		return ProtocolNone
	}
	if getenv("KITTY_WINDOW_ID") != "" || strings.Contains(strings.ToLower(getenv("TERM")), "kitty") {
		return ProtocolKitty
	}
	if getenv("GHOSTTY_RESOURCES_DIR") != "" {
		return ProtocolKitty
	}
	switch getenv("TERM_PROGRAM") {
	case "iTerm.app", "WezTerm":
		return ProtocolITerm
	case "ghostty", "kitty":
		return ProtocolKitty
	}
	return ProtocolNone
}

// InlineImages scans text for image paths (workspace-jail). Empty root → placeholders only.
func InlineImages(text string, cols int, root string) string {
	return InlineImagesEnv(text, cols, os.Getenv, root)
}

// InlineImagesEnv is Detect-injectable (tests).
func InlineImagesEnv(text string, cols int, getenv func(string) string, root string) string {
	paths := imagePaths(text)
	if len(paths) == 0 {
		return ""
	}
	proto := DetectImageProtocol(getenv)
	var b strings.Builder
	for i, p := range paths {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(encodeImage(root, p, cols, proto))
	}
	return b.String()
}

func imagePaths(text string) []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" || strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	for _, m := range mdImage.FindAllStringSubmatch(text, 4) {
		add(m[1])
	}
	for _, m := range pathLike.FindAllStringSubmatch(text, 4) {
		add(m[1])
	}
	if len(out) > 4 {
		out = out[:4]
	}
	return out
}

func jail(root, p string) (string, bool) {
	if root == "" || p == "" {
		return "", false
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", false
	}
	cand := p
	if !filepath.IsAbs(p) {
		cand = filepath.Join(rootAbs, p)
	}
	cand = filepath.Clean(cand)
	rel, err := filepath.Rel(rootAbs, cand)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", false
	}
	return cand, true
}

func encodeImage(root, path string, cols int, proto Protocol) string {
	label := filepath.Base(path)
	placeholder := "[image: " + label + "]"
	full, ok := jail(root, path)
	if !ok {
		return placeholder
	}
	st, err := os.Stat(full)
	if err != nil || st.IsDir() || st.Size() <= 0 || st.Size() > maxImageB {
		return placeholder
	}
	key := full + "\x00" + strconv.FormatInt(st.Size(), 10) + "\x00" + strconv.FormatInt(st.ModTime().UnixNano(), 10) + "\x00" + strconv.Itoa(cols) + "\x00" + string(proto)
	imgMu.Lock()
	if hit, ok := imgMemo[key]; ok {
		imgMu.Unlock()
		return hit
	}
	imgMu.Unlock()

	f, err := os.Open(full)
	if err != nil {
		return placeholder
	}
	data, err := io.ReadAll(io.LimitReader(f, maxImageB+1))
	_ = f.Close()
	if err != nil || len(data) == 0 || len(data) > maxImageB {
		return placeholder
	}
	out := placeholder
	if proto != ProtocolNone {
		if cols < 8 {
			cols = 8
		}
		if cols > 40 {
			cols = 40
		}
		switch proto {
		case ProtocolITerm:
			if imageWithinLimits(data) {
				out = itermInline(data, cols)
			}
		case ProtocolKitty:
			pngb, ok := asPNG(data)
			if ok {
				out = kittyInline(pngb, cols)
			}
		}
	}
	imgMu.Lock()
	if len(imgMemo) >= maxImgMemo {
		for k := range imgMemo {
			delete(imgMemo, k)
			break
		}
	}
	imgMemo[key] = out
	imgMu.Unlock()
	return out
}

func imageWithinLimits(data []byte) bool {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return false
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > maxImageSide || cfg.Height > maxImageSide {
		return false
	}
	if cfg.Width*cfg.Height > maxImagePx {
		return false
	}
	return true
}

func asPNG(data []byte) ([]byte, bool) {
	if !imageWithinLimits(data) {
		return nil, false
	}
	if bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4e, 0x47}) {
		return data, true
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, false
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, false
	}
	return buf.Bytes(), true
}

func itermInline(data []byte, cols int) string {
	b64 := base64.StdEncoding.EncodeToString(data)
	return "\x1b]1337;File=inline=1;width=" + strconv.Itoa(cols) + ":" + b64 + "\a"
}

func kittyInline(data []byte, cols int) string {
	b64 := base64.StdEncoding.EncodeToString(data)
	var b strings.Builder
	first := true
	for len(b64) > 0 {
		chunk := b64
		more := 0
		if len(chunk) > chunkKitty {
			chunk = b64[:chunkKitty]
			b64 = b64[chunkKitty:]
			more = 1
		} else {
			b64 = ""
		}
		if first {
			b.WriteString("\x1b_Ga=T,f=100,c=")
			b.WriteString(strconv.Itoa(cols))
			b.WriteString(",m=")
			b.WriteString(strconv.Itoa(more))
			b.WriteByte(';')
			first = false
		} else {
			b.WriteString("\x1b_Gm=")
			b.WriteString(strconv.Itoa(more))
			b.WriteByte(';')
		}
		b.WriteString(chunk)
		b.WriteString("\x1b\\")
	}
	return b.String()
}
