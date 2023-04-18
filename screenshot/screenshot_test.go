package screenshot

import (
	"testing"
)

func TestScreenshot(t *testing.T) {
	img, err := CaptureScreen()
	if err != nil {
		t.Fatal(err)
	}

	// 全屏截图,保存为png文件
	err = SaveToFile(img, "screenshot.tmp.png")
	if err != nil {
		t.Fatal(err)
	}
}
