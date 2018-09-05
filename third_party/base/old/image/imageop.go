package main

import (
	"flag"
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
	"strconv"
	"strings"
	"sync"
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

func getService(args map[string]string) (*rs.Service, *uc.Service) {
	username := args["username"]
	password := args["password"]

	rsHost := args["rsHost"]
	accHost := args["accHost"]
	ucHost := args["ucHost"]

	CLIENT_ID := args["clientID"]
	CLIENT_SECRET := args["clientSecret"]
	REDIRECT_URI := args["redirectURI"]
	AUTHORIZATION_ENDPOINT := args["authorizationEndpoint"]

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

func clearTable(args map[string]string, rs_service *rs.Service) {
	code, err := rs_service.Drop(args["tableName"])
	if err != nil {
		fmt.Println("Drop failed:", code, "-", err)
		os.Exit(-1)
	}
	fmt.Println("Table was cleared!")
}

func publishFiles(args map[string]string, rs_service *rs.Service) {
	imgs := map[string]string{"email": "img/email.png", "finger": "img/finger.jpg", "girl": "img/girl.jpeg", "iphone": "img/iphone.jpg", "sketch": "img/sketch.jpg"}

	ioHost := args["ioHost"]
	tableName := args["tableName"]
	domain := args["publishURI"]

	for img_key, img_name := range imgs {
		// Get file information
		mimeType := mime.TypeByExtension(getExt(img_name))
		fmt.Println("mimeType:", mimeType)
		entryURI := tableName + ":" + img_key

		file, err := os.Open(img_name)
		defer file.Close()
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
		code, err = rs_service.Publish(domain, tableName)
		if err != nil {
			fmt.Println("Publish file failed:", code, "-", err)
			os.Exit(-1)
		}
		fmt.Println("File was publish to public.")
	}
}

func setStyle(args map[string]string, uc_service *uc.Service) {
	style := map[string]string{"small": "square:128;q:85", "medium": "400x300;q:85", "xlarge": "800x;q:85", "ylarge": "x600;q:85"}
	for styleName, styleType := range style {
		uc_service.SetImagePreviewStyle(styleName, styleType)
	}
	sleepTime, _ := strconv.ParseInt(args["sleepTime"], 10, 64)
	duration := sleepTime * int64(time.Minute)
	time.Sleep(time.Duration(duration))
	fmt.Println("Set customed style successfully!")
}

func Prepare(args map[string]string) {
	rs_service, _ := getService(args)
	clearTable(args, rs_service)
	publishFiles(args, rs_service)
	//    setStyle(args, uc_service)
}

func GetImageWidthHeight(args map[string]string, img_key, img_type string) (int, int) {
	preview_url := args["publishURI"] + "/" + args["tableName"] + "/" + img_key + "?imagePreviewEx/" + img_type
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

func main() {
	username := flag.String("u", "test@qbox.net", "Login User(username/email)")
	password := flag.String("p", "test", "Password")

	rsHost := flag.String("rs", "http://rs.qbox.me:10100", "RS host address")
	accHost := flag.String("acc", "https://acc.qbox.me", "Accout host address")
	ucHost := flag.String("uc", "http://uc.qbox.me", "UC host address")
	ioHost := flag.String("io", "http://io.qbox.me", "IO host address")

	clientID := flag.String("ci", "abcd0c7edcdf914228ed8aa7c6cee2f2bc6155e2", "Client ID")
	clientSecret := flag.String("ss", "fc9ef8b171a74e197b17f85ba23799860ddf3b9c", "Client secret key")

	redirectURI := flag.String("re", "<RedirectURL>", "Redirect URL")
	authorizationEndpoint := flag.String("au", "<AuthURL>", "Authorization endpoint")
	publishURI := flag.String("pub", "http://io.qbox.me", "Publish domain name")

	tableName := flag.String("tb", "ikbear2", "User defined table name")

	sleepTime := flag.String("st", "6", "Sleep time in Minutes")

	flag.Parse()

	var args = map[string]string{
		"username":              *username,
		"password":              *password,
		"rsHost":                *rsHost,
		"accHost":               *accHost,
		"ucHost":                *ucHost,
		"ioHost":                *ioHost,
		"clientID":              *clientID,
		"clientSecret":          *clientSecret,
		"redirectURI":           *redirectURI,
		"authorizationEndpoint": *authorizationEndpoint,
		"publishURI":            *publishURI,
		"tableName":             *tableName,
		"sleepTime":             *sleepTime,
	}

	cases := []Case{
		{"email", "0", 160, 21},
		{"finger", "0", 498, 330},
		{"girl", "0", 399, 599},
		{"iphone", "0", -1, 600},
		{"sketch", "0", -1, 600},

		{"email", "1", 128, -1},
		{"finger", "1", 128, -1},
		{"girl", "1", 128, -1},
		{"iphone", "1", 128, -1},
		{"sketch", "1", 128, -1},

		{"finger", "1", -1, 128},
		{"girl", "1", -1, 128},
		{"iphone", "1", -1, 128},
		{"sketch", "1", -1, 128},

		{"email", "10", 150, -1},
		{"finger", "10", 150, 150},
		{"girl", "10", 150, 150},
		{"iphone", "10", 150, 150},
		{"sketch", "10", 150, 150},

		{"email", "21", 160, 21},
		{"finger", "21", 498, 330},
		{"girl", "21", 399, 599},
		{"iphone", "21", 640, 640},
		{"sketch", "21", -1, 640},

		{"email", "22", 160, 21},
		{"finger", "22", 498, 330},
		{"girl", "22", 399, 599},
		{"iphone", "22", 640, -1},
		{"sketch", "22", 600, 800},

		{"email", "23", 160, 21},
		{"finger", "23", 320, -1},
		{"girl", "23", 320, -1},
		{"iphone", "23", 320, -1},
		{"sketch", "23", 320, -1},

		{"email", "24", 160, 21},
		{"finger", "24", 498, 330},
		{"girl", "24", 399, 599},
		{"iphone", "24", -1, 640},
		{"sketch", "24", -1, 640},

		{"email", "25", 150, -1},
		{"finger", "25", 150, 150},
		{"girl", "25", 150, 150},
		{"iphone", "25", 150, 150},
		{"sketch", "25", 150, 150},

		{"email", "50", 60, -1},
		{"finger", "50", 60, 40},
		{"girl", "50", 60, 40},
		{"iphone", "50", 60, 40},
		{"sketch", "50", 60, 40},

		{"email", "small", 160, 21},
		{"finger", "small", 128, 128},
		{"girl", "small", 128, 128},
		{"iphone", "small", 128, 128},
		{"sketch", "small", 128, 128},

		{"email", "medium", 160, 21},
		{"finger", "medium", 400, 300},
		{"girl", "medium", 399, 599},
		{"iphone", "medium", 400, 300},
		{"sketch", "medium", 400, 300},

		{"email", "xlarge", 160, 21},
		{"finger", "xlarge", 498, 330},
		{"girl", "xlarge", 399, 599},
		{"iphone", "xlarge", 640, 960},
		{"sketch", "xlarge", 600, 800},

		{"email", "ylarge", 160, 21},
		{"finger", "ylarge", 498, 330},
		{"girl", "ylarge", 399, 599},
		{"iphone", "ylarge", -1, 600},
		{"sketch", "ylarge", -1, 600},
	}
	Prepare(args)
	for _, imgs := range cases {
		width, height := GetImageWidthHeight(args, imgs.img, imgs.style)
		if imgs.expectedWidth != -1 {
			if width != imgs.expectedWidth {
				os.Exit(-1)
			}
		}
		if imgs.expectedHeight != -1 {
			if height != imgs.expectedHeight {
				os.Exit(-1)
			}
		}
	}
	os.Exit(0)
}
