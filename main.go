package main

import (
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"time"
)

const ApiToken string = "5631948806:AAF3ie1LWFnF_smLZKl4Rp1ZZyGEK06PGm8"

func SendText(bot *tgbotapi.BotAPI, updateMsg *tgbotapi.Message, text string) error {
	msg := tgbotapi.NewMessage(updateMsg.Chat.ID, text)
	_, err := bot.Send(msg)
	return err
}

func SendImage(bot *tgbotapi.BotAPI, updateMsg *tgbotapi.Message) error {
	if updateMsg.Photo == nil {
		return SendText(bot, updateMsg, "Прикрепите картинку!")
	}
	fileID := (*updateMsg.Photo)[2].FileID
	url, _ := bot.GetFileDirectURL(fileID)
	fileName := fmt.Sprintf("image_%d.png", time.Now().Unix())
	err := DownloadFile(fileName, url)
	if err != nil {
		return err
	}

	newFileName, _ := ModifyImage(fileName, updateMsg.Caption)

	msg := tgbotapi.NewPhotoUpload(updateMsg.Chat.ID, newFileName)
	_, err = bot.Send(msg)
	return err
}

func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func ModifyImage(filename string, imgText string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	img, err := jpeg.Decode(file)
	if err != nil {
		return "", err
	}

	rp := img.Bounds().Size()
	w, h := rp.X, rp.Y

	k1, k2, k3 := 8, 25, 30

	newImg := image.NewRGBA(image.Rect(-w/k1, -h/k1, w+w/k1, h+h/k1*3))
	whiteImg := image.NewRGBA(image.Rect(-w/k2, -h/k2, w+w/k2, h+h/k2))
	blackinImg := image.NewRGBA(image.Rect(-w/k3, -h/k3, w+w/k3, h+h/k3))

	draw.Draw(newImg, whiteImg.Bounds(), &image.Uniform{C: color.White}, image.Point{}, 0)
	draw.Draw(newImg, blackinImg.Bounds(), &image.Uniform{C: color.Black}, image.Point{}, 0)
	draw.Draw(newImg, img.Bounds(), img, image.Point{X: 0, Y: 0}, 0)
	drawText(newImg, imgText)
	f, _ := os.Create("new_" + filename)
	err = jpeg.Encode(f, newImg, nil)
	if err != nil {
		return "", err
	}
	return f.Name(), nil
}

func drawText(canvas *image.RGBA, text string) error {
	var (
		fgColor  image.Image
		fontFace *truetype.Font
		err      error
		fontSize = 60.0
	)
	fgColor = image.White
	fontFace, err = freetype.ParseFont(goregular.TTF)
	fontDrawer := font.Drawer{
		Dst: canvas,
		Src: fgColor,
		Face: truetype.NewFace(fontFace, &truetype.Options{
			Size:    fontSize,
			Hinting: font.HintingFull,
		}),
	}
	textBounds, _ := fontDrawer.BoundString(text)
	xPosition := (fixed.I(canvas.Rect.Max.X) - fontDrawer.MeasureString(text)) / 5
	textHeight := textBounds.Max.Y - textBounds.Min.Y
	yPosition := fixed.I((canvas.Rect.Max.Y)-textHeight.Ceil())/10*9 + fixed.I(textHeight.Ceil())
	fontDrawer.Dot = fixed.Point26_6{
		X: xPosition,
		Y: yPosition,
	}
	fontDrawer.DrawString(text)
	return err
}

func main() {
	bot, err := tgbotapi.NewBotAPI(ApiToken)
	if err != nil {
		panic(err)
	}
	bot.Debug = true

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates, err := bot.GetUpdatesChan(updateConfig)
	if err != nil {
		panic(err)
	}

	for update := range updates {
		var err error
		switch update.Message.Text {
		case "/start":
			err = SendText(bot, update.Message, "Ну давай начнем!")
		default:
			err = SendImage(bot, update.Message)
		}
		if err != nil {
			_ = SendText(bot, update.Message, "Вы все сломали...")
			panic(err)
		}
		fmt.Println("____________________________________________________________________")
	}
}
