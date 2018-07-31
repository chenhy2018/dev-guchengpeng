package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"qbox.us/account"
	"github.com/qiniu/log.v1"
	saccount "qbox.us/servend/account"
	"qbox.us/servend/oauth"
	"strconv"
	"strings"
)

type UserToken struct {
	Agent string `json:"agent"`
	Appid uint32 `json:"appid"`
	Devid uint32 `json:"devid"`
	Uid   uint32 `json:"uid"`
	Utype uint32 `json:"Utype"`
}

type Request struct {
	ContentLength string    `json:"Content-Length"`
	ContentType   string    `json:"Content-Type"`
	Token         UserToken `json:"Token"`
}

type Response struct {
	Fh string `json:"fh"`
}

type BatchInfo struct {
	Op []string `json:"op"`
}

func replay(rspubHost string, acc saccount.Interface, line string) (err error) {
	querys := strings.Split(line, "\t")

	path := querys[4]
	requestJson := querys[5]
	responseJson := querys[8]

	r := new(Request)
	err = json.Unmarshal([]byte(requestJson), r)
	if err != nil {
		log.Error(err)
		return
	}
	res := new(Response)
	err = json.Unmarshal([]byte(responseJson), res)
	if err != nil {
		log.Error(err)
		return
	}

	var body io.Reader
	length := int64(-1)
	if strings.HasPrefix(path, "/batch") {
		batchInfo := querys[6]
		bi := new(BatchInfo)
		err = json.Unmarshal([]byte(batchInfo), bi)
		if err != nil {
			log.Error(err)
			return
		}
		b := url.Values(map[string][]string{"op": bi.Op}).Encode()
		body = strings.NewReader(b)
		length = int64(len(b))
	} else {
		var b []byte
		b, err = base64.StdEncoding.DecodeString(res.Fh)
		if err != nil {
			log.Error(err)
			return
		}
		body = bytes.NewReader(b)
		length = int64(len(b))
	}
	req, err := http.NewRequest("POST", rspubHost+path, body)
	if err != nil {
		log.Error(err)
		return
	}
	req.Header.Set("Content-Type", r.ContentType)
	req.Header.Set("User-Agent", "qiniu-rsreplay-v1.0")
	req.ContentLength = length

	user := saccount.UserInfo{
		Id:    "<unknown>",
		Agent: r.Token.Agent,
		Appid: r.Token.Appid,
		Devid: r.Token.Devid,
		Uid:   r.Token.Uid,
		Utype: r.Token.Utype,
	}
	conn := oauth.NewClient(acc.MakeAccessToken(user), nil)
	resp, err := conn.Do(req)
	if err != nil {
		log.Error(err)
		return
	}
	defer resp.Body.Close()
	b := new(bytes.Buffer)
	io.Copy(b, resp.Body)
	msg := b.String()
	if resp.StatusCode/100 != 2 {
		err = errors.New(strconv.Itoa(resp.StatusCode) + msg)
		log.Error(err)
		return
	}
	log.Info(line)
	log.Info(resp.StatusCode, msg)
	return
}

func main() {
	debugLevel := flag.Int("debuglevel", 2, "set the debug level, default is 2(Warning)")
	flag.Parse()
	log.SetOutputLevel(*debugLevel)
	if len(flag.Args()) != 2 {
		fmt.Println("\n\nUsage: rsreplay <rspub host> <auditlog.txt>\n\n")
		return
	}

	f, err := os.Open(flag.Arg(1))
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	rspubHost := flag.Arg(0)

	acc := account.Account{}
	if err != nil {
		log.Fatalln("digest_auth.New failed:", err)
	}

	br := bufio.NewReader(f)

	for line, err := br.ReadString('\n'); err == nil; line, err = br.ReadString('\n') {
		log.Debug(line)
		err := replay(rspubHost, acc, line)
		if err != nil {
			log.Fatalln(line, err)
		}
	}
}
