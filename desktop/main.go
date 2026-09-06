// Kunta — piccolo laboratorio testuale ShiduLab.
// Portable Windows GUI, solo API Win32 + libreria standard Go.
package main

import (
	_ "embed"
	"encoding/binary"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"
)

const appTitle = "Kunta"

const (
	WS_OVERLAPPED  = 0x00000000
	WS_CAPTION     = 0x00C00000
	WS_SYSMENU     = 0x00080000
	WS_THICKFRAME  = 0x00040000
	WS_MINIMIZEBOX = 0x00020000
	WS_MAXIMIZEBOX = 0x00010000
	WS_VISIBLE     = 0x10000000
	WS_CHILD       = 0x40000000
	WS_BORDER      = 0x00800000
	WS_VSCROLL     = 0x00200000
	WS_HSCROLL     = 0x00100000
	WS_TABSTOP     = 0x00010000

	ES_MULTILINE   = 0x0004
	ES_AUTOVSCROLL = 0x0040
	ES_AUTOHSCROLL = 0x0080
	ES_WANTRETURN  = 0x1000
	ES_READONLY    = 0x0800

	BS_PUSHBUTTON    = 0x00000000
	BS_DEFPUSHBUTTON = 0x00000001
	BS_AUTOCHECKBOX  = 0x00000003

	CBS_DROPDOWNLIST = 0x0003
	CBS_HASSTRINGS   = 0x0200
	CBN_SELCHANGE    = 1

	SS_LEFT  = 0x00000000
	SS_RIGHT = 0x00000002
	SS_ICON  = 0x00000003

	WM_CREATE    = 0x0001
	WM_DESTROY   = 0x0002
	WM_SIZE      = 0x0005
	WM_COMMAND   = 0x0111
	WM_SETFONT   = 0x0030
	WM_SETICON   = 0x0080
	WM_DROPFILES = 0x0233
	WM_COPY      = 0x0301
	WM_PASTE     = 0x0302

	EM_SETSEL       = 0x00B1
	EM_SETCUEBANNER = 0x1501

	BM_GETCHECK = 0x00F0
	BST_CHECKED = 1

	CB_ADDSTRING = 0x0143
	CB_GETCURSEL = 0x0147
	CB_SETCURSEL = 0x014E

	STM_SETICON = 0x0170

	COLOR_WINDOW       = 5
	IDC_ARROW          = 32512
	DEFAULT_GUI_FONT   = 17
	MB_OK              = 0x00000000
	MB_ICONINFORMATION = 0x00000040
	MB_ICONERROR       = 0x00000010
	OFN_FILEMUSTEXIST  = 0x00001000
	OFN_PATHMUSTEXIST  = 0x00000800
	LR_DEFAULTCOLOR    = 0x0000

	ICON_SMALL = 0
	ICON_BIG   = 1

	CW_USEDEFAULT = 0x80000000
)

const (
	IDPaste      = 101
	IDOpen       = 102
	IDClear      = 103
	IDInput      = 104
	IDOperation  = 105
	IDParam      = 106
	IDCase       = 107
	IDRun        = 108
	IDResult     = 109
	IDCopyResult = 110
	IDBrandIcon  = 111
	IDBrandText  = 112
	IDParamLabel = 113
)

type POINT struct{ X, Y int32 }
type RECT struct{ Left, Top, Right, Bottom int32 }
type MSG struct {
	Hwnd     uintptr
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       POINT
	LPrivate uint32
}
type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}
type OPENFILENAME struct {
	LStructSize       uint32
	HwndOwner         uintptr
	HInstance         uintptr
	LpstrFilter       *uint16
	LpstrCustomFilter *uint16
	NMaxCustFilter    uint32
	NFilterIndex      uint32
	LpstrFile         *uint16
	NMaxFile          uint32
	LpstrFileTitle    *uint16
	NMaxFileTitle     uint32
	LpstrInitialDir   *uint16
	LpstrTitle        *uint16
	Flags             uint32
	NFileOffset       uint16
	NFileExtension    uint16
	LpstrDefExt       *uint16
	LCustData         uintptr
	LpfnHook          uintptr
	LpTemplateName    *uint16
	PvReserved        uintptr
	DwReserved        uint32
	FlagsEx           uint32
}

//go:embed assets/Botolo.ico
var botoloICO []byte

//go:embed assets/Kunta.ico
var kuntaICO []byte

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	comdlg32 = syscall.NewLazyDLL("comdlg32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")

	pRegisterClassExW         = user32.NewProc("RegisterClassExW")
	pCreateWindowExW          = user32.NewProc("CreateWindowExW")
	pDefWindowProcW           = user32.NewProc("DefWindowProcW")
	pShowWindow               = user32.NewProc("ShowWindow")
	pUpdateWindow             = user32.NewProc("UpdateWindow")
	pGetMessageW              = user32.NewProc("GetMessageW")
	pTranslateMessage         = user32.NewProc("TranslateMessage")
	pDispatchMessageW         = user32.NewProc("DispatchMessageW")
	pPostQuitMessage          = user32.NewProc("PostQuitMessage")
	pSendMessageW             = user32.NewProc("SendMessageW")
	pSetWindowTextW           = user32.NewProc("SetWindowTextW")
	pGetWindowTextLengthW     = user32.NewProc("GetWindowTextLengthW")
	pGetWindowTextW           = user32.NewProc("GetWindowTextW")
	pMoveWindow               = user32.NewProc("MoveWindow")
	pGetClientRect            = user32.NewProc("GetClientRect")
	pLoadCursorW              = user32.NewProc("LoadCursorW")
	pMessageBoxW              = user32.NewProc("MessageBoxW")
	pCreateIconFromResourceEx = user32.NewProc("CreateIconFromResourceEx")
	pDestroyIcon              = user32.NewProc("DestroyIcon")
	pSetFocus                 = user32.NewProc("SetFocus")

	pGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	pGetStockObject   = gdi32.NewProc("GetStockObject")
	pGetOpenFileNameW = comdlg32.NewProc("GetOpenFileNameW")
	pDragAcceptFiles  = shell32.NewProc("DragAcceptFiles")
	pDragQueryFileW   = shell32.NewProc("DragQueryFileW")
	pDragFinish       = shell32.NewProc("DragFinish")
	pSetCurrentProcessExplicitAppUserModelID = shell32.NewProc("SetCurrentProcessExplicitAppUserModelID")
)

var (
	hwndMain       uintptr
	hwndInput      uintptr
	hwndOperation  uintptr
	hwndParam      uintptr
	hwndParamLabel uintptr
	hwndCase       uintptr
	hwndResult     uintptr
	hwndBrandIcon  uintptr
	hwndBrandText  uintptr
	hFont          uintptr
	hIconBig       uintptr
	hIconSmall     uintptr
	hBrandIcon     uintptr
)

var operations = []string{
	"Lettere indicate",
	"Caratteri",
	"Parole",
	"Righe",
	"Spazi e bianchi",
	"Cifre",
	"Punteggiatura",
	"Occorrenze parola/stringa",
	"Parole uniche",
	"Parole ripetute",
	"Frequenza lettere",
	"Frequenza parole",
	"Parola più corta / più lunga",
	"Palindromi (parole)",
	"Bifronti nel testo",
	"Coppie/gruppi di anagrammi",
	"Acrostico (iniziali righe)",
	"Telestico (finali righe)",
	"Inverti il testo",
	"Inverti ordine parole",
	"Ordina parole alfabeticamente",
	"Estrai solo numeri",
	"Schema rime / desinenze",
	"Rima baciata / inclusione",
	"LetterTransport",
	"Sciarade / salti di dominio",
	"Kunta il testo (fenomeni)",
}

func wptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func utf16Multi(s string) []uint16 {
	r := utf16.Encode([]rune(s))
	return append(r, 0)
}

func createIconFromICO(data []byte, desired int) uintptr {
	if len(data) < 6 || binary.LittleEndian.Uint16(data[0:2]) != 0 || binary.LittleEndian.Uint16(data[2:4]) != 1 {
		return 0
	}
	count := int(binary.LittleEndian.Uint16(data[4:6]))
	if count <= 0 || len(data) < 6+count*16 {
		return 0
	}
	best := -1
	bestDiff := int(^uint(0) >> 1)
	for i := 0; i < count; i++ {
		p := 6 + i*16
		ww, hh := int(data[p]), int(data[p+1])
		if ww == 0 {
			ww = 256
		}
		if hh == 0 {
			hh = 256
		}
		d := abs(ww-desired) + abs(hh-desired)
		if d < bestDiff {
			bestDiff, best = d, p
		}
	}
	if best < 0 {
		return 0
	}
	sz := int(binary.LittleEndian.Uint32(data[best+8 : best+12]))
	off := int(binary.LittleEndian.Uint32(data[best+12 : best+16]))
	if sz <= 0 || off < 0 || off+sz > len(data) {
		return 0
	}
	r, _, _ := pCreateIconFromResourceEx.Call(
		uintptr(unsafe.Pointer(&data[off])), uintptr(sz), 1, 0x00030000,
		uintptr(desired), uintptr(desired), LR_DEFAULTCOLOR)
	return r
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func createControl(class, text string, style uintptr, id int) uintptr {
	h, _, _ := pCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(wptr(class))),
		uintptr(unsafe.Pointer(wptr(text))),
		WS_CHILD|WS_VISIBLE|style,
		0, 0, 10, 10,
		hwndMain, uintptr(id), 0, 0)
	if h != 0 && hFont != 0 {
		pSendMessageW.Call(h, WM_SETFONT, hFont, 1)
	}
	return h
}

func setText(hwnd uintptr, s string) {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.ReplaceAll(s, "\n", "\r\n")
	pSetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(wptr(s))))
}

func getText(hwnd uintptr) string {
	n, _, _ := pGetWindowTextLengthW.Call(hwnd)
	if n == 0 {
		return ""
	}
	buf := make([]uint16, int(n)+1)
	pGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), n+1)
	s := syscall.UTF16ToString(buf)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}

func message(text string, flags uintptr) {
	pMessageBoxW.Call(hwndMain, uintptr(unsafe.Pointer(wptr(text))), uintptr(unsafe.Pointer(wptr(appTitle))), flags)
}

func isChecked(hwnd uintptr) bool {
	r, _, _ := pSendMessageW.Call(hwnd, BM_GETCHECK, 0, 0)
	return r == BST_CHECKED
}

func loword(v uintptr) uint16 { return uint16(v & 0xffff) }
func hiword(v uintptr) uint16 { return uint16((v >> 16) & 0xffff) }

func layout() {
	if hwndMain == 0 {
		return
	}
	var rc RECT
	pGetClientRect.Call(hwndMain, uintptr(unsafe.Pointer(&rc)))
	w, h := int(rc.Right-rc.Left), int(rc.Bottom-rc.Top)
	if w < 720 {
		w = 720
	}
	if h < 580 {
		h = 580
	}
	margin := 12
	top := 10

	move(IDPaste, margin, top, 90, 28)
	move(IDOpen, margin+98, top, 110, 28)
	move(IDClear, margin+216, top, 80, 28)

	inputTop := top + 38
	inputH := (h - 190) / 2
	if inputH < 150 {
		inputH = 150
	}
	move(IDInput, margin, inputTop, w-2*margin, inputH)

	y := inputTop + inputH + 10
	move(IDOperation, margin, y, 250, 220)
	move(IDParamLabel, margin+262, y+4, 155, 22)
	move(IDParam, margin+420, y, 190, 26)
	move(IDCase, margin+620, y+2, 190, 24)
	move(IDRun, w-margin-105, y-1, 105, 29)

	resultTop := y + 38
	brandH := 34
	resultH := h - resultTop - brandH - 12
	if resultH < 120 {
		resultH = 120
	}
	move(IDResult, margin, resultTop, w-2*margin, resultH)
	move(IDCopyResult, margin, resultTop+resultH+6, 120, 26)

	brandY := resultTop + resultH + 3
	move(IDBrandIcon, w-margin-125, brandY, 32, 32)
	move(IDBrandText, w-margin-90, brandY+7, 90, 22)
}

func move(id, x, y, w, h int) {
	var hwnd uintptr
	switch id {
	case IDPaste:
		hwnd = getDlgItem(IDPaste)
	case IDOpen:
		hwnd = getDlgItem(IDOpen)
	case IDClear:
		hwnd = getDlgItem(IDClear)
	case IDInput:
		hwnd = hwndInput
	case IDOperation:
		hwnd = hwndOperation
	case IDParam:
		hwnd = hwndParam
	case IDParamLabel:
		hwnd = hwndParamLabel
	case IDCase:
		hwnd = hwndCase
	case IDRun:
		hwnd = getDlgItem(IDRun)
	case IDResult:
		hwnd = hwndResult
	case IDCopyResult:
		hwnd = getDlgItem(IDCopyResult)
	case IDBrandIcon:
		hwnd = hwndBrandIcon
	case IDBrandText:
		hwnd = hwndBrandText
	}
	if hwnd != 0 {
		pMoveWindow.Call(hwnd, uintptr(x), uintptr(y), uintptr(w), uintptr(h), 1)
	}
}

func getDlgItem(id int) uintptr {
	proc := user32.NewProc("GetDlgItem")
	h, _, _ := proc.Call(hwndMain, uintptr(id))
	return h
}

func updateParamHint() {
	idx, _, _ := pSendMessageW.Call(hwndOperation, CB_GETCURSEL, 0, 0)
	label := "Parametro:"
	cue := "facoltativo"
	switch int(idx) {
	case 0:
		label, cue = "Lettere da contare:", "es. aeiou"
	case 7:
		label, cue = "Parola/stringa:", "es. uomo"
	case 13, 14, 15:
		label, cue = "Lunghezza minima:", "default 3"
	case 22:
		label, cue = "Lettere desinenza:", "default 3"
	case 23:
		label, cue = "Nucleo minimo:", "default 3"
	case 24:
		label, cue = "Trasporto minimo:", "default 3"
	case 25:
		label, cue = "Segmento minimo:", "default 2"
	case 26:
		label, cue = "Sensibilità:", "facoltativo"
	}
	setText(hwndParamLabel, label)
	pSendMessageW.Call(hwndParam, EM_SETCUEBANNER, 1, uintptr(unsafe.Pointer(wptr(cue))))
}

func openTextFile() {
	buf := make([]uint16, 32768)
	filter := utf16Multi("Testi (*.txt;*.md;*.log;*.csv;*.tsv)\x00*.txt;*.md;*.log;*.csv;*.tsv\x00Tutti i file (*.*)\x00*.*\x00\x00")
	ofn := OPENFILENAME{
		LStructSize: uint32(unsafe.Sizeof(OPENFILENAME{})),
		HwndOwner:   hwndMain,
		LpstrFilter: &filter[0],
		LpstrFile:   &buf[0],
		NMaxFile:    uint32(len(buf)),
		LpstrTitle:  wptr("Apri un testo"),
		Flags:       OFN_FILEMUSTEXIST | OFN_PATHMUSTEXIST,
	}
	ok, _, _ := pGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if ok == 0 {
		return
	}
	path := syscall.UTF16ToString(buf)
	loadFile(path)
}

func loadFile(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		message("Non riesco ad aprire il file:\n"+err.Error(), MB_OK|MB_ICONERROR)
		return
	}
	text := decodeText(b)
	setText(hwndInput, text)
	setText(hwndResult, fmt.Sprintf("Caricato: %s\n%d byte · %d caratteri", path, len(b), utf8.RuneCountInString(text)))
}

func decodeText(b []byte) string {
	if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		return string(b[3:])
	}
	if len(b) >= 2 && b[0] == 0xFF && b[1] == 0xFE {
		u := make([]uint16, 0, (len(b)-2)/2)
		for i := 2; i+1 < len(b); i += 2 {
			u = append(u, binary.LittleEndian.Uint16(b[i:i+2]))
		}
		return string(utf16.Decode(u))
	}
	if len(b) >= 2 && b[0] == 0xFE && b[1] == 0xFF {
		u := make([]uint16, 0, (len(b)-2)/2)
		for i := 2; i+1 < len(b); i += 2 {
			u = append(u, binary.BigEndian.Uint16(b[i:i+2]))
		}
		return string(utf16.Decode(u))
	}
	if utf8.Valid(b) {
		return string(b)
	}
	// Fallback prudente: conserva ogni byte come rune 0..255, utile per vecchi TXT ANSI.
	r := make([]rune, len(b))
	for i, c := range b {
		r[i] = rune(c)
	}
	return string(r)
}

func droppedFile(hdrop uintptr) {
	n, _, _ := pDragQueryFileW.Call(hdrop, 0xFFFFFFFF, 0, 0)
	if n == 0 {
		pDragFinish.Call(hdrop)
		return
	}
	ln, _, _ := pDragQueryFileW.Call(hdrop, 0, 0, 0)
	buf := make([]uint16, int(ln)+1)
	pDragQueryFileW.Call(hdrop, 0, uintptr(unsafe.Pointer(&buf[0])), ln+1)
	pDragFinish.Call(hdrop)
	loadFile(syscall.UTF16ToString(buf))
}

func normalizeCase(s string, sensitive bool) string {
	if sensitive {
		return s
	}
	return strings.ToLower(s)
}

func normalizeLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func tokenizeWords(s string) []string {
	var out []string
	var b []rune
	flush := func() {
		if len(b) == 0 {
			return
		}
		w := strings.Trim(string(b), "-'’")
		if w != "" {
			out = append(out, w)
		}
		b = b[:0]
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '\'' || r == '’' || r == '-' {
			b = append(b, r)
		} else {
			flush()
		}
	}
	flush()
	return out
}

func cleanWord(s string, sensitive bool) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return normalizeCase(b.String(), sensitive)
}

func reverseRunes(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func parseMinLen(param string) int {
	n, err := strconv.Atoi(strings.TrimSpace(param))
	if err != nil || n < 1 {
		return 3
	}
	return n
}

func analyze(text string, op int, param string, sensitive bool) string {
	words := tokenizeWords(text)
	lines := normalizeLines(text)
	switch op {
	case 0:
		if strings.TrimSpace(param) == "" {
			return "Scrivi nel campo Parametro le lettere da contare."
		}
		seen := map[rune]bool{}
		var targets []rune
		for _, r := range param {
			if unicode.IsSpace(r) {
				continue
			}
			rr := r
			if !sensitive {
				rr = []rune(strings.ToLower(string(r)))[0]
			}
			if !seen[rr] {
				seen[rr] = true
				targets = append(targets, rr)
			}
		}
		counts := map[rune]int{}
		source := text
		if !sensitive {
			source = strings.ToLower(source)
		}
		for _, r := range source {
			if seen[r] {
				counts[r]++
			}
		}
		var b strings.Builder
		total := 0
		for _, r := range targets {
			fmt.Fprintf(&b, "%c = %d\n", r, counts[r])
			total += counts[r]
		}
		fmt.Fprintf(&b, "\nTotale caratteri cercati: %d", total)
		return b.String()
	case 1:
		return fmt.Sprintf("Caratteri (Unicode): %d\nByte UTF-8: %d", utf8.RuneCountInString(text), len([]byte(text)))
	case 2:
		return fmt.Sprintf("Parole: %d\nParole distinte: %d", len(words), distinctWordCount(words, sensitive))
	case 3:
		return fmt.Sprintf("Righe: %d\nRighe non vuote: %d", len(lines), nonEmptyLines(lines))
	case 4:
		spaces, white := 0, 0
		for _, r := range text {
			if r == ' ' {
				spaces++
			}
			if unicode.IsSpace(r) {
				white++
			}
		}
		return fmt.Sprintf("Spazi semplici: %d\nCaratteri bianchi totali: %d", spaces, white)
	case 5:
		n := 0
		for _, r := range text {
			if unicode.IsDigit(r) {
				n++
			}
		}
		return fmt.Sprintf("Cifre: %d", n)
	case 6:
		n := 0
		for _, r := range text {
			if unicode.IsPunct(r) {
				n++
			}
		}
		return fmt.Sprintf("Segni di punteggiatura: %d", n)
	case 7:
		q := strings.TrimSpace(param)
		if q == "" {
			return "Scrivi una parola o stringa da cercare."
		}
		src, needle := normalizeCase(text, sensitive), normalizeCase(q, sensitive)
		return fmt.Sprintf("%q\nOccorrenze: %d", q, strings.Count(src, needle))
	case 8:
		freq, display := wordFrequency(words, sensitive)
		keys := make([]string, 0, len(freq))
		for k := range freq {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		fmt.Fprintf(&b, "Parole uniche: %d\n\n", len(keys))
		for _, k := range keys {
			fmt.Fprintf(&b, "%s\n", display[k])
		}
		return limitLines(b.String(), 700)
	case 9:
		freq, display := wordFrequency(words, sensitive)
		type pair struct {
			k string
			n int
		}
		var a []pair
		for k, n := range freq {
			if n > 1 {
				a = append(a, pair{k, n})
			}
		}
		sort.Slice(a, func(i, j int) bool {
			if a[i].n == a[j].n {
				return a[i].k < a[j].k
			}
			return a[i].n > a[j].n
		})
		var b strings.Builder
		fmt.Fprintf(&b, "Parole ripetute: %d\n\n", len(a))
		for _, p := range a {
			fmt.Fprintf(&b, "%s = %d\n", display[p.k], p.n)
		}
		return limitLines(b.String(), 700)
	case 10:
		freq := map[rune]int{}
		for _, r := range normalizeCase(text, sensitive) {
			if unicode.IsLetter(r) {
				freq[r]++
			}
		}
		type rp struct {
			r rune
			n int
		}
		a := make([]rp, 0, len(freq))
		for r, n := range freq {
			a = append(a, rp{r, n})
		}
		sort.Slice(a, func(i, j int) bool {
			if a[i].n == a[j].n {
				return a[i].r < a[j].r
			}
			return a[i].n > a[j].n
		})
		var b strings.Builder
		for _, p := range a {
			fmt.Fprintf(&b, "%c = %d\n", p.r, p.n)
		}
		return b.String()
	case 11:
		freq, display := wordFrequency(words, sensitive)
		type pair struct {
			k string
			n int
		}
		a := make([]pair, 0, len(freq))
		for k, n := range freq {
			a = append(a, pair{k, n})
		}
		sort.Slice(a, func(i, j int) bool {
			if a[i].n == a[j].n {
				return a[i].k < a[j].k
			}
			return a[i].n > a[j].n
		})
		var b strings.Builder
		for _, p := range a {
			fmt.Fprintf(&b, "%s = %d\n", display[p.k], p.n)
		}
		return limitLines(b.String(), 700)
	case 12:
		if len(words) == 0 {
			return "Nessuna parola."
		}
		min, max := 1<<30, 0
		var mins, maxs []string
		for _, w := range words {
			n := utf8.RuneCountInString(cleanWord(w, true))
			if n == 0 {
				continue
			}
			if n < min {
				min = n
				mins = []string{w}
			} else if n == min {
				mins = appendUnique(mins, w)
			}
			if n > max {
				max = n
				maxs = []string{w}
			} else if n == max {
				maxs = appendUnique(maxs, w)
			}
		}
		return fmt.Sprintf("Più corta (%d): %s\n\nPiù lunga (%d): %s", min, strings.Join(mins, ", "), max, strings.Join(maxs, ", "))
	case 13:
		return findPalindromes(words, sensitive, parseMinLen(param))
	case 14:
		return findBifronti(words, sensitive, parseMinLen(param))
	case 15:
		return findAnagrams(words, sensitive, parseMinLen(param))
	case 16:
		var r []rune
		used := 0
		for _, line := range lines {
			t := strings.TrimSpace(line)
			if t == "" {
				continue
			}
			for _, c := range t {
				if unicode.IsLetter(c) || unicode.IsDigit(c) {
					r = append(r, c)
					used++
					break
				}
			}
		}
		return fmt.Sprintf("Acrostico (%d righe):\n\n%s", used, string(r))
	case 17:
		var r []rune
		used := 0
		for _, line := range lines {
			t := strings.TrimSpace(line)
			if t == "" {
				continue
			}
			rr := []rune(t)
			for i := len(rr) - 1; i >= 0; i-- {
				if unicode.IsLetter(rr[i]) || unicode.IsDigit(rr[i]) {
					r = append(r, rr[i])
					used++
					break
				}
			}
		}
		return fmt.Sprintf("Telestico (%d righe):\n\n%s", used, string(r))
	case 18:
		return reverseRunes(text)
	case 19:
		a := append([]string(nil), words...)
		for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
			a[i], a[j] = a[j], a[i]
		}
		return strings.Join(a, " ")
	case 20:
		a := append([]string(nil), words...)
		sort.SliceStable(a, func(i, j int) bool { return normalizeCase(a[i], sensitive) < normalizeCase(a[j], sensitive) })
		return strings.Join(a, "\n")
	case 21:
		var nums []string
		var b strings.Builder
		flush := func() {
			if b.Len() > 0 {
				nums = append(nums, b.String())
				b.Reset()
			}
		}
		for _, r := range text {
			if unicode.IsDigit(r) {
				b.WriteRune(r)
			} else {
				flush()
			}
		}
		flush()
		return fmt.Sprintf("Sequenze numeriche: %d\n\n%s", len(nums), strings.Join(nums, "\n"))
	case 22:
		return analyzeRhymeScheme(text, sensitive, parsePositive(param, 3, 1, 12))
	case 23:
		return findInclusionChains(words, sensitive, parsePositive(param, 3, 2, 20))
	case 24:
		return findLetterTransport(words, sensitive, parsePositive(param, 3, 2, 20))
	case 25:
		return findSciarades(text, sensitive, parsePositive(param, 2, 1, 8))
	case 26:
		return kuntaPhenomena(text, sensitive)
	}
	return "Operazione non riconosciuta."
}

func distinctWordCount(words []string, sensitive bool) int {
	m := map[string]struct{}{}
	for _, w := range words {
		k := cleanWord(w, sensitive)
		if k != "" {
			m[k] = struct{}{}
		}
	}
	return len(m)
}
func nonEmptyLines(lines []string) int {
	n := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			n++
		}
	}
	return n
}
func wordFrequency(words []string, sensitive bool) (map[string]int, map[string]string) {
	f := map[string]int{}
	d := map[string]string{}
	for _, w := range words {
		k := cleanWord(w, sensitive)
		if k == "" {
			continue
		}
		f[k]++
		if _, ok := d[k]; !ok {
			d[k] = w
		}
	}
	return f, d
}
func appendUnique(a []string, s string) []string {
	for _, v := range a {
		if v == s {
			return a
		}
	}
	return append(a, s)
}

func findPalindromes(words []string, sensitive bool, minLen int) string {
	f, d := wordFrequency(words, sensitive)
	type p struct {
		k string
		n int
	}
	var a []p
	total := 0
	for k, n := range f {
		if utf8.RuneCountInString(k) >= minLen && k == reverseRunes(k) {
			a = append(a, p{k, n})
			total += n
		}
	}
	sort.Slice(a, func(i, j int) bool {
		if a[i].n == a[j].n {
			return a[i].k < a[j].k
		}
		return a[i].n > a[j].n
	})
	var b strings.Builder
	fmt.Fprintf(&b, "Palindromi distinti: %d\nOccorrenze totali: %d\nLunghezza minima: %d\n\n", len(a), total, minLen)
	for _, p := range a {
		fmt.Fprintf(&b, "%s = %d\n", d[p.k], p.n)
	}
	return b.String()
}

func findBifronti(words []string, sensitive bool, minLen int) string {
	f, d := wordFrequency(words, sensitive)
	seen := map[string]bool{}
	var pairs [][2]string
	for k := range f {
		if utf8.RuneCountInString(k) < minLen {
			continue
		}
		rev := reverseRunes(k)
		if rev == k {
			continue
		}
		if _, ok := f[rev]; !ok {
			continue
		}
		key := k + "\x00" + rev
		key2 := rev + "\x00" + k
		if seen[key] || seen[key2] {
			continue
		}
		seen[key] = true
		pairs = append(pairs, [2]string{d[k], d[rev]})
	}
	sort.Slice(pairs, func(i, j int) bool { return strings.ToLower(pairs[i][0]) < strings.ToLower(pairs[j][0]) })
	var b strings.Builder
	fmt.Fprintf(&b, "Bifronti trovati: %d\nLunghezza minima: %d\n\n", len(pairs), minLen)
	for _, p := range pairs {
		fmt.Fprintf(&b, "%s ↔ %s\n", p[0], p[1])
	}
	return b.String()
}

func sortedRunes(s string) string {
	r := []rune(s)
	sort.Slice(r, func(i, j int) bool { return r[i] < r[j] })
	return string(r)
}
func findAnagrams(words []string, sensitive bool, minLen int) string {
	_, d := wordFrequency(words, sensitive)
	groups := map[string][]string{}
	for k, disp := range d {
		if utf8.RuneCountInString(k) < minLen {
			continue
		}
		sig := sortedRunes(k)
		groups[sig] = append(groups[sig], disp)
	}
	type g struct {
		sig   string
		words []string
	}
	var a []g
	for sig, ws := range groups {
		if len(ws) > 1 {
			sort.Slice(ws, func(i, j int) bool { return strings.ToLower(ws[i]) < strings.ToLower(ws[j]) })
			a = append(a, g{sig, ws})
		}
	}
	sort.Slice(a, func(i, j int) bool { return strings.ToLower(a[i].words[0]) < strings.ToLower(a[j].words[0]) })
	var b strings.Builder
	fmt.Fprintf(&b, "Gruppi di anagrammi: %d\nLunghezza minima: %d\n\n", len(a), minLen)
	for _, g := range a {
		fmt.Fprintf(&b, "%s\n", strings.Join(g.words, " · "))
	}
	return limitLines(b.String(), 700)
}

func parsePositive(param string, def, min, max int) int {
	n, err := strconv.Atoi(strings.TrimSpace(param))
	if err != nil {
		return def
	}
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

func runeSuffix(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return string(r)
	}
	return string(r[len(r)-n:])
}

func finalWordOfLine(line string, sensitive bool) string {
	ws := tokenizeWords(line)
	if len(ws) == 0 {
		return ""
	}
	return cleanWord(ws[len(ws)-1], sensitive)
}

func rhymeLabel(n int) string {
	labels := []string{"x", "y", "z", "w", "v", "u", "t", "s", "r", "q", "p", "o", "n", "m", "l", "k", "j", "i", "h", "g", "f", "e", "d", "c", "b", "a"}
	if n < len(labels) {
		return labels[n]
	}
	return fmt.Sprintf("r%d", n+1)
}

func analyzeRhymeScheme(text string, sensitive bool, depth int) string {
	lines := normalizeLines(text)
	var stanzas [][]string
	var cur []string
	flush := func() {
		if len(cur) > 0 {
			stanzas = append(stanzas, cur)
			cur = nil
		}
	}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			flush()
			continue
		}
		cur = append(cur, line)
	}
	flush()
	if len(stanzas) == 0 {
		return "Nessun verso rilevato."
	}

	var out strings.Builder
	fmt.Fprintf(&out, "Rima = desinenza grafica · profondità: %d lettere\n", depth)
	fmt.Fprintf(&out, "Strofe rilevate: %d\n\n", len(stanzas))
	for si, stanza := range stanzas {
		type row struct{ word, suffix, label string }
		var rows []row
		counts := map[string]int{}
		var order []string
		seenSuffix := map[string]bool{}
		for _, line := range stanza {
			w := finalWordOfLine(line, sensitive)
			if w == "" {
				continue
			}
			suffix := runeSuffix(w, depth)
			rows = append(rows, row{word: w, suffix: suffix})
			counts[suffix]++
			if !seenSuffix[suffix] {
				seenSuffix[suffix] = true
				order = append(order, suffix)
			}
		}

		groups := map[string]string{}
		// Convenzione Kunta: famiglia dominante + un verso isolato = x-x-x-z.
		if len(order) == 2 {
			var singleton, dominant string
			for _, suf := range order {
				if counts[suf] == 1 {
					singleton = suf
				} else {
					dominant = suf
				}
			}
			if singleton != "" && dominant != "" {
				groups[dominant] = "x"
				groups[singleton] = "z"
			}
		}
		if len(groups) == 0 {
			for i, suf := range order {
				groups[suf] = rhymeLabel(i)
			}
		}
		var scheme []string
		for i := range rows {
			rows[i].label = groups[rows[i].suffix]
			scheme = append(scheme, rows[i].label)
		}
		if len(rows) == 0 {
			continue
		}
		if len(stanzas) > 1 {
			fmt.Fprintf(&out, "STROFA %d\n", si+1)
		}
		for i, r := range rows {
			fmt.Fprintf(&out, "%2d. %-22s  → -%-10s → %s\n", i+1, r.word, r.suffix, r.label)
		}
		fmt.Fprintf(&out, "\nSchema: %s\n", strings.Join(scheme, "-"))

		// Mostra anche la risalita progressiva della coda: 1, 2 ... depth lettere.
		fmt.Fprintf(&out, "Coda progressiva:\n")
		for d := 1; d <= depth; d++ {
			freq := map[string]int{}
			for _, r := range rows {
				freq[runeSuffix(r.word, d)]++
			}
			var keys []string
			for k, n := range freq {
				if n > 1 {
					keys = append(keys, k)
				}
			}
			sort.Strings(keys)
			if len(keys) == 0 {
				continue
			}
			var parts []string
			for _, k := range keys {
				parts = append(parts, fmt.Sprintf("-%s ×%d", k, freq[k]))
			}
			fmt.Fprintf(&out, "  %d: %s\n", d, strings.Join(parts, " · "))
		}
		if si < len(stanzas)-1 {
			out.WriteString("\n")
		}
	}
	return limitLines(out.String(), 700)
}

func findInclusionChains(words []string, sensitive bool, minLen int) string {
	_, display := wordFrequency(words, sensitive)
	var keys []string
	for k := range display {
		if utf8.RuneCountInString(k) >= minLen {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		li, lj := utf8.RuneCountInString(keys[i]), utf8.RuneCountInString(keys[j])
		if li == lj {
			return keys[i] < keys[j]
		}
		return li < lj
	})

	children := map[string][]string{}
	incoming := map[string]int{}
	for _, a := range keys {
		la := utf8.RuneCountInString(a)
		var candidates []string
		for _, b := range keys {
			lb := utf8.RuneCountInString(b)
			if lb > la && strings.HasSuffix(b, a) {
				candidates = append(candidates, b)
			}
		}
		for _, b := range candidates {
			lb := utf8.RuneCountInString(b)
			immediate := true
			for _, c := range candidates {
				lc := utf8.RuneCountInString(c)
				if lc <= la || lc >= lb || c == b {
					continue
				}
				if strings.HasSuffix(c, a) && strings.HasSuffix(b, c) {
					immediate = false
					break
				}
			}
			if immediate {
				children[a] = append(children[a], b)
				incoming[b]++
			}
		}
		sort.Slice(children[a], func(i, j int) bool {
			li, lj := utf8.RuneCountInString(children[a][i]), utf8.RuneCountInString(children[a][j])
			if li == lj {
				return children[a][i] < children[a][j]
			}
			return li < lj
		})
	}

	var chains [][]string
	var dfs func(string, []string)
	dfs = func(node string, path []string) {
		path = append(path, node)
		if len(children[node]) == 0 {
			if len(path) >= 2 {
				cp := append([]string(nil), path...)
				chains = append(chains, cp)
			}
			return
		}
		for _, ch := range children[node] {
			dfs(ch, path)
		}
	}
	for _, k := range keys {
		if incoming[k] == 0 && len(children[k]) > 0 {
			dfs(k, nil)
		}
	}
	sort.Slice(chains, func(i, j int) bool {
		if len(chains[i]) == len(chains[j]) {
			return chains[i][0] < chains[j][0]
		}
		return len(chains[i]) > len(chains[j])
	})

	var out strings.Builder
	fmt.Fprintf(&out, "Rima baciata · inclusione progressiva\nNucleo minimo: %d lettere\nCatene trovate: %d\n\n", minLen, len(chains))
	for i, chain := range chains {
		var shown []string
		for _, k := range chain {
			shown = append(shown, display[k])
		}
		fmt.Fprintf(&out, "%2d. %s\n", i+1, strings.Join(shown, " → "))
		fmt.Fprintf(&out, "    nucleo: %s · profondità: %d\n", display[chain[0]], len(chain))
		if i >= 99 {
			out.WriteString("\n… altre catene omesse.\n")
			break
		}
	}
	if len(chains) == 0 {
		out.WriteString("Nessuna inclusione progressiva rilevata.")
	}
	return out.String()
}

func longestCommonSubstring(a, b string) string {
	ar, br := []rune(a), []rune(b)
	if len(ar) == 0 || len(br) == 0 {
		return ""
	}
	prev := make([]int, len(br)+1)
	bestLen, bestEnd := 0, 0
	for i := 1; i <= len(ar); i++ {
		cur := make([]int, len(br)+1)
		for j := 1; j <= len(br); j++ {
			if ar[i-1] == br[j-1] {
				cur[j] = prev[j-1] + 1
				if cur[j] > bestLen {
					bestLen, bestEnd = cur[j], i
				}
			}
		}
		prev = cur
	}
	if bestLen == 0 {
		return ""
	}
	return string(ar[bestEnd-bestLen : bestEnd])
}

func findLetterTransport(words []string, sensitive bool, minLen int) string {
	if len(words) < 2 {
		return "Servono almeno due parole."
	}
	var out strings.Builder
	found := 0
	fmt.Fprintf(&out, "LetterTransport · parole consecutive\nTrasporto minimo: %d lettere\n\n", minLen)
	for i := 0; i+1 < len(words); i++ {
		a0, b0 := words[i], words[i+1]
		a, b := cleanWord(a0, sensitive), cleanWord(b0, sensitive)
		if a == "" || b == "" || a == b {
			continue
		}
		common := longestCommonSubstring(a, b)
		if utf8.RuneCountInString(common) < minLen {
			continue
		}
		found++
		fmt.Fprintf(&out, "%3d. %s → %s\n     trasporta: %q (%d lettere)\n", found, a0, b0, common, utf8.RuneCountInString(common))
		if found >= 200 {
			out.WriteString("\n… output limitato a 200 passaggi.\n")
			break
		}
	}
	if found == 0 {
		out.WriteString("Nessun trasporto consecutivo rilevato.")
	}
	return out.String()
}

func phraseExists(tokens []string, parts []string) bool {
	if len(parts) == 0 || len(parts) > len(tokens) {
		return false
	}
	for i := 0; i+len(parts) <= len(tokens); i++ {
		ok := true
		for j := range parts {
			if tokens[i+j] != parts[j] {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

func findSciarades(text string, sensitive bool, minSeg int) string {
	raw := tokenizeWords(text)
	tokens := make([]string, 0, len(raw))
	counts := map[string]int{}
	display := map[string]string{}
	for _, w := range raw {
		k := cleanWord(w, sensitive)
		if k == "" {
			continue
		}
		tokens = append(tokens, k)
		counts[k]++
		if _, ok := display[k]; !ok {
			display[k] = w
		}
	}
	var targets []string
	for k := range counts {
		if utf8.RuneCountInString(k) >= minSeg*2 {
			targets = append(targets, k)
		}
	}
	sort.Slice(targets, func(i, j int) bool {
		li, lj := utf8.RuneCountInString(targets[i]), utf8.RuneCountInString(targets[j])
		if li == lj {
			return targets[i] < targets[j]
		}
		return li > lj
	})

	var out strings.Builder
	found := 0
	fmt.Fprintf(&out, "Sciarade interne · segmenti presenti nel testo\nSegmento minimo: %d lettere\n\n", minSeg)
	for _, target := range targets {
		rr := []rune(target)
		segSet := map[string]bool{}
		var segs []string
		// Due segmenti.
		for i := minSeg; i <= len(rr)-minSeg; i++ {
			a, b := string(rr[:i]), string(rr[i:])
			if counts[a] == 0 || counts[b] == 0 {
				continue
			}
			parts := []string{a, b}
			if !phraseExists(tokens, parts) {
				continue
			}
			key := strings.Join(parts, "|")
			if !segSet[key] {
				segSet[key] = true
				segs = append(segs, key)
			}
		}
		// Tre segmenti.
		for i := minSeg; i <= len(rr)-2*minSeg; i++ {
			for j := i + minSeg; j <= len(rr)-minSeg; j++ {
				a, b, c := string(rr[:i]), string(rr[i:j]), string(rr[j:])
				if counts[a] == 0 || counts[b] == 0 || counts[c] == 0 {
					continue
				}
				parts := []string{a, b, c}
				if !phraseExists(tokens, parts) {
					continue
				}
				key := strings.Join(parts, "|")
				if !segSet[key] {
					segSet[key] = true
					segs = append(segs, key)
				}
			}
		}
		if len(segs) == 0 {
			continue
		}
		found++
		fmt.Fprintf(&out, "%s", display[target])
		if counts[target] > 1 {
			fmt.Fprintf(&out, "  [forma intera ×%d]", counts[target])
		}
		out.WriteString("\n")
		sort.Strings(segs)
		for _, seg := range segs {
			parts := strings.Split(seg, "|")
			var shown []string
			for _, p := range parts {
				if d, ok := display[p]; ok {
					shown = append(shown, d)
				} else {
					shown = append(shown, p)
				}
			}
			fmt.Fprintf(&out, "  → %s\n", strings.Join(shown, " + "))
		}
		if counts[target] > 1 {
			out.WriteString("  ↳ possibile salto di dominio: stessa forma intera + segmentazioni diverse; verifica il senso nel contesto.\n")
		}
		out.WriteString("\n")
		if found >= 100 {
			out.WriteString("… altri candidati omessi.\n")
			break
		}
	}
	if found == 0 {
		out.WriteString("Nessuna sciarada interna rilevata con il vocabolario del testo.")
	}
	return out.String()
}

func repeatedWordsSummary(text string, sensitive bool, maxItems int) string {
	words := tokenizeWords(text)
	freq, display := wordFrequency(words, sensitive)
	type pair struct {
		k string
		n int
	}
	var a []pair
	for k, n := range freq {
		if n > 1 {
			a = append(a, pair{k, n})
		}
	}
	sort.Slice(a, func(i, j int) bool {
		if a[i].n == a[j].n {
			return a[i].k < a[j].k
		}
		return a[i].n > a[j].n
	})
	if len(a) == 0 {
		return "Nessuna parola ripetuta."
	}
	if len(a) > maxItems {
		a = a[:maxItems]
	}
	var b strings.Builder
	for _, p := range a {
		fmt.Fprintf(&b, "%s = %d\n", display[p.k], p.n)
	}
	return b.String()
}

func kuntaPhenomena(text string, sensitive bool) string {
	var b strings.Builder
	b.WriteString("KUNTA IL TESTO\n")
	b.WriteString("==============================\n\n")
	b.WriteString("RIPETIZIONI\n")
	b.WriteString(repeatedWordsSummary(text, sensitive, 20))
	b.WriteString("\n\nSCHEMA RIME / DESINENZE\n")
	b.WriteString(analyzeRhymeScheme(text, sensitive, 3))
	b.WriteString("\n\nRIMA BACIATA / INCLUSIONE PROGRESSIVA\n")
	b.WriteString(findInclusionChains(tokenizeWords(text), sensitive, 3))
	b.WriteString("\n\nLETTERTRANSPORT\n")
	b.WriteString(findLetterTransport(tokenizeWords(text), sensitive, 3))
	b.WriteString("\n\nSCIARADE / SALTI DI DOMINIO\n")
	b.WriteString(findSciarades(text, sensitive, 2))
	return limitLines(b.String(), 900)
}

func limitLines(s string, max int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= max {
		return s
	}
	return strings.Join(lines[:max], "\n") + fmt.Sprintf("\n\n… output limitato a %d righe.", max)
}

func runAnalysis() {
	text := getText(hwndInput)
	if text == "" {
		message("Non c'è testo da analizzare.", MB_OK|MB_ICONINFORMATION)
		return
	}
	idx, _, _ := pSendMessageW.Call(hwndOperation, CB_GETCURSEL, 0, 0)
	if int(idx) < 0 || int(idx) >= len(operations) {
		idx = 0
	}
	result := analyze(text, int(idx), getText(hwndParam), isChecked(hwndCase))
	setText(hwndResult, operations[int(idx)]+"\n"+strings.Repeat("—", 42)+"\n"+result)
}

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_CREATE:
		hwndMain = hwnd
		hFont, _, _ = pGetStockObject.Call(DEFAULT_GUI_FONT)
		createControl("BUTTON", "Incolla", BS_PUSHBUTTON|WS_TABSTOP, IDPaste)
		createControl("BUTTON", "Apri testo...", BS_PUSHBUTTON|WS_TABSTOP, IDOpen)
		createControl("BUTTON", "Pulisci", BS_PUSHBUTTON|WS_TABSTOP, IDClear)
		hwndInput = createControl("EDIT", "", WS_BORDER|WS_VSCROLL|WS_HSCROLL|ES_MULTILINE|ES_AUTOVSCROLL|ES_AUTOHSCROLL|ES_WANTRETURN|WS_TABSTOP, IDInput)
		hwndOperation = createControl("COMBOBOX", "", CBS_DROPDOWNLIST|CBS_HASSTRINGS|WS_VSCROLL|WS_TABSTOP, IDOperation)
		for _, op := range operations {
			pSendMessageW.Call(hwndOperation, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(wptr(op))))
		}
		pSendMessageW.Call(hwndOperation, CB_SETCURSEL, 0, 0)
		hwndParamLabel = createControl("STATIC", "Lettere da contare:", SS_LEFT, IDParamLabel)
		hwndParam = createControl("EDIT", "", WS_BORDER|ES_AUTOHSCROLL|WS_TABSTOP, IDParam)
		hwndCase = createControl("BUTTON", "Maiuscole/minuscole distinte", BS_AUTOCHECKBOX|WS_TABSTOP, IDCase)
		createControl("BUTTON", "KUNTA!", BS_DEFPUSHBUTTON|WS_TABSTOP, IDRun)
		hwndResult = createControl("EDIT", "", WS_BORDER|WS_VSCROLL|WS_HSCROLL|ES_MULTILINE|ES_AUTOVSCROLL|ES_AUTOHSCROLL|ES_READONLY, IDResult)
		createControl("BUTTON", "Copia risultato", BS_PUSHBUTTON|WS_TABSTOP, IDCopyResult)
		hwndBrandIcon = createControl("STATIC", "", SS_ICON, IDBrandIcon)
		hwndBrandText = createControl("STATIC", "ShiduLab", SS_RIGHT, IDBrandText)
		if hBrandIcon != 0 {
			pSendMessageW.Call(hwndBrandIcon, STM_SETICON, hBrandIcon, 0)
		}
		pSendMessageW.Call(hwndInput, EM_SETCUEBANNER, 1, uintptr(unsafe.Pointer(wptr("Incolla qui il testo, oppure usa Apri testo..."))))
		updateParamHint()
		pDragAcceptFiles.Call(hwnd, 1)
		layout()
		return 0
	case WM_SIZE:
		layout()
		return 0
	case WM_DROPFILES:
		droppedFile(wParam)
		return 0
	case WM_COMMAND:
		id := int(loword(wParam))
		code := int(hiword(wParam))
		switch id {
		case IDPaste:
			pSetFocus.Call(hwndInput)
			pSendMessageW.Call(hwndInput, WM_PASTE, 0, 0)
		case IDOpen:
			openTextFile()
		case IDClear:
			setText(hwndInput, "")
			setText(hwndResult, "")
			setText(hwndParam, "")
			pSetFocus.Call(hwndInput)
		case IDOperation:
			if code == CBN_SELCHANGE {
				updateParamHint()
			}
		case IDRun:
			runAnalysis()
		case IDCopyResult:
			pSendMessageW.Call(hwndResult, EM_SETSEL, 0, ^uintptr(0))
			pSendMessageW.Call(hwndResult, WM_COPY, 0, 0)
			pSendMessageW.Call(hwndResult, EM_SETSEL, ^uintptr(0), ^uintptr(0))
		}
		return 0
	case WM_DESTROY:
		pPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}

func main() {
	runtime.LockOSThread()
	// Identita stabile per taskbar/pinning di Windows.
	pSetCurrentProcessExplicitAppUserModelID.Call(uintptr(unsafe.Pointer(wptr("ShiduLab.Kunta"))))
	hInst, _, _ := pGetModuleHandleW.Call(0)
	className := wptr("ShiduLabKuntaWindowClass")
	cursor, _, _ := pLoadCursorW.Call(0, IDC_ARROW)
	hIconBig = createIconFromICO(kuntaICO, 32)
	hIconSmall = createIconFromICO(kuntaICO, 16)
	hBrandIcon = createIconFromICO(botoloICO, 32)
	wc := WNDCLASSEX{CbSize: uint32(unsafe.Sizeof(WNDCLASSEX{})), LpfnWndProc: syscall.NewCallback(wndProc), HInstance: hInst, HIcon: hIconBig, HCursor: cursor, HbrBackground: COLOR_WINDOW + 1, LpszClassName: className, HIconSm: hIconSmall}
	if r, _, _ := pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		panic("RegisterClassExW")
	}
	hwnd, _, _ := pCreateWindowExW.Call(0, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(wptr(appTitle))), WS_OVERLAPPED|WS_CAPTION|WS_SYSMENU|WS_THICKFRAME|WS_MINIMIZEBOX|WS_MAXIMIZEBOX|WS_VISIBLE, CW_USEDEFAULT, CW_USEDEFAULT, 980, 720, 0, 0, hInst, 0)
	if hwnd == 0 {
		panic("CreateWindowExW")
	}
	if hIconBig != 0 {
		pSendMessageW.Call(hwnd, WM_SETICON, ICON_BIG, hIconBig)
	}
	if hIconSmall != 0 {
		pSendMessageW.Call(hwnd, WM_SETICON, ICON_SMALL, hIconSmall)
	}
	pShowWindow.Call(hwnd, 5)
	pUpdateWindow.Call(hwnd)
	var m MSG
	for {
		r, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
	if hBrandIcon != 0 {
		pDestroyIcon.Call(hBrandIcon)
	}
	if hIconSmall != 0 {
		pDestroyIcon.Call(hIconSmall)
	}
	if hIconBig != 0 && hIconBig != hIconSmall {
		pDestroyIcon.Call(hIconBig)
	}
}
