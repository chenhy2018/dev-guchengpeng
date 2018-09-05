package image

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"mime"
	"net/http"
	"os"
	"qbox.us/api/rs"
	"qbox.us/api/uc"
	"qbox.us/oauth"
	"strings"
	"sync"
	"testing"
	"time"
)

var once sync.Once

type Case struct {
	img, style                    string
	expectedWidth, expectedHeight int
}

func getExt(name string) string {
	x := strings.LastIndex(name, ".")
	if x != -1 {
		return strings.ToLower(name[x:])
	}
	return ""
}

func getService() (*rs.Service, *uc.Service) {
	username := "test@qbox.net"
	password := "test"

	rsHost := "http://rs.qbox.me:10100"
	accHost := "https://acc.qbox.me"
	ucHost := "http://uc.qbox.me"

	var CLIENT_ID = "abcd0c7edcdf914228ed8aa7c6cee2f2bc6155e2"
	var CLIENT_SECRET = "fc9ef8b171a74e197b17f85ba23799860ddf3b9c"
	var REDIRECT_URI = "<RedirectURL>"
	var AUTHORIZATION_ENDPOINT = "<AuthURL>"

	TOKEN_ENDPOINT := accHost + "/oauth2/token"

	var config = &oauth.Config{
		ClientId:     CLIENT_ID,
		ClientSecret: CLIENT_SECRET,
		Scope:        "Scope",
		AuthURL:      AUTHORIZATION_ENDPOINT,
		TokenURL:     TOKEN_ENDPOINT,
		RedirectURL:  REDIRECT_URI,
	}

	transport := &oauth.Transport{
		Config:    config,
		Transport: http.DefaultTransport,
	}

	// User Login
	_, code, err := transport.ExchangeByPassword(username, password)
	if code != 200 {
		fmt.Println("Login by password failed:", code, "-", err)
		os.Exit(-1)
	}

	rs_service := rs.New(rsHost, transport)
	uc_service := uc.New(ucHost, transport)
	return rs_service, uc_service
}

func clearTable(rs_service *rs.Service, table_name string) {
	code, err := rs_service.Drop(table_name)
	if err != nil {
		fmt.Println("Drop failed:", code, "-", err)
		os.Exit(-1)
	}
	fmt.Println("Table was cleared!")
}

func publishFiles(rs_service *rs.Service, table_name string) {
	imgs := map[string]string{"email": "img/email.png", "finger": "img/finger.jpg", "girl": "img/girl.jpeg", "iphone": "img/iphone.jpg", "sketch": "img/sketch.jpg"}

	ioHost := "http://io.qbox.me"

	for img_key, img_name := range imgs {
		// Get file information
		mimeType := mime.TypeByExtension(getExt(img_name))
		fmt.Println("mimeType:", mimeType)
		entryURI := table_name + ":" + img_key

		file, err := os.Open(img_name)
		if err != nil {
			fmt.Println("Open file failed:", err)
			os.Exit(-1)
		}
		attr, _ := os.Stat(img_name)
		fileSize := attr.Size()

		// Put file to server
		_, code, err := rs_service.Put(ioHost, entryURI, mimeType, file, fileSize)
		if err != nil {
			fmt.Println("Put failed:", code, "-", err)
			os.Exit(-1)
		}
		fmt.Println("File was put to server.")

		// Publish file to public
		domain := "io.qbox.me/ikbear2"
		code, err = rs_service.Publish(domain, table_name)
		if err != nil {
			fmt.Println("Publish file failed:", code, "-", err)
			os.Exit(-1)
		}
		fmt.Println("File was publish to public.")
	}
}

func setStyle(uc_service *uc.Service) {
	style := map[string]string{"small": "square:128;q:85", "medium": "400x300;q:85", "xlarge": "800x;q:85", "ylarge": "x600;q:85"}
	for styleName, styleType := range style {
		uc_service.SetImagePreviewStyle(styleName, styleType)
	}
	time.Sleep(6 * time.Minute)
	fmt.Println("Set customed style successfully!")
}

func publishAndSetStyle() {
	rs_service, uc_service := getService()
	table_name := "ikbear2"
	clearTable(rs_service, table_name)
	publishFiles(rs_service, table_name)
	setStyle(uc_service)
}

func getImageWidthHeight(img_key, image_type string) (int, int) {
	preview_url := "http://io.qbox.me/ikbear2/" + img_key + "?imagePreviewEx/" + image_type
	preview_image, err := http.Get(preview_url)
	if err != nil {
		fmt.Println("Read preview image failed:", err)
		os.Exit(-1)
	}
	defer preview_image.Body.Close()
	config, _, err := image.DecodeConfig(preview_image.Body)
	if err != nil {
		fmt.Println("Decode preview image failed:", err)
		os.Exit(-1)
	}
	return config.Width, config.Height
}

func testImageFramework(t *testing.T, cases []Case) {
	once.Do(publishAndSetStyle)
	for _, imgs := range cases {
		width, height := getImageWidthHeight(imgs.img, imgs.style)
		if imgs.expectedWidth != -1 {
			if width != imgs.expectedWidth {
				t.Log("Image:", imgs.img, "Unexpected Width:", width)
				t.Fail()
			}
		}
		if imgs.expectedHeight != -1 {
			if height != imgs.expectedHeight {
				t.Log("Image:", imgs.img, "Unexpected Height:", height)
				t.Fail()
			}
		}
	}
}

func TestCase_0(t *testing.T) {
	cases := []Case{
		{"email", "0", 160, 21},
		{"finger", "0", 498, 330},
		{"girl", "0", 399, 599},
		{"iphone", "0", -1, 600},
		{"sketch", "0", -1, 600},
	}
	testImageFramework(t, cases)
}

func TestCase_1(t *testing.T) {
	cases := []Case{
		{"email", "1", 128, -1},
		{"finger", "1", 128, -1},
		{"girl", "1", 128, -1},
		{"iphone", "1", 128, -1},
		{"sketch", "1", 128, -1},

		{"finger", "1", -1, 128},
		{"girl", "1", -1, 128},
		{"iphone", "1", -1, 128},
		{"sketch", "1", -1, 128},
	}
	testImageFramework(t, cases)
}

func TestCase_10(t *testing.T) {
	cases := []Case{
		{"email", "10", 150, -1},
		{"finger", "10", 150, 150},
		{"girl", "10", 150, 150},
		{"iphone", "10", 150, 150},
		{"sketch", "10", 150, 150},
	}
	testImageFramework(t, cases)
}

func TestCase_21(t *testing.T) {
	cases := []Case{
		{"email", "21", 160, 21},
		{"finger", "21", 498, 330},
		{"girl", "21", 399, 599},
		{"iphone", "21", 640, 640},
		{"sketch", "21", -1, 640},
	}
	testImageFramework(t, cases)
}

func TestCase_22(t *testing.T) {
	cases := []Case{
		{"email", "22", 160, 21},
		{"finger", "22", 498, 330},
		{"girl", "22", 399, 599},
		{"iphone", "22", 640, -1},
		{"sketch", "22", 600, 800},
	}
	testImageFramework(t, cases)
}

func TestCase_23(t *testing.T) {
	cases := []Case{
		{"email", "23", 160, 21},
		{"finger", "23", 320, -1},
		{"girl", "23", 320, -1},
		{"iphone", "23", 320, -1},
		{"sketch", "23", 320, -1},
	}
	testImageFramework(t, cases)
}

func TestCase_24(t *testing.T) {
	cases := []Case{
		{"email", "24", 160, 21},
		{"finger", "24", 498, 330},
		{"girl", "24", 399, 599},
		{"iphone", "24", -1, 640},
		{"sketch", "24", -1, 640},
	}
	testImageFramework(t, cases)
}

func TestCase_25(t *testing.T) {
	cases := []Case{
		{"email", "25", 150, -1},
		{"finger", "25", 150, 150},
		{"girl", "25", 150, 150},
		{"iphone", "25", 150, 150},
		{"sketch", "25", 150, 150},
	}
	testImageFramework(t, cases)
}

func TestCase_50(t *testing.T) {
	cases := []Case{
		{"email", "50", 60, -1},
		{"finger", "50", 60, 40},
		{"girl", "50", 60, 40},
		{"iphone", "50", 60, 40},
		{"sketch", "50", 60, 40},
	}
	testImageFramework(t, cases)
}

func TestCase_Small(t *testing.T) {
	cases := []Case{
		{"email", "small", 160, 21},
		{"finger", "small", 128, 128},
		{"girl", "small", 128, 128},
		{"iphone", "small", 128, 128},
		{"sketch", "small", 128, 128},
	}
	testImageFramework(t, cases)
}

func TestCase_Medium(t *testing.T) {
	cases := []Case{
		{"email", "medium", 160, 21},
		{"finger", "medium", 400, 300},
		{"girl", "medium", 399, 599},
		{"iphone", "medium", 400, 300},
		{"sketch", "medium", 400, 300},
	}
	testImageFramework(t, cases)
}

func TestCase_Xlarge(t *testing.T) {
	cases := []Case{
		{"email", "xlarge", 160, 21},
		{"finger", "xlarge", 498, 330},
		{"girl", "xlarge", 399, 599},
		{"iphone", "xlarge", 640, 960},
		{"sketch", "xlarge", 600, 800},
	}
	testImageFramework(t, cases)
}

func TestCase_Ylarge(t *testing.T) {
	cases := []Case{
		{"email", "ylarge", 160, 21},
		{"finger", "ylarge", 498, 330},
		{"girl", "ylarge", 399, 599},
		{"iphone", "ylarge", -1, 600},
		{"sketch", "ylarge", -1, 600},
	}
	testImageFramework(t, cases)
}
