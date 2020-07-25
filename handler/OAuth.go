package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

/*
	基于GIN的 OAUTH1.0 签名与身份认证
	URL中的query参数一律不参与签名
*/

type oauth struct {
	Key     string
	Secret  string
	Strict bool     //严格模式，body会参与签名
	Timeout int32  //超时秒
}

const OAuth_Signature_Name  = "oauth_signature"

var oauthParams = []string{"oauth_consumer_key", "oauth_nonce", "oauth_timestamp", "oauth_version", "oauth_signature_method"}

func NewOAuth() *oauth {
	return &oauth{Key: "oauth1.0", Secret: "szmzbzbzlp@20200712", Timeout: 5}
}

func (this *oauth) NewOAuthParams() map[string]string {
	oauth := make(map[string]string)
	oauth["oauth_consumer_key"] = this.Key
	oauth["oauth_version"] = "1.0"
	oauth["oauth_signature_method"] = "HMAC-SHA1"
	oauth["oauth_nonce"] = strconv.FormatInt(int64(rand.Int31n(8999)+1000), 10)
	oauth["oauth_timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	return oauth
}

//签名Signature
//method GET Insert
//url:protocol://hostname/path
//body JSON字符串
func (this *oauth) Signature(method, url string, oauth map[string]string,body string) string {
	arr := []string{method, url}
	for _, k := range oauthParams {
		arr = append(arr, k+"="+oauth[k])
	}
	arr = append(arr, body,this.Secret)
	str := strings.Join(arr, "&")
	return HMACSHA1(this.Secret, str)
}

//Verify http(s)验签
func (this *oauth) Verify(ctX *gin.Context) error {
	signature := ctX.GetHeader(OAuth_Signature_Name)
	if signature == ""{
		return errors.New("OAuth Signature empty")
	}

	OAuthMap := make(map[string]string)
	for _, k := range oauthParams {
		v := ctX.GetHeader(k)
		if v == "" {
			return errors.New("OAuth Params empty")
		}
		OAuthMap[k] = v
	}
	if OAuthMap["oauth_consumer_key"] != this.Key {
		return errors.New("oauth_consumer_key error")
	}
	oauthTimeStamp, err := strconv.ParseInt(OAuthMap["oauth_timestamp"], 10, 64)
	if err != nil {
		return err
	}

	requestTime := time.Now().Unix() - oauthTimeStamp
	if requestTime < 0 || requestTime > int64(this.Timeout) {
		return errors.New("oauth request timeout")
	}

	var arrUrl []string
	var strBody string
	if this.Strict{
		//TODO
	}
	proto := strings.Split(ctX.Request.Proto, "/")
	arrUrl = append(arrUrl, strings.ToLower(proto[0]), "://", ctX.Request.Host, ctX.Request.URL.Path)
	newSignature := this.Signature(ctX.Request.Method, strings.Join(arrUrl, ""), OAuthMap,strBody)

	if signature != newSignature {
		return errors.New("oauth signature error")
	}
	return nil
}

//args
func (this *oauth) request(method, url string, data []byte, header map[string]string) *Message {
	var (
		err   error
		req   *http.Request
		res   *http.Response
		reply []byte
	)
	req, err = http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return NewErrMsgReply(err.Error())
	}
	defer req.Body.Close()

	if header == nil {
		header = make(map[string]string)
	}

	if _, ok := header["content-type"]; !ok {
		header["content-type"] = "application/json;charset=utf-8"
	}
	for k, v := range header {
		req.Header.Add(k, v)
	}
	//设置签名
	OAuthMap := this.NewOAuthParams()
	for k, v := range OAuthMap {
		req.Header.Add(k, v)
	}
	arrUrl := strings.Split(url,"?")
	var strBody string
	if this.Strict{
		strBody = string(data)
	}
	signature := this.Signature(method, arrUrl[0], OAuthMap,strBody)
	req.Header.Add(OAuth_Signature_Name, signature)

	client := &http.Client{Timeout: time.Duration(this.Timeout) * time.Second}
	res, err = client.Do(req)
	if err != nil {
		return NewMsgError(err)
	}
	defer res.Body.Close()

	reply, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return NewMsgError(err)
	} else if res.StatusCode != http.StatusOK {
		return NewErrMsgReply(res.Status)
	}

	if Config.ErrHeaderName != "" {
		var code int
		code, err = strconv.Atoi(res.Header.Get(Config.ErrHeaderName))
		if err != nil {
			return NewMsgError(err)
		} else if code != GetErrCode(ErrMsg_NAME_SUCCESS) {
			return NewErrMsgReply(string(reply), code)
		}
	}
	return NewMsgReply(reply)
}

func (this *oauth) GET(url string, data []byte, header map[string]string) *Message {
	return this.request("GET", url, data, header)
}

func (this *oauth) POST(url string, data []byte, header map[string]string) *Message {
	return this.request("POST", url, data, header)
}

func (this *oauth) PostJson(url string, query interface{}) *Message {
	data,err := json.Marshal(query)
	if err!=nil{
		return NewErrMsgReply(err.Error())
	}
	return this.request("POST", url, data, nil)
}