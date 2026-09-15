//go:build windows

package main

import (
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
    "syscall"
    "unicode/utf16"
    "unsafe"

    "shidulab/kunta/analysis"
)

var (
    user32 = syscall.NewLazyDLL("user32.dll")
    kernel32 = syscall.NewLazyDLL("kernel32.dll")
    shell32 = syscall.NewLazyDLL("shell32.dll")
    comdlg32 = syscall.NewLazyDLL("comdlg32.dll")
    pCreateWindowExW = user32.NewProc("CreateWindowExW")
    pDefWindowProcW = user32.NewProc("DefWindowProcW")
    pDispatchMessageW = user32.NewProc("DispatchMessageW")
    pGetMessageW = user32.NewProc("GetMessageW")
    pPostQuitMessage = user32.NewProc("PostQuitMessage")
    pRegisterClassExW = user32.NewProc("RegisterClassExW")
    pTranslateMessage = user32.NewProc("TranslateMessage")
    pShowWindow = user32.NewProc("ShowWindow")
    pUpdateWindow = user32.NewProc("UpdateWindow")
    pSendMessageW = user32.NewProc("SendMessageW")
    pGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
    pGetWindowTextW = user32.NewProc("GetWindowTextW")
    pSetWindowTextW = user32.NewProc("SetWindowTextW")
    pMessageBoxW = user32.NewProc("MessageBoxW")
    pLoadImageW = user32.NewProc("LoadImageW")
    pGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
    pDragAcceptFiles = shell32.NewProc("DragAcceptFiles")
    pDragQueryFileW = shell32.NewProc("DragQueryFileW")
    pDragFinish = shell32.NewProc("DragFinish")
    pGetOpenFileNameW = comdlg32.NewProc("GetOpenFileNameW")
)

const (
    WS_OVERLAPPEDWINDOW=0x00CF0000; WS_VISIBLE=0x10000000; WS_CHILD=0x40000000
    WS_TABSTOP=0x00010000; WS_VSCROLL=0x00200000; WS_HSCROLL=0x00100000
    ES_MULTILINE=0x0004; ES_AUTOVSCROLL=0x0040; ES_AUTOHSCROLL=0x0080; ES_READONLY=0x0800; ES_WANTRETURN=0x1000
    CBS_DROPDOWNLIST=0x0003; BS_PUSHBUTTON=0; BS_AUTOCHECKBOX=3
    WM_DESTROY=0x0002; WM_COMMAND=0x0111; WM_DROPFILES=0x0233; WM_SETICON=0x0080
    BM_GETCHECK=0x00F0; BST_CHECKED=1
    CB_ADDSTRING=0x0143; CB_SETCURSEL=0x014E; CB_GETCURSEL=0x0147
    IMAGE_ICON=1; LR_LOADFROMFILE=0x0010; ICON_SMALL=0; ICON_BIG=1
    SW_SHOW=5; CW_USEDEFAULT=0x80000000
    ID_INPUT=101; ID_OPERATION=102; ID_PARAM=103; ID_CASE=104; ID_RUN=105; ID_OUTPUT=106; ID_OPEN=107; ID_PASTE=108
)

type POINT struct{X,Y int32}
type MSG struct{Hwnd uintptr; Message uint32; WParam,LParam uintptr; Time uint32; Pt POINT; LPrivate uint32}
type WNDCLASSEX struct{CbSize uint32; Style uint32; LpfnWndProc uintptr; CbClsExtra,CbWndExtra int32; HInstance,HIcon,HCursor,HbrBackground uintptr; LpszMenuName,LpszClassName *uint16; HIconSm uintptr}
type OpenFileName struct {
    LStructSize uint32; HwndOwner uintptr; HInstance uintptr; LpstrFilter *uint16; LpstrCustomFilter *uint16; NMaxCustFilter uint32; NFilterIndex uint32; LpstrFile *uint16; NMaxFile uint32; LpstrFileTitle *uint16; NMaxFileTitle uint32; LpstrInitialDir *uint16; LpstrTitle *uint16; Flags uint32; NFileOffset uint16; NFileExtension uint16; LpstrDefExt *uint16; LCustData uintptr; LpfnHook uintptr; LpTemplateName *uint16; PvReserved uintptr; DwReserved uint32; FlagsEx uint32
}

var hwndMain, hInput,hOperation,hParam,hCase,hOutput uintptr
var operations=[]string{
    "Kunta il testo — riepilogo",
    "Lettere indicate",
    "Occorrenze parola/stringa",
    "Frequenza lettere",
    "Frequenza parole",
    "Parole più corte / più lunghe",
    "Palindromi",
    "Bifronti presenti nel testo",
    "Gruppi di anagrammi",
    "Giochi/Peculiarità — Acrostico",
    "Giochi/Peculiarità — Telestico",
    "Isovocaliche",
    "Isoconsonantiche",
    "Omovocaliche",
    "Omovocaliche iniziali",
    "Omovocaliche finali",
    "Omoconsonantiche",
    "Omoconsonantiche iniziali",
    "Omoconsonantiche finali",
    "Inverti testo",
    "Inverti ordine parole",
    "Ordina parole alfabeticamente",
    "Estrai solo numeri",
}

func ptr(s string)*uint16{return syscall.StringToUTF16Ptr(s)}
func loword(v uintptr)uint16{return uint16(v&0xffff)}
func getText(h uintptr)string{n,_,_:=pGetWindowTextLengthW.Call(h);b:=make([]uint16,n+1);pGetWindowTextW.Call(h,uintptr(unsafe.Pointer(&b[0])),n+1);return syscall.UTF16ToString(b)}
func setText(h uintptr,s string){pSetWindowTextW.Call(h,uintptr(unsafe.Pointer(ptr(s))))}
func msgbox(s string){pMessageBoxW.Call(hwndMain,uintptr(unsafe.Pointer(ptr(s))),uintptr(unsafe.Pointer(ptr("Kunta"))),0x40)}
func create(class,text string,style uint32,x,y,w,h int32,id int)uintptr{r,_,_:=pCreateWindowExW.Call(0,uintptr(unsafe.Pointer(ptr(class))),uintptr(unsafe.Pointer(ptr(text))),uintptr(style),uintptr(x),uintptr(y),uintptr(w),uintptr(h),hwndMain,uintptr(id),0,0);return r}

func fmtGroups(title string, gs []analysis.Group)string{var b strings.Builder;b.WriteString(title+"\r\n\r\n");if len(gs)==0{b.WriteString("Nessun gruppo trovato.");return b.String()};for _,g:=range gs{if g.Key!=""{b.WriteString("["+g.Key+"] ")};b.WriteString(strings.Join(g.Words," · "));b.WriteString("\r\n")};return b.String()}
func intParam(def int)int{p:=strings.TrimSpace(getText(hParam));n,e:=strconv.Atoi(p);if e!=nil||n<1{return def};return n}
func runAnalysis(){
    text:=getText(hInput); if text==""{setText(hOutput,"Inserisci o incolla un testo.");return}
    idx,_,_:=pSendMessageW.Call(hOperation,CB_GETCURSEL,0,0); if int(idx)<0||int(idx)>=len(operations){idx=0}
    cs,_,_:=pSendMessageW.Call(hCase,BM_GETCHECK,0,0); caseSensitive:=cs==BST_CHECKED
    min:=intParam(3); edge:=intParam(2); var out string
    switch int(idx){
    case 0:s:=analysis.SummaryOf(text);out=fmt.Sprintf("KUNTA IL TESTO\r\n\r\nCaratteri: %d\r\nLettere: %d\r\nParole: %d\r\nRighe: %d\r\nSpazi/bianchi: %d\r\nCifre: %d\r\nPunteggiatura: %d\r\nParole uniche: %d\r\nParole ripetute: %d",s.Characters,s.Letters,s.Words,s.Lines,s.Spaces,s.Digits,s.Punctuation,s.UniqueWords,s.RepeatedWords)
    case 1:m:=analysis.CountLetters(text,getText(hParam),caseSensitive);ks:=make([]rune,0,len(m));for k:=range m{ks=append(ks,k)};sort.Slice(ks,func(i,j int)bool{return ks[i]<ks[j]});var b strings.Builder;for _,k:=range ks{fmt.Fprintf(&b,"%c: %d\r\n",k,m[k])};out=b.String()
    case 2:out=fmt.Sprintf("Occorrenze: %d",analysis.Occurrences(text,getText(hParam),caseSensitive))
    case 3:out=fmtGroups("FREQUENZA LETTERE",analysis.LetterFrequency(text,caseSensitive))
    case 4:out=fmtGroups("FREQUENZA PAROLE",analysis.WordFrequency(text,caseSensitive))
    case 5:a,b:=analysis.ShortestLongest(text);out="PIÙ CORTE\r\n"+strings.Join(a," · ")+"\r\n\r\nPIÙ LUNGHE\r\n"+strings.Join(b," · ")
    case 6:out="PALINDROMI\r\n\r\n"+strings.Join(analysis.Palindromes(text,min),"\r\n")
    case 7:out=fmtGroups("BIFRONTI",analysis.Bifronts(text,min))
    case 8:out=fmtGroups("ANAGRAMMI",analysis.Anagrams(text,min))
    case 9:out="ACROSTICO\r\n\r\n"+analysis.Acrostic(text)
    case 10:out="TELESTICO\r\n\r\n"+analysis.Telestic(text)
    case 11:out=fmtGroups("ISOVOCALICHE — stessa sequenza vocalica",analysis.Isovocalic(text,min))
    case 12:out=fmtGroups("ISOCONSONANTICHE — stessa sequenza consonantica",analysis.Isoconsonantic(text,min))
    case 13:out=fmtGroups("OMOVOCALICHE — stesso patrimonio vocalico",analysis.Homovocalic(text,min))
    case 14:out=fmtGroups(fmt.Sprintf("OMOVOCALICHE INIZIALI — prime %d vocali",edge),analysis.HomovocalicInitial(text,2,edge))
    case 15:out=fmtGroups(fmt.Sprintf("OMOVOCALICHE FINALI — ultime %d vocali",edge),analysis.HomovocalicFinal(text,2,edge))
    case 16:out=fmtGroups("OMOCONSONANTICHE — stesso patrimonio consonantico",analysis.Homoconsonantic(text,min))
    case 17:out=fmtGroups(fmt.Sprintf("OMOCONSONANTICHE INIZIALI — prime %d consonanti",edge),analysis.HomoconsonanticInitial(text,2,edge))
    case 18:out=fmtGroups(fmt.Sprintf("OMOCONSONANTICHE FINALI — ultime %d consonanti",edge),analysis.HomoconsonanticFinal(text,2,edge))
    case 19:out=analysis.ReverseText(text)
    case 20:out=analysis.ReverseWords(text)
    case 21:out=analysis.SortWords(text)
    case 22:out=strings.Join(analysis.Numbers(text),"\r\n")
    }
    setText(hOutput,out)
}

func openText(){buf:=make([]uint16,4096);filter:=utf16.Encode([]rune("Testi\x00*.txt;*.md;*.log;*.csv;*.tsv\x00Tutti i file\x00*.*\x00\x00"));ofn:=OpenFileName{LStructSize:uint32(unsafe.Sizeof(OpenFileName{})),HwndOwner:hwndMain,LpstrFilter:&filter[0],LpstrFile:&buf[0],NMaxFile:uint32(len(buf)),Flags:0x00001000|0x00000800};r,_,_:=pGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn)));if r==0{return};p:=syscall.UTF16ToString(buf);b,e:=os.ReadFile(p);if e!=nil{msgbox(e.Error());return};setText(hInput,string(b))}
func dropped(wp uintptr){n,_,_:=pDragQueryFileW.Call(wp,0xffffffff,0,0);if n>0{l,_,_:=pDragQueryFileW.Call(wp,0,0,0);b:=make([]uint16,l+1);pDragQueryFileW.Call(wp,0,uintptr(unsafe.Pointer(&b[0])),l+1);data,e:=os.ReadFile(syscall.UTF16ToString(b));if e==nil{setText(hInput,string(data))}};pDragFinish.Call(wp)}
func paste(){// WM_PASTE to input edit
    pSendMessageW.Call(hInput,0x0302,0,0)
}

func wndProc(hwnd uintptr,msg uint32,wp,lp uintptr)uintptr{
    switch msg{
    case WM_DESTROY:pPostQuitMessage.Call(0);return 0
    case WM_DROPFILES:dropped(wp);return 0
    case WM_COMMAND:switch int(loword(wp)){case ID_RUN:runAnalysis();case ID_OPEN:openText();case ID_PASTE:paste()};return 0
    }
    r,_,_:=pDefWindowProcW.Call(hwnd,uintptr(msg),wp,lp);return r
}

func main(){
    inst,_,_:=pGetModuleHandleW.Call(0); cls:=ptr("KuntaWindow"); wc:=WNDCLASSEX{CbSize:uint32(unsafe.Sizeof(WNDCLASSEX{})),LpfnWndProc:syscall.NewCallback(wndProc),HInstance:inst,HbrBackground:6,LpszClassName:cls}; pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
    r,_,_:=pCreateWindowExW.Call(0,uintptr(unsafe.Pointer(cls)),uintptr(unsafe.Pointer(ptr("Kunta — ShiduLab"))),WS_OVERLAPPEDWINDOW|WS_VISIBLE,CW_USEDEFAULT,CW_USEDEFAULT,1100,760,0,0,inst,0);hwndMain=r
    create("STATIC","TESTO",0x50000000,20,15,100,20,0)
    hInput=create("EDIT","",WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_VSCROLL|ES_MULTILINE|ES_AUTOVSCROLL|ES_WANTRETURN,20,40,1040,260,ID_INPUT)
    create("STATIC","Operazione",0x50000000,20,315,90,20,0)
    hOperation=create("COMBOBOX","",WS_CHILD|WS_VISIBLE|WS_TABSTOP|CBS_DROPDOWNLIST,110,310,465,400,ID_OPERATION)
    for _,op:=range operations{pSendMessageW.Call(hOperation,CB_ADDSTRING,0,uintptr(unsafe.Pointer(ptr(op))))};pSendMessageW.Call(hOperation,CB_SETCURSEL,0,0)
    create("STATIC","Parametro",0x50000000,590,315,80,20,0);hParam=create("EDIT","3",WS_CHILD|WS_VISIBLE|WS_TABSTOP|0x00800000,670,310,90,25,ID_PARAM)
    hCase=create("BUTTON","Maiuscole/minuscole distinte",WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX,775,307,235,30,ID_CASE)
    create("BUTTON","APRI TESTO...",WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON,20,350,130,32,ID_OPEN)
    create("BUTTON","INCOLLA",WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON,160,350,110,32,ID_PASTE)
    create("BUTTON","KUNTA",WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON,930,345,130,38,ID_RUN)
    create("STATIC","RISULTATO",0x50000000,20,400,100,20,0)
    hOutput=create("EDIT","",WS_CHILD|WS_VISIBLE|WS_VSCROLL|WS_HSCROLL|ES_MULTILINE|ES_AUTOVSCROLL|ES_AUTOHSCROLL|ES_READONLY,20,425,1040,270,ID_OUTPUT)
    pDragAcceptFiles.Call(hwndMain,1)
    exe,_:=os.Executable();ico:=filepath.Join(filepath.Dir(exe),"assets","Kunta.ico");hi,_,_:=pLoadImageW.Call(0,uintptr(unsafe.Pointer(ptr(ico))),IMAGE_ICON,0,0,LR_LOADFROMFILE);if hi!=0{pSendMessageW.Call(hwndMain,WM_SETICON,ICON_BIG,hi);pSendMessageW.Call(hwndMain,WM_SETICON,ICON_SMALL,hi)}
    pShowWindow.Call(hwndMain,SW_SHOW);pUpdateWindow.Call(hwndMain)
    var m MSG;for{q,_,_:=pGetMessageW.Call(uintptr(unsafe.Pointer(&m)),0,0,0);if int32(q)<=0{break};pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)));pDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))}
}
