package base

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"io"
	"io/ioutil"
	. "logger"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strings"
)

// Http Get
func Get(url string, params map[string]interface{}) ([]byte, error) {

	if len(params) > 0 {
		url += ("?" + ToRequest(params))
	}

	DEBUG("HTTP GET: %s", url)

	resp, err := http.Get(url)

	if nil != err {
		ERROR("HTTP.GET: %s %s", url, err.Error())
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, ToError("%s", resp.Status)
	}

	// 读取所有返回内容
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		ERROR("HTTP.GET.ReadAll: %s %s", url, err.Error())
		return nil, err
	}

	return body, nil
}

// Http Post
// application/octet-stream
// image/jpeg
func Post(url string, bodyType string, data []byte) ([]byte, error) {

	buffer := bytes.NewBuffer(data)

	resp, err := http.Post(url, bodyType, buffer)

	if err != nil {
		ERROR("HTTP.POST: %s %s", url, err.Error())
		return nil, err
	}

	defer resp.Body.Close()

	// 读取所有返回内容
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		ERROR("HTTP.POST.ReadAll: %s %s", url, err.Error())
		return nil, err
	}

	return body, nil
}

// Http PostForm
func PostForm(surl string, params map[string]interface{}) ([]byte, error) {

	form := url.Values{}

	for k, v := range params {

		if !IsSlice(v) {
			form.Add(k, ToString(v))
			continue
		}

		val := reflect.ValueOf(v)

		for i, length := 0, val.Len(); i < length; i++ {
			elem := val.Index(i).Interface()
			form.Add(k, ToString(elem))
		}
	}

	resp, err := http.PostForm(surl, form)

	if err != nil {
		ERROR("HTTP.PostForm: %s %s", surl, err.Error())
		return nil, err
	}

	defer resp.Body.Close()

	// 读取所有返回内容
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		ERROR("HTTP.POST.ReadAll: %s %s", surl, err.Error())
		return nil, err
	}

	return body, nil
}

// Http Post Json
func PostJson(url string, object interface{}) ([]byte, error) {

	data, err := ToJsonData(object)

	if err != nil {
		return nil, err
	}

	return Post(url, "text/json", data)
}

// 下载指定文件到指定目录
func AsynDownload(url, folder string, callback func(error, string, int64)) {

	go func() {
		err, name, num := Download(url, folder)

		callback(err, name, num)
	}()
}

// 判断文件是否已经存在
func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return nil == err || os.IsExist(err)
}

// http 下载
func Download(url, folder string, suffix ...string) (error, string, int64) {

	fileName := FileName(url)

	if err := os.MkdirAll(folder, 0777); err != nil {
		if !os.IsExist(err) {
			LOG_ERROR("创建文件夹失败! %s", err.Error())
			return err, fileName, 0
		}
	}

	if len(suffix) > 0 {
		if '.' != suffix[0][0] {
			fileName += "."
		}

		fileName += suffix[0]
	}

	st, err := os.Stat(folder + fileName)

	if nil == err || os.IsExist(err) {
		return err, fileName, st.Size()
	}

	resp, err := http.Get(url)

	if err != nil {
		return err, fileName, 0
	}

	if folder[len(folder)-1] != '/' {
		folder += "/"
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ToError("%s", resp.Status), fileName, 0
	}

	file, err := os.Create(folder + fileName)

	if err != nil {
		LOG_ERROR("创建文件失败: %s", err.Error())
		return err, fileName, 0
	}

	io.Copy(file, resp.Body)

	file.Close()

	return nil, fileName, resp.ContentLength
}

func ToCookies(cookies map[string]interface{}) string {

	var cookiepair = make([]string, 0, len(cookies))

	for key, val := range cookies {
		cookiepair = append(cookiepair, Sprintf("%s=%v", key, val))
	}

	return strings.Join(cookiepair, ";")
}

func ToRequest(params map[string]interface{}) string {

	var requestpair = make([]string, 0, len(params))

	for key, val := range params {
		requestpair = append(requestpair, Sprintf("%s=%v", key, val))
	}

	return strings.Join(requestpair, "&")
}

type RequestItem struct {
	Key   string
	Value interface{}
}

func NewRequestItem(key string, val interface{}) *RequestItem {
	return &RequestItem{Key: key, Value: val}
}

type RequestItems []*RequestItem

func (m RequestItems) Len() int           { return len(m) }
func (m RequestItems) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m RequestItems) Less(i, j int) bool { return m[i].Key < m[j].Key }

func ToSortedRequest(params map[string]interface{}) string {

	requestItems := make(RequestItems, 0, len(params))

	for key, val := range params {
		requestItems = append(requestItems, NewRequestItem(key, val))
	}

	sort.Sort(requestItems)

	reqpair := make([]string, 0, len(params))

	for _, item := range requestItems {
		reqpair = append(reqpair, Sprintf("%s=%v", item.Key, item.Value))
	}

	return strings.Join(reqpair, "&")
}

// 专为访问腾讯支付接口
func TencentGet(url_s, app_key string, cookies, params map[string]interface{}, result interface{}) error {

	delete(params, "sig")

	u, err := url.Parse(url_s)

	if err != nil {
		ERROR("url.Parse: %s", err.Error())
		return err
	}

	cookies["org_loc"] = url.QueryEscape(u.Path)

	requests := ToSortedRequest(params)

	source := Sprintf("GET&%s&%s", url.QueryEscape(u.Path), url.QueryEscape(requests))

	hash := hmac.New(sha1.New, []byte((app_key + "&")))
	hash.Write([]byte(source))

	sig := Base64Byte(hash.Sum(nil))

	url_s = Sprintf("%s?%s&sig=%s", url_s, ToRequest(params), url.QueryEscape(sig))

	//	LOG_DEBUG( "URL: %s", url_s )

	httpClient := &http.Client{}

	req, err := http.NewRequest("GET", url_s, nil)

	if err != nil {
		ERROR("http.NewRequest: %s", err.Error())
		return err
	}

	req.Header.Set("Cookie", ToCookies(cookies))
	req.Header.Set("Connection", "keep-alive")

	resp, err := httpClient.Do(req)

	if err != nil {
		ERROR("httpClient.Do: %s", err.Error())
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		ERROR("ioutil.ReadAll: %s", err.Error())
		return err
	}

	DEBUG("TencentGet: %s %s", u.Path, string(body))

	if err := ToJsonObject(body, result); err != nil {
		LOG_ERROR("ToJsonObject: %s", string(body))
		return err
	}

	return nil
}
