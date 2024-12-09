package controllers

import (
	"bytes"
	"fmt"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/chai2010/webp" // 引入webp解码库
	"github.com/nfnt/resize"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
)

// 主控制器
type ColorController struct {
	beego.Controller
}

// 获取图像的主色
func (c *ColorController) GetDominantColor(img image.Image) (int, int, int) {
	// 获取图片的宽高
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// 用来存储RGB色值的计数
	colorMap := make(map[color.RGBA]int)

	// 遍历图像中的每个像素点
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// 获取像素点的颜色
			at := img.At(x, y)
			// 转换为 RGBA
			r, g, b, _ := at.RGBA()
			colorMap[color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}]++
		}
	}

	// 找到出现最多的颜色
	var dominantColor color.RGBA
	maxCount := 0
	for col, count := range colorMap {
		if count > maxCount {
			maxCount = count
			dominantColor = col
		}
	}
	return int(dominantColor.R), int(dominantColor.G), int(dominantColor.B)
}

// Get 获取图片的主色
func (c *ColorController) Get() {
	// 获取URL参数
	urlStr := c.GetString("url")
	if urlStr == "" {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Ctx.WriteString("url parameter is missing")
		return
	}

	// 验证并解析URL
	_, err := url.ParseRequestURI(urlStr)
	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		c.Ctx.WriteString("invalid URL")
		return
	}

	// 下载图片
	resp, err := http.Get(urlStr)
	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		c.Ctx.WriteString("failed to fetch image: " + err.Error())
		return
	}
	defer resp.Body.Close()

	// 读取图片数据
	imgData, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		c.Ctx.WriteString("failed to read image data: " + err.Error())
		return
	}

	// 判断图像格式并解码
	var img image.Image
	// 尝试解码 PNG, JPEG, GIF 格式
	img, _, err = image.Decode(bytes.NewReader(imgData))
	if err != nil {
		// 如果不能解码常见格式，尝试解码 WebP 格式
		img, err = webp.Decode(bytes.NewReader(imgData))
		if err != nil {
			c.Ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
			c.Ctx.WriteString("failed to decode image: " + err.Error())
			return
		}
	}

	// 调整图片大小减少计算量（可选）
	img = resize.Resize(100, 0, img, resize.Lanczos3)

	// 获取主色
	r, g, b := c.GetDominantColor(img)

	// 返回结果
	hexColor := fmt.Sprintf("#%02X%02X%02X", r, g, b)
	response := map[string]string{"RGB": hexColor}

	// 输出 JSON 响应
	c.Data["json"] = response
	c.ServeJSON()
}
