//go:build windows

package screenshot

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"reflect"
	"syscall"
	"unsafe"
)

// copy github.com/vo va616/screenshot

func ScreenRect() (image.Rectangle, error) {
	hDC := GetDC(0)
	if hDC == 0 {
		return image.Rectangle{}, fmt.Errorf("Could not Get primary display err:%w\n", syscall.GetLastError())
	}
	defer ReleaseDC(0, hDC)
	// 注意多显示器,这里只返回主显示器宽高
	return image.Rect(0, 0,
		GetDeviceCaps(hDC, DESKTOPHORZRES),
		GetDeviceCaps(hDC, DESKTOPVERTRES),
	), nil
}

func CaptureScreen() (image.Image, error) {
	r, e := ScreenRect()
	if e != nil {
		return nil, e
	}
	if cnt := GetSystemMetrics(SM_CMONITORS); cnt > 1 {
		r.Max.X *= cnt // 按照显示器个数扩展宽度
	}
	return CaptureRect(r)
}

func SaveToFile(img image.Image, p string) error {
	fw, err := os.Create(p)
	if err != nil {
		return err
	}
	err = png.Encode(fw, img)
	_ = fw.Close()
	return err
}

func CaptureRect(rect image.Rectangle) (image.Image, error) {
	hDC := GetDC(0)
	if hDC == 0 {
		return nil, fmt.Errorf("Could not Get primary display err:%w.\n", syscall.GetLastError())
	}
	defer ReleaseDC(0, hDC)

	mHDC := CreateCompatibleDC(hDC)
	if mHDC == 0 {
		return nil, fmt.Errorf("Could not Create Compatible DC err:%w.\n", syscall.GetLastError())
	}
	defer DeleteDC(mHDC)

	x, y := rect.Dx(), rect.Dy()

	bt := BITMAPINFO{}
	bt.BmiHeader.BiSize = uint32(reflect.TypeOf(bt.BmiHeader).Size())
	bt.BmiHeader.BiWidth = int32(x)
	bt.BmiHeader.BiHeight = int32(-y)
	bt.BmiHeader.BiPlanes = 1
	bt.BmiHeader.BiBitCount = 32
	bt.BmiHeader.BiCompression = BI_RGB

	//goland:noinspection GoVetUnsafePointer
	ptr := unsafe.Pointer(uintptr(0))

	mHBmp := CreateDIBSection(mHDC, &bt, DIB_RGB_COLORS, &ptr, 0, 0)
	if mHBmp == 0 {
		return nil, fmt.Errorf("Could not Create DIB Section err:%w.\n", syscall.GetLastError())
	}
	if mHBmp == InvalidParameter {
		return nil, fmt.Errorf("One or more of the input parameters is invalid while calling CreateDIBSection.\n")
	}
	defer DeleteObject(mHBmp)

	obj := SelectObject(mHDC, mHBmp)
	if obj == 0 {
		return nil, fmt.Errorf("error occurred and the selected object is not a region err:%w.\n", syscall.GetLastError())
	}
	if obj == 0xffffffff { // GDI_ERROR
		return nil, fmt.Errorf("GDI_ERROR while calling SelectObject err:%w.\n", syscall.GetLastError())
	}
	defer DeleteObject(obj)

	if !BitBlt(mHDC, 0, 0, x, y, hDC, rect.Min.X, rect.Min.Y, SRCCOPY) {
		return nil, fmt.Errorf("BitBlt failed err:%w.\n", syscall.GetLastError())
	}

	var slice []byte
	hDrp := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	hDrp.Data = uintptr(ptr)
	hDrp.Len = x * y * 4
	hDrp.Cap = hDrp.Len

	pix := make([]byte, hDrp.Len)
	for i := 0; i < hDrp.Len; i += 4 { // R B G A
		pix[i], pix[i+2], pix[i+1], pix[i+3] = slice[i+2], slice[i], slice[i+1], slice[i+3]
	}

	return &image.RGBA{
		Pix:    pix,
		Stride: 4 * x,
		Rect:   image.Rect(0, 0, x, y),
	}, nil
}

func GetDeviceCaps(hdc HDC, index int) int {
	ret, _, _ := procGetDeviceCaps.Call(uintptr(hdc), uintptr(index))

	return int(ret)
}

func GetSystemMetrics(index int) int {
	ret, _, _ := procGetSystemMetrics.Call(uintptr(index))

	return int(ret)
}

func GetDC(h HWND) HDC {
	ret, _, _ := procGetDC.Call(uintptr(h))

	return HDC(ret)
}

func ReleaseDC(h HWND, hDC HDC) bool {
	ret, _, _ := procReleaseDC.Call(uintptr(h), uintptr(hDC))

	return ret != 0
}

func DeleteDC(hdc HDC) bool {
	ret, _, _ := procDeleteDC.Call(uintptr(hdc))

	return ret != 0
}

func BitBlt(hdcDest HDC, nXDest, nYDest, nWidth, nHeight int, hdcSrc HDC, nXSrc, nYSrc int, dwRop uint) bool {
	ret, _, _ := procBitBlt.Call(
		uintptr(hdcDest),
		uintptr(nXDest),
		uintptr(nYDest),
		uintptr(nWidth),
		uintptr(nHeight),
		uintptr(hdcSrc),
		uintptr(nXSrc),
		uintptr(nYSrc),
		uintptr(dwRop))

	return ret != 0
}

func SelectObject(hdc HDC, hGdiObj HGDIOBJ) HGDIOBJ {
	ret, _, _ := procSelectObject.Call(uintptr(hdc), uintptr(hGdiObj))

	if ret == 0 {
		panic("SelectObject failed")
	}

	return HGDIOBJ(ret)
}

func DeleteObject(hObject HGDIOBJ) bool {
	ret, _, _ := procDeleteObject.Call(uintptr(hObject))

	return ret != 0
}

func CreateDIBSection(hdc HDC, pBmi *BITMAPINFO, iUsage uint, ppvBits *unsafe.Pointer, hSection HANDLE, dwOffset uint) HGDIOBJ {
	ret, _, _ := procCreateDIBSection.Call(
		uintptr(hdc),
		uintptr(unsafe.Pointer(pBmi)),
		uintptr(iUsage),
		uintptr(unsafe.Pointer(ppvBits)),
		uintptr(hSection),
		uintptr(dwOffset))

	return HGDIOBJ(ret)
}

func CreateCompatibleDC(hdc HDC) HDC {
	ret, _, _ := procCreateCompatibleDC.Call(uintptr(hdc))

	if ret == 0 {
		panic("Create compatible DC failed")
	}

	return HDC(ret)
}

//goland:noinspection ALL
type (
	HANDLE  uintptr
	HWND    HANDLE
	HGDIOBJ HANDLE
	HDC     HANDLE

	BITMAPINFO struct {
		BmiHeader BITMAPINFOHEADER
		BmiColors *RGBQUAD
	}

	BITMAPINFOHEADER struct {
		BiSize          uint32
		BiWidth         int32
		BiHeight        int32
		BiPlanes        uint16
		BiBitCount      uint16
		BiCompression   uint32
		BiSizeImage     uint32
		BiXPelsPerMeter int32
		BiYPelsPerMeter int32
		BiClrUsed       uint32
		BiClrImportant  uint32
	}

	RGBQUAD struct {
		RgbBlue     byte
		RgbGreen    byte
		RgbRed      byte
		RgbReserved byte
	}
)

//goland:noinspection ALL
const (
	HORZRES          = 8   // 屏幕的宽度(物理)
	VERTRES          = 10  // 屏幕的高度(物理)
	DESKTOPHORZRES   = 118 // 屏幕的宽度(真实)
	DESKTOPVERTRES   = 117 // 屏幕的高度(真实)
	SM_CMONITORS     = 80  // 显示器个数
	BI_RGB           = 0
	InvalidParameter = 2
	DIB_RGB_COLORS   = 0
	SRCCOPY          = 0x00CC0020
)

//goland:noinspection ALL
var (
	modgdi32               = syscall.NewLazyDLL("gdi32.dll")
	moduser32              = syscall.NewLazyDLL("user32.dll")
	procGetDC              = moduser32.NewProc("GetDC")
	procReleaseDC          = moduser32.NewProc("ReleaseDC")
	procGetSystemMetrics   = moduser32.NewProc("GetSystemMetrics")
	procDeleteDC           = modgdi32.NewProc("DeleteDC")
	procBitBlt             = modgdi32.NewProc("BitBlt")
	procDeleteObject       = modgdi32.NewProc("DeleteObject")
	procSelectObject       = modgdi32.NewProc("SelectObject")
	procCreateDIBSection   = modgdi32.NewProc("CreateDIBSection")
	procCreateCompatibleDC = modgdi32.NewProc("CreateCompatibleDC")
	procGetDeviceCaps      = modgdi32.NewProc("GetDeviceCaps")
)
