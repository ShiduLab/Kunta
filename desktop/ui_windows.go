//go:build windows

package main

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	ole32    = syscall.NewLazyDLL("ole32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	uxtheme  = syscall.NewLazyDLL("uxtheme.dll")
	dwmapi   = syscall.NewLazyDLL("dwmapi.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")

	procRegisterClassExW     = user32.NewProc("RegisterClassExW")
	procCreateWindowExW      = user32.NewProc("CreateWindowExW")
	procDefWindowProcW       = user32.NewProc("DefWindowProcW")
	procGetMessageW          = user32.NewProc("GetMessageW")
	procTranslateMessage     = user32.NewProc("TranslateMessage")
	procDispatchMessageW     = user32.NewProc("DispatchMessageW")
	procPostQuitMessage      = user32.NewProc("PostQuitMessage")
	procMoveWindow           = user32.NewProc("MoveWindow")
	procSendMessageW         = user32.NewProc("SendMessageW")
	procPostMessageW         = user32.NewProc("PostMessageW")
	procSetWindowTextW       = user32.NewProc("SetWindowTextW")
	procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	procGetWindowTextW       = user32.NewProc("GetWindowTextW")
	procGetClientRect        = user32.NewProc("GetClientRect")
	procSetFocus             = user32.NewProc("SetFocus")
	procMessageBoxW          = user32.NewProc("MessageBoxW")
	procLoadCursorW          = user32.NewProc("LoadCursorW")
	procLoadIconW            = user32.NewProc("LoadIconW")
	procBeginPaint           = user32.NewProc("BeginPaint")
	procEndPaint             = user32.NewProc("EndPaint")
	procFillRect             = user32.NewProc("FillRect")
	procFrameRect            = user32.NewProc("FrameRect")
	procDrawTextW            = user32.NewProc("DrawTextW")
	procDrawIconEx           = user32.NewProc("DrawIconEx")
	procInvalidateRect       = user32.NewProc("InvalidateRect")
	procUpdateWindow         = user32.NewProc("UpdateWindow")
	procRedrawWindow         = user32.NewProc("RedrawWindow")
	procEnableWindow         = user32.NewProc("EnableWindow")
	procGetDpiForWindow      = user32.NewProc("GetDpiForWindow")

	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	procCreateFontW      = gdi32.NewProc("CreateFontW")
	procCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
	procCreatePen        = gdi32.NewProc("CreatePen")
	procDeleteObject     = gdi32.NewProc("DeleteObject")
	procSelectObject     = gdi32.NewProc("SelectObject")
	procRoundRect        = gdi32.NewProc("RoundRect")
	procSetTextColor     = gdi32.NewProc("SetTextColor")
	procSetBkColor       = gdi32.NewProc("SetBkColor")
	procSetBkMode        = gdi32.NewProc("SetBkMode")
	procStretchDIBits    = gdi32.NewProc("StretchDIBits")

	procCoInitializeEx        = ole32.NewProc("CoInitializeEx")
	procCoUninitialize        = ole32.NewProc("CoUninitialize")
	procCoCreateInstance      = ole32.NewProc("CoCreateInstance")
	procCoTaskMemFree         = ole32.NewProc("CoTaskMemFree")
	procSetWindowTheme        = uxtheme.NewProc("SetWindowTheme")
	procDwmSetWindowAttribute = dwmapi.NewProc("DwmSetWindowAttribute")
	procDragAcceptFiles       = shell32.NewProc("DragAcceptFiles")
	procDragQueryFileW        = shell32.NewProc("DragQueryFileW")
	procDragFinish            = shell32.NewProc("DragFinish")
	procShellExecuteW         = shell32.NewProc("ShellExecuteW")
	procSHChangeNotify        = shell32.NewProc("SHChangeNotify")
	procRegCreateKeyExW       = advapi32.NewProc("RegCreateKeyExW")
	procRegSetValueExW        = advapi32.NewProc("RegSetValueExW")
	procRegCloseKey           = advapi32.NewProc("RegCloseKey")
	procRegDeleteTreeW        = advapi32.NewProc("RegDeleteTreeW")
)

const (
	WS_OVERLAPPEDWINDOW  = 0x00CF0000
	WS_VISIBLE           = 0x10000000
	WS_CHILD             = 0x40000000
	WS_TABSTOP           = 0x00010000
	WS_VSCROLL           = 0x00200000
	WS_HSCROLL           = 0x00100000
	WS_BORDER            = 0x00800000
	WS_CLIPCHILDREN      = 0x02000000
	WS_EX_CLIENTEDGE     = 0x00000200
	ES_MULTILINE         = 0x0004
	ES_AUTOVSCROLL       = 0x0040
	ES_AUTOHSCROLL       = 0x0080
	ES_WANTRETURN        = 0x1000
	ES_READONLY          = 0x0800
	ES_NOHIDESEL         = 0x0100
	BS_OWNERDRAW         = 0x000B
	BS_AUTOCHECKBOX      = 0x0003
	CBS_DROPDOWNLIST     = 0x0003
	CBS_HASSTRINGS       = 0x0200
	CBS_OWNERDRAWFIXED   = 0x0010
	CBS_NOINTEGRALHEIGHT = 0x0400
	SS_LEFT              = 0x0000

	WM_DESTROY           = 0x0002
	WM_SIZE              = 0x0005
	WM_PAINT             = 0x000F
	WM_ERASEBKGND        = 0x0014
	WM_GETMINMAXINFO     = 0x0024
	WM_COMMAND           = 0x0111
	WM_DRAWITEM          = 0x002B
	WM_MEASUREITEM       = 0x002C
	WM_SETFONT           = 0x0030
	WM_CTLCOLORMSGBOX    = 0x0132
	WM_CTLCOLOREDIT      = 0x0133
	WM_CTLCOLORLISTBOX   = 0x0134
	WM_CTLCOLORBTN       = 0x0135
	WM_CTLCOLORDLG       = 0x0136
	WM_CTLCOLORSCROLLBAR = 0x0137
	WM_CTLCOLORSTATIC    = 0x0138
	WM_DROPFILES         = 0x0233
	WM_PASTE             = 0x0302
	WM_COPY              = 0x0301
	WM_SETREDRAW         = 0x000B
	WM_APP_ANALYSIS_DONE = 0x8001

	EM_SETSEL          = 0x00B1
	EM_SETLIMITTEXT    = 0x00C5
	EM_SETMARGINS      = 0x00D3
	EC_LEFTMARGIN      = 0x0001
	EC_RIGHTMARGIN     = 0x0002
	CB_ADDSTRING       = 0x0143
	CB_GETCOUNT        = 0x0146
	CB_GETLBTEXT       = 0x0148
	CB_GETLBTEXTLEN    = 0x0149
	CB_GETCURSEL       = 0x0147
	CB_SETCURSEL       = 0x014E
	CB_SETITEMHEIGHT   = 0x0153
	CB_SETMINVISIBLE   = 0x1701
	CB_SHOWDROPDOWN    = 0x014F
	CB_SETDROPPEDWIDTH = 0x0160
	BM_GETCHECK        = 0x00F0
	BST_CHECKED        = 1
	CW_USEDEFAULT      = 0x80000000

	ODT_BUTTON       = 4
	ODT_COMBOBOX     = 3
	ODS_SELECTED     = 0x0001
	ODS_FOCUS        = 0x0010
	ODS_COMBOBOXEDIT = 0x1000
	DT_LEFT          = 0x0000
	DT_CENTER        = 0x0001
	DT_VCENTER       = 0x0004
	DT_BOTTOM        = 0x0008
	DT_SINGLELINE    = 0x0020
	DT_END_ELLIPSIS  = 0x8000
	TRANSPARENT      = 1
	OPAQUE           = 2
	PS_SOLID         = 0
	DI_NORMAL        = 0x0003

	RDW_INVALIDATE = 0x0001
	RDW_ERASE      = 0x0004
	RDW_UPDATENOW  = 0x0100
	RDW_FRAME      = 0x0400

	HKEY_CURRENT_USER  = 0x80000001
	KEY_WRITE          = 0x20006
	REG_SZ             = 1
	SHCNE_ASSOCCHANGED = 0x08000000

	ID_TITLE           = 90
	ID_TAGLINE         = 91
	ID_EDIT_INPUT      = 101
	ID_PASTE           = 102
	ID_OPEN            = 103
	ID_CLEAR           = 104
	ID_QUICK_AZ        = 105
	ID_QUICK_REP       = 106
	ID_QUICK_KUNTA     = 107
	ID_QUICK_VOW       = 108
	ID_QUICK_CONS      = 109
	ID_COMBO           = 110
	ID_PARAM           = 111
	ID_CASE            = 112
	ID_RUN             = 113
	ID_HINT            = 114
	ID_OUTPUT          = 115
	ID_COPY            = 116
	ID_STATUS          = 117
	ID_LABEL_OPERATION = 118
	ID_LABEL_PARAM     = 119
	ID_FOOTER          = 120
)

const repoURL = "https://github.com/ShiduLab/Kunta"

type POINT struct{ X, Y int32 }
type MSG struct {
	Hwnd           uintptr
	Message        uint32
	WParam, LParam uintptr
	Time           uint32
	Pt             POINT
}
type RECT struct{ Left, Top, Right, Bottom int32 }
type WNDCLASSEX struct {
	CbSize                                   uint32
	Style                                    uint32
	LpfnWndProc                              uintptr
	CbClsExtra, CbWndExtra                   int32
	HInstance, HIcon, HCursor, HbrBackground uintptr
	LpszMenuName, LpszClassName              *uint16
	HIconSm                                  uintptr
}
type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}
type COMDLG_FILTERSPEC struct {
	Name *uint16
	Spec *uint16
}
type BITMAPINFOHEADER struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}
type BITMAPINFO struct {
	Header BITMAPINFOHEADER
	Colors [1]uint32
}
type PAINTSTRUCT struct {
	Hdc         uintptr
	FErase      int32
	RcPaint     RECT
	FRestore    int32
	FIncUpdate  int32
	RgbReserved [32]byte
}
type DRAWITEMSTRUCT struct {
	CtlType    uint32
	CtlID      uint32
	ItemID     uint32
	ItemAction uint32
	ItemState  uint32
	HwndItem   uintptr
	HDC        uintptr
	RcItem     RECT
	ItemData   uintptr
}
type MEASUREITEMSTRUCT struct {
	CtlType, CtlID, ItemID, ItemWidth, ItemHeight uint32
	ItemData                                      uintptr
}
type MINMAXINFO struct {
	PtReserved, PtMaxSize, PtMaxPosition, PtMinTrackSize, PtMaxTrackSize POINT
}

func rgb(r, g, b byte) uintptr { return uintptr(r) | uintptr(g)<<8 | uintptr(b)<<16 }

var (
	hwndMain                                                                          uintptr
	hTitle, hTagline, hInput, hParam, hCombo, hCase, hHint, hOutput, hStatus, hFooter uintptr
	appIcon                                                                           uintptr
	controlsByID                                                                      = map[int]uintptr{}
	panelStatics                                                                      = map[uintptr]bool{}

	fontUI, fontSmall, fontTitle, fontMono, fontBold                   uintptr
	brushBG, brushPanel, brushInput, brushAccent, brushSoft, brushLine uintptr

	colBG        = rgb(16, 20, 15)
	colPanel     = rgb(23, 28, 22)
	colInput     = rgb(18, 24, 18)
	colText      = rgb(238, 242, 235)
	colMuted     = rgb(174, 184, 170)
	colAccent    = rgb(124, 186, 47)
	colAccentInk = rgb(13, 22, 6)
	colSoft      = rgb(32, 42, 29)
	colLine      = rgb(57, 67, 55)

	controlPanel  RECT
	currentDPI    = 96
	busyMu        sync.Mutex
	busy          bool
	resultMu      sync.Mutex
	pendingResult string

	botoloPixels []byte
	botoloWidth  int
	botoloHeight int
	botoloStride int
)

// Botolo originale della PWA, senza riquadri colorati; il fondo coincide con quello della UI.
const botoloBMPBase64 = `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DxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQERYSDxQQOD094+XmZWRjLy0tNzU1MzExMjAwMS8vMC4uLy0tLy0tLy0tLSsrLCoqKykpKykpLCkpLCkoKykoMi0sMi0sJyIhJyEgQT4+a2trk5WVrbCwt7u6u769u729trm3rK6umZmZd3d2SkhHKyYlKSMhMCspKiYiKCIfIx4dLiopLywqMC0rNDAuLykmQT06x8nI6OvrZ2pnGh4cDxQQDxQQEBURDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQEBURDxQQFRoYztPUn5+eKykpOjg4MzExMjAwMC4vLy0tLy0tLiwsLy0tLSsrLSsrLSsrKikpKigoKigoKygoLCkoLikoMCsqMi0sLCYlJB4dIh0cKCIhLysqNjMyODQzMzAvKygnJB4dIhwbJyEeLikmLyooKyYkPTk4S0dFZWJiLSgmLignLiooLispNC8tLyonPDc0p6al7/LyzdDQkJOSLTIvDxQQEBURDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQERYSDxQQhIeJ4+XlQkBANTIzNTQ0MjAwMS4vLy0tLy0tLy0tLiwsLSsrLSorKykpKykpKScnKignKignKygnLCgnLCgnLScnLSgpMCsqMCoqLykoLCcmKyUkKiUkKyYlLScmLyopLyopLSgmLigmJSAeYV5eo6KikZCQjIqLc3FwIRwZMCsoLignLyoqNC8sMi0pLSckYl5bnJua4ePiyMnKIiYiDxQQEBURDxQQDxQQDxQQDxQQDxQQDxQQDxQQERYSDxQQKi8t3uHjnZ6dKScnOjg4MzEyMjAwMC4uLy0tLy0tLiwsLCoqKykpKykpKigoKScnKSgoKicnKigoKigoLCcoKycmKScmKicmKycmLCcmLCcmLSgnLScnLCcmLCcmLCclLCcmKyYlLScnJSAfg4OCj42OmpqajIuKX1taIh0cLyoqLCclLikmLyonMi0qNjAtMyomIxkVeHZ18vX2TVBODxQQERYSDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQEhcTDxQQcXZ49fj2dHNzKCYmOTY2NDIyMTAwMC4uLy0tLy0tLSsrKykpKigoKigoKScnKScnKScnKScnKicnKyYnKSYmKSYlKSYlKyYlKiUkKSUkKSUlKiUkKiUkKiUkKiUk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x8fG2drZ9fj32NrbWVtbDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxMPEBUSDxQQEBURDxQQDxQQEBUSEhcTDxQQDxQQDxQQDxQQDxQQDhMPDhMPDxMQDxQQDxQQDxQQDxQQGBwaTlFQkJSTtbi5u77Aq6+whYaHNjk4DxQQERYSEBURDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQERYSDxQQDxQQEhcTDxQQDxQQDxQQDxQQDxQQDxQRERYSEhcTEhcTDxQQDxQQDxQQEBUREhcTEhcTEBUREhcTDxQQDxQQDxQQDhMPDhMQDhMPDxQQDxQQERYSDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQEBUREBURDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQREBURDxQQDxQQDxQQDxQQDxQQEBUREhcTEBURDxQQDxQQDxQQEBURERcTDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQDxQQ`

var botoloBMP = func() []byte {
	b, err := base64.StdEncoding.DecodeString(botoloBMPBase64)
	if err != nil {
		return nil
	}
	return b
}()

var (
	clsidFileOpenDialog = GUID{0xDC1C5A9C, 0xE88A, 0x4DDE, [8]byte{0xA5, 0xA1, 0x60, 0xF8, 0x2A, 0x20, 0xAE, 0xF7}}
	iidIFileOpenDialog  = GUID{0xD57C7288, 0xD4AD, 0x4768, [8]byte{0xBE, 0x02, 0x9D, 0x96, 0x95, 0x32, 0xD9, 0x60}}
)

func u16(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }
func textOf(h uintptr) string {
	n, _, _ := procGetWindowTextLengthW.Call(h)
	if n == 0 {
		return ""
	}
	buf := make([]uint16, int(n)+1)
	procGetWindowTextW.Call(h, uintptr(unsafe.Pointer(&buf[0])), n+1)
	return syscall.UTF16ToString(buf)
}
func windowsText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.ReplaceAll(s, "\n", "\r\n")
}
func setText(h uintptr, s string) {
	procSetWindowTextW.Call(h, uintptr(unsafe.Pointer(u16(windowsText(s)))))
}
func replaceOutput(s string) {
	if hOutput == 0 {
		return
	}
	procSendMessageW.Call(hOutput, WM_SETREDRAW, 0, 0)
	procSetWindowTextW.Call(hOutput, uintptr(unsafe.Pointer(u16(""))))
	if s != "" {
		procSetWindowTextW.Call(hOutput, uintptr(unsafe.Pointer(u16(windowsText(s)))))
	}
	procSendMessageW.Call(hOutput, EM_SETSEL, 0, 0)
	procSendMessageW.Call(hOutput, WM_SETREDRAW, 1, 0)
	procRedrawWindow.Call(hOutput, 0, 0, RDW_INVALIDATE|RDW_ERASE|RDW_FRAME|RDW_UPDATENOW)
}
func move(h uintptr, x, y, w, hh int) {
	if h != 0 && w > 0 && hh > 0 {
		procMoveWindow.Call(h, uintptr(x), uintptr(y), uintptr(w), uintptr(hh), 1)
	}
}
func message(s string) {
	procMessageBoxW.Call(hwndMain, uintptr(unsafe.Pointer(u16(s))), uintptr(unsafe.Pointer(u16("Kunta"))), 0x40)
}
func scale(v int) int { return v * currentDPI / 96 }
func setDarkTheme(h uintptr) {
	if h == 0 {
		return
	}
	procSetWindowTheme.Call(h, uintptr(unsafe.Pointer(u16("DarkMode_Explorer"))), 0)
}
func setImmersiveDarkTitlebar(hwnd uintptr) {
	on := int32(1)
	// 20 is DWMWA_USE_IMMERSIVE_DARK_MODE on current Windows 10/11; 19 on older builds.
	if r, _, _ := procDwmSetWindowAttribute.Call(hwnd, 20, uintptr(unsafe.Pointer(&on)), unsafe.Sizeof(on)); r != 0 {
		procDwmSetWindowAttribute.Call(hwnd, 19, uintptr(unsafe.Pointer(&on)), unsafe.Sizeof(on))
	}
}
func fontForID(id int) uintptr {
	switch id {
	case ID_TITLE:
		return fontTitle
	case ID_TAGLINE, ID_HINT, ID_STATUS, ID_LABEL_OPERATION, ID_LABEL_PARAM, ID_FOOTER:
		return fontSmall
	case ID_EDIT_INPUT, ID_OUTPUT:
		return fontMono
	case ID_RUN:
		return fontBold
	default:
		return fontUI
	}
}
func create(class, text string, style, ex uint32, id int) uintptr {
	r, _, _ := procCreateWindowExW.Call(uintptr(ex), uintptr(unsafe.Pointer(u16(class))), uintptr(unsafe.Pointer(u16(text))), uintptr(style), 0, 0, 10, 10, hwndMain, uintptr(id), 0, 0)
	if r != 0 {
		controlsByID[id] = r
		if f := fontForID(id); f != 0 {
			procSendMessageW.Call(r, WM_SETFONT, f, 1)
		}
		if id != ID_COMBO && id != ID_OUTPUT {
			setDarkTheme(r)
		}
	}
	return r
}
func addButton(text string, id int) uintptr {
	return create("BUTTON", text, WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_OWNERDRAW, 0, id)
}
func addStatic(text string, id int, panel bool) uintptr {
	h := create("STATIC", text, WS_CHILD|WS_VISIBLE|SS_LEFT, 0, id)
	if panel {
		panelStatics[h] = true
	}
	return h
}

var textExtensions = []string{
	".txt", ".md", ".markdown", ".log", ".csv", ".tsv", ".json", ".xml",
	".html", ".htm", ".css", ".js", ".ts", ".py", ".go", ".java", ".kt",
	".c", ".cpp", ".h", ".hpp", ".ini", ".cfg", ".conf", ".yaml", ".yml", ".toml",
}

func regSetString(root uintptr, subkey, name, value string) bool {
	var h uintptr
	sk := u16(subkey)
	r, _, _ := procRegCreateKeyExW.Call(root, uintptr(unsafe.Pointer(sk)), 0, 0, 0, KEY_WRITE, 0, uintptr(unsafe.Pointer(&h)), 0)
	if r != 0 || h == 0 {
		return false
	}
	defer procRegCloseKey.Call(h)
	data := syscall.StringToUTF16(value)
	var namePtr uintptr
	if name != "" {
		n := u16(name)
		namePtr = uintptr(unsafe.Pointer(n))
	}
	r, _, _ = procRegSetValueExW.Call(h, namePtr, 0, REG_SZ, uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)*2))
	return r == 0
}

func registerShellIntegration() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	exe, _ = filepath.Abs(exe)
	cmd := fmt.Sprintf("\"%s\" \"%%1\"", exe)
	icon := fmt.Sprintf("\"%s\",0", exe)
	ok := true
	set := func(k, n, v string) {
		if !regSetString(HKEY_CURRENT_USER, k, n, v) {
			ok = false
		}
	}

	set(`Software\Microsoft\Windows\CurrentVersion\App Paths\Kunta.exe`, "", exe)
	app := `Software\Classes\Applications\Kunta.exe`
	set(app, "FriendlyAppName", "Kunta")
	set(app+`\DefaultIcon`, "", icon)
	set(app+`\shell\open\command`, "", cmd)
	for _, ext := range textExtensions {
		set(app+`\SupportedTypes`, ext, "")
	}

	// Verbo diretto per i file percepiti come testo.
	base := `Software\Classes\SystemFileAssociations\text\shell\Kunta`
	set(base, "MUIVerb", "Apri in Kunta")
	set(base, "Icon", icon)
	set(base+`\command`, "", cmd)

	// Alcune estensioni non dichiarano PerceivedType=text: registrale senza cambiare l'app predefinita.
	for _, ext := range textExtensions {
		b := `Software\Classes\SystemFileAssociations\` + ext + `\shell\Kunta`
		set(b, "MUIVerb", "Apri in Kunta")
		set(b, "Icon", icon)
		set(b+`\command`, "", cmd)
	}
	procSHChangeNotify.Call(SHCNE_ASSOCCHANGED, 0, 0, 0)
	return ok
}

func unregisterShellIntegration() {
	keys := []string{
		`Software\Classes\SystemFileAssociations\text\shell\Kunta`,
		`Software\Classes\Applications\Kunta.exe`,
		`Software\Microsoft\Windows\CurrentVersion\App Paths\Kunta.exe`,
	}
	for _, ext := range textExtensions {
		keys = append(keys, `Software\Classes\SystemFileAssociations\`+ext+`\shell\Kunta`)
	}
	for _, k := range keys {
		p := u16(k)
		procRegDeleteTreeW.Call(HKEY_CURRENT_USER, uintptr(unsafe.Pointer(p)))
	}
	procSHChangeNotify.Call(SHCNE_ASSOCCHANGED, 0, 0, 0)
}

func safeUI(label string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			setBusy(false)
			message(fmt.Sprintf("%s: errore interno intercettato.\n%v", label, r))
		}
	}()
	fn()
}

func decodeTextBytes(b []byte) string {
	if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		return string(b[3:])
	}
	if len(b) >= 2 && ((b[0] == 0xFF && b[1] == 0xFE) || (b[0] == 0xFE && b[1] == 0xFF)) {
		le := b[0] == 0xFF
		b = b[2:]
		u := make([]uint16, 0, len(b)/2)
		for i := 0; i+1 < len(b); i += 2 {
			if le {
				u = append(u, binary.LittleEndian.Uint16(b[i:i+2]))
			} else {
				u = append(u, binary.BigEndian.Uint16(b[i:i+2]))
			}
		}
		return string(utf16.Decode(u))
	}
	return strings.ToValidUTF8(string(b), "�")
}
func loadTextPath(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		message("Impossibile leggere il file.")
		return
	}
	setText(hInput, decodeTextBytes(b))
	setText(hStatus, fmt.Sprintf("Testo caricato · %.1f KB", float64(len(b))/1024.0))
	procSetFocus.Call(hInput)
}
func hresultFailed(hr uintptr) bool { return int32(uint32(hr)) < 0 }
func comCall(obj uintptr, method int, args ...uintptr) uintptr {
	if obj == 0 {
		return uintptr(0x80004003)
	}
	vtbl := *(*uintptr)(unsafe.Pointer(obj))
	fn := *(*uintptr)(unsafe.Pointer(vtbl + uintptr(method)*unsafe.Sizeof(uintptr(0))))
	callArgs := make([]uintptr, 0, len(args)+1)
	callArgs = append(callArgs, obj)
	callArgs = append(callArgs, args...)
	r, _, _ := syscall.SyscallN(fn, callArgs...)
	return r
}
func utf16PtrToString(p uintptr) string {
	if p == 0 {
		return ""
	}
	a := (*[1 << 28]uint16)(unsafe.Pointer(p))
	n := 0
	for n < len(a) && a[n] != 0 {
		n++
	}
	return syscall.UTF16ToString(a[:n])
}
func openTextFile() {
	var dlg uintptr
	hr, _, _ := procCoCreateInstance.Call(
		uintptr(unsafe.Pointer(&clsidFileOpenDialog)),
		0,
		1, // CLSCTX_INPROC_SERVER
		uintptr(unsafe.Pointer(&iidIFileOpenDialog)),
		uintptr(unsafe.Pointer(&dlg)),
	)
	if hresultFailed(hr) || dlg == 0 {
		message(fmt.Sprintf("Impossibile aprire il selettore file.\nHRESULT: 0x%08X", uint32(hr)))
		return
	}
	defer comCall(dlg, 2) // IUnknown::Release

	n1, s1 := syscall.StringToUTF16("Testi"), syscall.StringToUTF16("*.txt;*.md;*.log;*.csv;*.tsv")
	n2, s2 := syscall.StringToUTF16("Tutti i file"), syscall.StringToUTF16("*.*")
	filters := []COMDLG_FILTERSPEC{{&n1[0], &s1[0]}, {&n2[0], &s2[0]}}
	comCall(dlg, 4, uintptr(len(filters)), uintptr(unsafe.Pointer(&filters[0]))) // SetFileTypes
	comCall(dlg, 9, uintptr(0x40|0x800|0x1000))                                  // FOS_FORCEFILESYSTEM | PATHMUSTEXIST | FILEMUSTEXIST
	title := syscall.StringToUTF16("Apri testo in Kunta")
	comCall(dlg, 17, uintptr(unsafe.Pointer(&title[0]))) // SetTitle

	hr = comCall(dlg, 3, hwndMain) // IModalWindow::Show
	if hresultFailed(hr) {
		// 0x800704C7 = annullato dall'utente: non è un errore.
		if uint32(hr) != 0x800704C7 {
			message(fmt.Sprintf("Il selettore file non si è aperto.\nHRESULT: 0x%08X", uint32(hr)))
		}
		return
	}

	var item uintptr
	hr = comCall(dlg, 20, uintptr(unsafe.Pointer(&item))) // IFileDialog::GetResult
	if hresultFailed(hr) || item == 0 {
		return
	}
	defer comCall(item, 2)

	var pathPtr uintptr
	hr = comCall(item, 5, uintptr(0x80058000), uintptr(unsafe.Pointer(&pathPtr))) // IShellItem::GetDisplayName(SIGDN_FILESYSPATH)
	if hresultFailed(hr) || pathPtr == 0 {
		return
	}
	path := utf16PtrToString(pathPtr)
	procCoTaskMemFree.Call(pathPtr)
	if path != "" {
		loadTextPath(path)
	}
}
func handleDrop(hDrop uintptr) {
	n, _, _ := procDragQueryFileW.Call(hDrop, 0xFFFFFFFF, 0, 0)
	if n > 0 {
		ln, _, _ := procDragQueryFileW.Call(hDrop, 0, 0, 0)
		buf := make([]uint16, int(ln)+1)
		procDragQueryFileW.Call(hDrop, 0, uintptr(unsafe.Pointer(&buf[0])), ln+1)
		loadTextPath(syscall.UTF16ToString(buf))
	}
	procDragFinish.Call(hDrop)
}
func updateHint() {
	i, _, _ := procSendMessageW.Call(hCombo, CB_GETCURSEL, 0, 0)
	if int(i) >= 0 && int(i) < len(operations) {
		setText(hHint, operations[int(i)].Hint)
	}
}
func setBusy(v bool) {
	busyMu.Lock()
	busy = v
	busyMu.Unlock()
	ids := []int{ID_RUN, ID_QUICK_AZ, ID_QUICK_REP, ID_QUICK_KUNTA, ID_QUICK_VOW, ID_QUICK_CONS}
	enabled := uintptr(1)
	if v {
		enabled = 0
	}
	for _, id := range ids {
		procEnableWindow.Call(controlsByID[id], enabled)
	}
	procEnableWindow.Call(hCombo, enabled)
	if v {
		setText(hStatus, "Kunta sta lavorando…")
	} else {
		setText(hStatus, fmt.Sprintf("Locale · %d operazioni · menu destro attivo · nessun invio esterno", len(operations)))
	}
}
func runOp(op int) {
	busyMu.Lock()
	already := busy
	busyMu.Unlock()
	if already {
		return
	}
	txt := textOf(hInput)
	if txt == "" {
		replaceOutput("Non c’è testo da analizzare.")
		return
	}
	param := textOf(hParam)
	checked, _, _ := procSendMessageW.Call(hCase, BM_GETCHECK, 0, 0)
	sensitive := checked == BST_CHECKED
	replaceOutput("")
	setBusy(true)
	go func() {
		res := ""
		defer func() {
			if r := recover(); r != nil {
				res = fmt.Sprintf("Errore interno durante l’analisi: %v", r)
			}
			resultMu.Lock()
			pendingResult = res
			resultMu.Unlock()
			procPostMessageW.Call(hwndMain, WM_APP_ANALYSIS_DONE, 0, 0)
		}()
		res = analyze(txt, op, param, sensitive)
	}()
}

func layout() {
	var r RECT
	procGetClientRect.Call(hwndMain, uintptr(unsafe.Pointer(&r)))
	w, h := int(r.Right-r.Left), int(r.Bottom-r.Top)
	pad, gap := scale(14), scale(9)
	if w < scale(820) {
		w = scale(820)
	}
	if h < scale(680) {
		h = scale(680)
	}

	y := scale(13)
	iconW := scale(50)
	move(hTitle, pad+iconW+scale(10), y+scale(2), scale(320), scale(31))
	move(hTagline, pad+iconW+scale(10), y+scale(31), scale(440), scale(20))
	y += scale(62)

	btnH := scale(40)
	move(controlsByID[ID_PASTE], pad, y, scale(105), btnH)
	move(controlsByID[ID_OPEN], pad+scale(114), y, scale(138), btnH)
	move(controlsByID[ID_CLEAR], pad+scale(261), y, scale(105), btnH)
	y += btnH + gap

	bottomReserve := scale(385)
	inputH := (h - y - bottomReserve) * 46 / 100
	if inputH < scale(135) {
		inputH = scale(135)
	}
	move(hInput, pad, y, w-2*pad, inputH)
	y += inputH + gap

	quickW := (w - 2*pad - 4*gap) / 5
	qs := []int{ID_QUICK_AZ, ID_QUICK_REP, ID_QUICK_KUNTA, ID_QUICK_VOW, ID_QUICK_CONS}
	x := pad
	for _, id := range qs {
		move(controlsByID[id], x, y, quickW, btnH)
		x += quickW + gap
	}
	y += btnH + gap

	panelX, panelY := pad, y
	panelW, panelH := w-2*pad, scale(112)
	controlPanel = RECT{int32(panelX), int32(panelY), int32(panelX + panelW), int32(panelY + panelH)}
	inner := scale(10)
	labelH := scale(18)
	comboW := panelW * 46 / 100
	paramW := panelW * 18 / 100
	rightX := panelX + comboW + paramW + 2*gap
	rightW := panelX + panelW - inner - rightX

	move(controlsByID[ID_LABEL_OPERATION], panelX+inner, panelY+scale(8), comboW-inner*2, labelH)
	move(controlsByID[ID_LABEL_PARAM], panelX+comboW+gap+inner, panelY+scale(8), paramW-inner*2, labelH)
	fieldY := panelY + scale(31)
	move(hCombo, panelX+inner, fieldY, comboW-inner*2, scale(540))
	move(hParam, panelX+comboW+gap+inner, fieldY, paramW-inner*2, scale(36))
	runW := scale(105)
	caseW := rightW - runW - gap
	if caseW < scale(190) {
		caseW = scale(190)
	}
	move(hCase, rightX, fieldY, caseW, scale(36))
	move(controlsByID[ID_RUN], panelX+panelW-inner-runW, fieldY, runW, scale(36))
	move(hHint, panelX+inner, panelY+scale(75), panelW-inner*2, scale(25))

	y = panelY + panelH + gap
	footerH := scale(54)
	outH := h - y - footerH - pad
	if outH < scale(125) {
		outH = scale(125)
	}
	move(hOutput, pad, y, w-2*pad, outH)
	y += outH + gap
	move(controlsByID[ID_COPY], pad, y, scale(155), scale(36))
	footerW := scale(190)
	footerX := w - pad - footerW
	statusX := pad + scale(169)
	statusW := footerX - statusX - gap
	if statusW < scale(120) {
		statusW = scale(120)
	}
	move(hStatus, statusX, y+scale(8), statusW, scale(24))
	move(hFooter, footerX, y, footerW, scale(46))
	procInvalidateRect.Call(hwndMain, 0, 0)
}

func buttonStyle(id uint32) (bg, fg uintptr) {
	switch id {
	case ID_RUN:
		return colAccent, colAccentInk
	case ID_QUICK_AZ, ID_QUICK_REP, ID_QUICK_KUNTA, ID_QUICK_VOW, ID_QUICK_CONS:
		return colSoft, colText
	default:
		return colPanel, colText
	}
}
func drawRoundedBox(hdc uintptr, rc RECT, bg, border uintptr, radius int) {
	b, _, _ := procCreateSolidBrush.Call(bg)
	p, _, _ := procCreatePen.Call(PS_SOLID, 1, border)
	oldB, _, _ := procSelectObject.Call(hdc, b)
	oldP, _, _ := procSelectObject.Call(hdc, p)
	procRoundRect.Call(hdc, uintptr(rc.Left), uintptr(rc.Top), uintptr(rc.Right), uintptr(rc.Bottom), uintptr(radius), uintptr(radius))
	procSelectObject.Call(hdc, oldB)
	procSelectObject.Call(hdc, oldP)
	procDeleteObject.Call(b)
	procDeleteObject.Call(p)
}
func drawBotoloAt(hdc uintptr, x, y, dw, dh int) {
	if len(botoloPixels) == 0 || botoloWidth <= 0 || botoloHeight <= 0 {
		return
	}
	bmi := BITMAPINFO{Header: BITMAPINFOHEADER{
		Size: uint32(unsafe.Sizeof(BITMAPINFOHEADER{})), Width: int32(botoloWidth), Height: int32(botoloHeight),
		Planes: 1, BitCount: 24, Compression: 0, SizeImage: uint32(botoloStride * botoloHeight),
	}}
	procStretchDIBits.Call(hdc, uintptr(x), uintptr(y), uintptr(dw), uintptr(dh), 0, 0,
		uintptr(botoloWidth), uintptr(botoloHeight), uintptr(unsafe.Pointer(&botoloPixels[0])), uintptr(unsafe.Pointer(&bmi)), 0, 0x00CC0020)
}

func drawOwnerButton(di *DRAWITEMSTRUCT) {
	if di.CtlID == ID_FOOTER {
		bg := colBG
		if di.ItemState&ODS_SELECTED != 0 {
			bg = colSoft
		}
		b, _, _ := procCreateSolidBrush.Call(bg)
		procFillRect.Call(di.HDC, uintptr(unsafe.Pointer(&di.RcItem)), b)
		procDeleteObject.Call(b)

		icon := scale(44)
		x := int(di.RcItem.Left) + scale(5)
		y := int(di.RcItem.Bottom) - icon - scale(1)
		drawBotoloAt(di.HDC, x, y, icon, icon)

		procSetBkMode.Call(di.HDC, TRANSPARENT)
		nameR := di.RcItem
		nameR.Left = int32(x + icon + scale(9))
		nameR.Top = int32(y)
		nameR.Bottom = int32(y + icon)
		nameR.Right -= int32(scale(4))
		oldFont, _, _ := procSelectObject.Call(di.HDC, fontBold)
		procSetTextColor.Call(di.HDC, colText)
		procDrawTextW.Call(di.HDC, uintptr(unsafe.Pointer(u16("ShiduLab"))), ^uintptr(0), uintptr(unsafe.Pointer(&nameR)), DT_LEFT|DT_BOTTOM|DT_SINGLELINE)
		if oldFont != 0 {
			procSelectObject.Call(di.HDC, oldFont)
		}
		return
	}
	bg, fg := buttonStyle(di.CtlID)
	if di.ItemState&ODS_SELECTED != 0 {
		if di.CtlID == ID_RUN {
			bg = rgb(103, 160, 37)
		} else {
			bg = rgb(43, 53, 40)
		}
	}
	drawRoundedBox(di.HDC, di.RcItem, bg, colLine, scale(7))
	procSetBkMode.Call(di.HDC, TRANSPARENT)
	procSetTextColor.Call(di.HDC, fg)
	txt := textOf(di.HwndItem)
	rr := di.RcItem
	procDrawTextW.Call(di.HDC, uintptr(unsafe.Pointer(u16(txt))), ^uintptr(0), uintptr(unsafe.Pointer(&rr)), DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)
}
func comboItemText(hwnd uintptr, itemID uint32) string {
	if itemID == 0xFFFFFFFF {
		i, _, _ := procSendMessageW.Call(hwnd, CB_GETCURSEL, 0, 0)
		if int(i) < 0 {
			return ""
		}
		itemID = uint32(i)
	}
	ln, _, _ := procSendMessageW.Call(hwnd, CB_GETLBTEXTLEN, uintptr(itemID), 0)
	if int32(ln) < 0 {
		return ""
	}
	buf := make([]uint16, int(ln)+1)
	procSendMessageW.Call(hwnd, CB_GETLBTEXT, uintptr(itemID), uintptr(unsafe.Pointer(&buf[0])))
	return syscall.UTF16ToString(buf)
}
func drawOwnerCombo(di *DRAWITEMSTRUCT) {
	bg := colPanel
	if di.ItemState&ODS_SELECTED != 0 {
		bg = colSoft
	}
	b, _, _ := procCreateSolidBrush.Call(bg)
	procFillRect.Call(di.HDC, uintptr(unsafe.Pointer(&di.RcItem)), b)
	procDeleteObject.Call(b)
	procSetBkMode.Call(di.HDC, TRANSPARENT)
	procSetTextColor.Call(di.HDC, colText)
	rr := di.RcItem
	rr.Left += int32(scale(9))
	rr.Right -= int32(scale(6))
	txt := comboItemText(di.HwndItem, di.ItemID)
	procDrawTextW.Call(di.HDC, uintptr(unsafe.Pointer(u16(txt))), ^uintptr(0), uintptr(unsafe.Pointer(&rr)), DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)
}
func initBotolo() {
	if len(botoloBMP) < 54 || string(botoloBMP[:2]) != "BM" {
		return
	}
	off := int(binary.LittleEndian.Uint32(botoloBMP[10:14]))
	w := int(int32(binary.LittleEndian.Uint32(botoloBMP[18:22])))
	h := int(int32(binary.LittleEndian.Uint32(botoloBMP[22:26])))
	bpp := int(binary.LittleEndian.Uint16(botoloBMP[28:30]))
	if off <= 0 || w <= 0 || h == 0 || bpp != 24 {
		return
	}
	if h < 0 {
		h = -h
	}
	stride := ((w*3 + 3) / 4) * 4
	need := off + stride*h
	if need > len(botoloBMP) {
		return
	}
	pix := append([]byte(nil), botoloBMP[off:need]...)
	botoloPixels, botoloWidth, botoloHeight, botoloStride = pix, w, h, stride
}

func paintBackground(hwnd uintptr) {
	var ps PAINTSTRUCT
	hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	if hdc != 0 {
		var r RECT
		procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
		procFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), brushBG)
		if controlPanel.Right > controlPanel.Left {
			drawRoundedBox(hdc, controlPanel, colPanel, colLine, scale(9))
		}
		if appIcon != 0 {
			procDrawIconEx.Call(hdc, uintptr(scale(14)), uintptr(scale(14)), appIcon, uintptr(scale(48)), uintptr(scale(48)), 0, 0, DI_NORMAL)
		}
	}
	procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
}
func applyDPI(hwnd uintptr) {
	if dpi, _, _ := procGetDpiForWindow.Call(hwnd); dpi >= 72 && dpi <= 384 {
		currentDPI = int(dpi)
	}
}
func setControlFonts() {
	h := func(px int, weight int, face string) uintptr {
		height := -scale(px)
		r, _, _ := procCreateFontW.Call(uintptr(int64(height)), 0, 0, 0, uintptr(weight), 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(u16(face))))
		return r
	}
	fontUI = h(17, 400, "Segoe UI")
	fontSmall = h(14, 400, "Segoe UI")
	fontTitle = h(27, 600, "Segoe UI")
	fontMono = h(15, 400, "Consolas")
	fontBold = h(17, 700, "Segoe UI")
}

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	case WM_ERASEBKGND:
		return 1
	case WM_GETMINMAXINFO:
		if lParam != 0 {
			m := (*MINMAXINFO)(unsafe.Pointer(lParam))
			m.PtMinTrackSize.X = int32(scale(820))
			m.PtMinTrackSize.Y = int32(scale(680))
			return 0
		}
	case WM_SIZE:
		layout()
		return 0
	case WM_PAINT:
		paintBackground(hwnd)
		return 0
	case WM_DROPFILES:
		handleDrop(wParam)
		return 0
	case WM_DRAWITEM:
		if lParam != 0 {
			di := (*DRAWITEMSTRUCT)(unsafe.Pointer(lParam))
			if di.CtlType == ODT_BUTTON {
				drawOwnerButton(di)
				return 1
			}
		}
	case WM_CTLCOLOREDIT:
		hdc := wParam
		procSetTextColor.Call(hdc, colText)
		procSetBkColor.Call(hdc, colInput)
		return brushInput
	case WM_CTLCOLORLISTBOX:
		hdc := wParam
		procSetTextColor.Call(hdc, colText)
		procSetBkColor.Call(hdc, colPanel)
		return brushPanel
	case WM_CTLCOLORSTATIC:
		hdc := wParam
		ctl := lParam
		if ctl == hOutput {
			procSetBkMode.Call(hdc, OPAQUE)
			procSetTextColor.Call(hdc, colText)
			procSetBkColor.Call(hdc, colInput)
			return brushInput
		}
		procSetBkMode.Call(hdc, TRANSPARENT)
		if ctl == hTitle {
			procSetTextColor.Call(hdc, colText)
		} else {
			procSetTextColor.Call(hdc, colMuted)
		}
		if panelStatics[ctl] {
			return brushPanel
		}
		return brushBG
	case WM_CTLCOLORBTN:
		hdc := wParam
		procSetTextColor.Call(hdc, colText)
		procSetBkColor.Call(hdc, colPanel)
		return brushPanel
	case WM_APP_ANALYSIS_DONE:
		resultMu.Lock()
		res := pendingResult
		pendingResult = ""
		resultMu.Unlock()
		replaceOutput(res)
		setBusy(false)
		return 0
	case WM_COMMAND:
		id := int(wParam & 0xffff)
		code := int((wParam >> 16) & 0xffff)
		switch id {
		case ID_PASTE:
			if code != 0 {
				return 0
			} // BN_CLICKED
			procSetFocus.Call(hInput)
			procSendMessageW.Call(hInput, WM_PASTE, 0, 0)
		case ID_OPEN:
			if code != 0 {
				return 0
			} // BN_CLICKED
			safeUI("Apri testo", openTextFile)
		case ID_CLEAR:
			if code != 0 {
				return 0
			} // BN_CLICKED
			setText(hInput, "")
			replaceOutput("")
			setText(hParam, "")
			setText(hStatus, fmt.Sprintf("Locale · %d operazioni · menu destro attivo · nessun invio esterno", len(operations)))
			procSetFocus.Call(hInput)
		case ID_QUICK_AZ:
			if code != 0 {
				return 0
			} // BN_CLICKED
			procSendMessageW.Call(hCombo, CB_SETCURSEL, 20, 0)
			updateHint()
			safeUI("A–Z", func() { runOp(20) })
		case ID_QUICK_REP:
			if code != 0 {
				return 0
			} // BN_CLICKED
			procSendMessageW.Call(hCombo, CB_SETCURSEL, 9, 0)
			updateHint()
			safeUI("Ripetizioni", func() { runOp(9) })
		case ID_QUICK_KUNTA:
			if code != 0 {
				return 0
			} // BN_CLICKED
			procSendMessageW.Call(hCombo, CB_SETCURSEL, 26, 0)
			updateHint()
			safeUI("Kunta il testo", func() { runOp(26) })
		case ID_QUICK_VOW:
			if code != 0 {
				return 0
			} // BN_CLICKED
			procSendMessageW.Call(hCombo, CB_SETCURSEL, 27, 0)
			updateHint()
			safeUI("Isovocaliche", func() { runOp(27) })
		case ID_QUICK_CONS:
			if code != 0 {
				return 0
			} // BN_CLICKED
			procSendMessageW.Call(hCombo, CB_SETCURSEL, 28, 0)
			updateHint()
			safeUI("Isoconsonantiche", func() { runOp(28) })
		case ID_COMBO:
			if code == 1 {
				updateHint()
			}
		case ID_RUN:
			if code != 0 {
				return 0
			} // BN_CLICKED
			i, _, _ := procSendMessageW.Call(hCombo, CB_GETCURSEL, 0, 0)
			safeUI("KUNTA", func() { runOp(int(i)) })
		case ID_COPY:
			if code != 0 {
				return 0
			} // BN_CLICKED
			procSendMessageW.Call(hOutput, EM_SETSEL, 0, ^uintptr(0))
			procSendMessageW.Call(hOutput, WM_COPY, 0, 0)
			setText(hStatus, "Risultato copiato.")
		case ID_FOOTER:
			if code != 0 {
				return 0
			}
			procShellExecuteW.Call(hwndMain,
				uintptr(unsafe.Pointer(u16("open"))),
				uintptr(unsafe.Pointer(u16(repoURL))),
				0, 0, 1)
		}
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}

func main() {
	runtime.LockOSThread()

	if len(os.Args) > 1 && os.Args[1] == "--remove-context-menu" {
		unregisterShellIntegration()
		return
	}
	contextOK := registerShellIntegration()
	startupFile := ""
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--") {
			continue
		}
		if fi, err := os.Stat(arg); err == nil && !fi.IsDir() {
			startupFile = arg
			break
		}
	}
	hrCOM, _, _ := procCoInitializeEx.Call(0, 0x2) // COINIT_APARTMENTTHREADED
	if !hresultFailed(hrCOM) {
		defer procCoUninitialize.Call()
	}
	initBotolo()
	hInst, _, _ := procGetModuleHandleW.Call(0)
	cur, _, _ := procLoadCursorW.Call(0, 32512)
	appIcon, _, _ = procLoadIconW.Call(hInst, 1)

	brushBG, _, _ = procCreateSolidBrush.Call(colBG)
	brushPanel, _, _ = procCreateSolidBrush.Call(colPanel)
	brushInput, _, _ = procCreateSolidBrush.Call(colInput)
	brushAccent, _, _ = procCreateSolidBrush.Call(colAccent)
	brushSoft, _, _ = procCreateSolidBrush.Call(colSoft)
	brushLine, _, _ = procCreateSolidBrush.Call(colLine)

	class := u16("KuntaNativeWindowFinal")
	wc := WNDCLASSEX{CbSize: uint32(unsafe.Sizeof(WNDCLASSEX{})), LpfnWndProc: syscall.NewCallback(wndProc), HInstance: hInst, HIcon: appIcon, HCursor: cur, HbrBackground: brushBG, LpszClassName: class, HIconSm: appIcon}
	if r, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		return
	}

	hwndMain, _, _ = procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(u16("Kunta"))), WS_OVERLAPPEDWINDOW|WS_VISIBLE|WS_CLIPCHILDREN, CW_USEDEFAULT, CW_USEDEFAULT, 1120, 860, 0, 0, hInst, 0)
	if hwndMain == 0 {
		return
	}
	applyDPI(hwndMain)
	setImmersiveDarkTitlebar(hwndMain)
	setControlFonts()
	procDragAcceptFiles.Call(hwndMain, 1)

	hTitle = addStatic("Kunta", ID_TITLE, false)
	hTagline = addStatic("Conta · osserva · ordina · rac-conta", ID_TAGLINE, false)

	addButton("Incolla", ID_PASTE)
	addButton("Apri testo…", ID_OPEN)
	addButton("Pulisci", ID_CLEAR)

	hInput = create("EDIT", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_VSCROLL|ES_MULTILINE|ES_AUTOVSCROLL|ES_WANTRETURN|ES_NOHIDESEL, WS_EX_CLIENTEDGE, ID_EDIT_INPUT)
	procSendMessageW.Call(hInput, EM_SETMARGINS, EC_LEFTMARGIN|EC_RIGHTMARGIN, uintptr(scale(10)|(scale(10)<<16)))
	procSendMessageW.Call(hInput, EM_SETLIMITTEXT, 0x7FFFFFFE, 0)

	addButton("A–Z", ID_QUICK_AZ)
	addButton("Ripetizioni", ID_QUICK_REP)
	addButton("Kunta il testo", ID_QUICK_KUNTA)
	addButton("Isovocaliche", ID_QUICK_VOW)
	addButton("Isoconsonantiche", ID_QUICK_CONS)

	addStatic(fmt.Sprintf("Operazione · %d voci", len(operations)), ID_LABEL_OPERATION, true)
	addStatic("Parametro", ID_LABEL_PARAM, true)
	hCombo = create("COMBOBOX", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_VSCROLL|CBS_DROPDOWNLIST|CBS_HASSTRINGS|CBS_NOINTEGRALHEIGHT, 0, ID_COMBO)
	for _, op := range operations {
		procSendMessageW.Call(hCombo, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(u16(op.Name))))
	}
	procSendMessageW.Call(hCombo, CB_SETCURSEL, 20, 0)
	procSendMessageW.Call(hCombo, CB_SETDROPPEDWIDTH, uintptr(scale(520)), 0)
	procSendMessageW.Call(hCombo, CB_SETMINVISIBLE, 18, 0)
	if n, _, _ := procSendMessageW.Call(hCombo, CB_GETCOUNT, 0, 0); int(n) != len(operations) {
		message(fmt.Sprintf("Errore interno: operazioni caricate %d su %d.", int(n), len(operations)))
	}

	hParam = create("EDIT", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|ES_AUTOHSCROLL, WS_EX_CLIENTEDGE, ID_PARAM)
	procSendMessageW.Call(hParam, EM_SETMARGINS, EC_LEFTMARGIN|EC_RIGHTMARGIN, uintptr(scale(8)|(scale(8)<<16)))
	hCase = create("BUTTON", "Maiuscole/minuscole distinte", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX, 0, ID_CASE)
	addButton("KUNTA!", ID_RUN)
	hHint = addStatic("", ID_HINT, true)
	updateHint()

	hOutput = create("EDIT", "", WS_CHILD|WS_VISIBLE|WS_VSCROLL|ES_MULTILINE|ES_AUTOVSCROLL|ES_WANTRETURN|ES_READONLY|ES_NOHIDESEL, WS_EX_CLIENTEDGE, ID_OUTPUT)
	procSendMessageW.Call(hOutput, EM_SETMARGINS, EC_LEFTMARGIN|EC_RIGHTMARGIN, uintptr(scale(10)|(scale(10)<<16)))
	procSendMessageW.Call(hOutput, EM_SETLIMITTEXT, 0x7FFFFFFE, 0)
	addButton("Copia risultato", ID_COPY)
	statusText := fmt.Sprintf("Locale · %d operazioni · nessun invio esterno", len(operations))
	if contextOK {
		statusText = fmt.Sprintf("Locale · %d operazioni · menu destro attivo · nessun invio esterno", len(operations))
	}
	hStatus = addStatic(statusText, ID_STATUS, false)
	hFooter = addButton("ShiduLab", ID_FOOTER)

	layout()
	procInvalidateRect.Call(hwndMain, 0, 1)
	if startupFile != "" {
		safeUI("Apertura da Esplora file", func() { loadTextPath(startupFile) })
	} else {
		procSetFocus.Call(hInput)
	}

	var msg MSG
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}
