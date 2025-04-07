package global

import (
	"fmt"
	"github.com/vova616/screenshot"
	"image/png"
	"os"
	"time"
)

func Screenshot() {
	img, err := screenshot.CaptureScreen()
	if err != nil {
		fmt.Println("Ошибка при делании скриншота")
		return
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("screen_%s.png", timestamp)

	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("Ошибка при сохранении скрина")
		return
	}
	err = png.Encode(f, img)
	if err != nil {
		fmt.Println(err)
	}
	f.Close()
}
