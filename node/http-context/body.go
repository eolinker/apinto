package http_context

import (
	"bytes"
	"encoding/json"
	"encoding/xml"

	"github.com/valyala/fasthttp"

	http_service "github.com/eolinker/eosc/http-service"

	"io/ioutil"
	"net/http"

	"mime"
	"mime/multipart"
	"net/url"
)

const defaultMultipartMemory = 32 << 20 // 32 MB

const (
	MultipartForm  = "multipart/form-data"
	FormData       = "application/x-www-form-urlencoded"
	TEXT           = "text/plain"
	JSON           = "application/json"
	JavaScript     = "application/javascript"
	AppLicationXML = "application/xml"
	TextXML        = "text/xml"
	Html           = "text/html"
)

//BodyRequestHandler body请求处理器
type BodyRequestHandler struct {
	request         *fasthttp.Request
	form            url.Values
	rawBody         []byte
	orgContentParam map[string]string
	contentType     string
	files           map[string]*http_service.FileHeader

	isInit     bool
	isWriteRaw bool

	object interface{}
}

func NewBodyRequestHandler(request *fasthttp.Request) *BodyRequestHandler {
	return &BodyRequestHandler{request: request}
}

//GetForm 获取表单参数
func (b *BodyRequestHandler) GetForm(key string) string {

	contentType, _, _ := mime.ParseMediaType(b.contentType)
	if contentType != FormData && contentType != MultipartForm {
		return ""
	}
	b.parse()

	if !b.isInit || b.form == nil {
		return ""
	}
	return b.form.Get(key)
}

//ContentType 获取contentType
func (b *BodyRequestHandler) ContentType() string {
	return b.contentType
}

//BodyForm 获取表单参数
func (b *BodyRequestHandler) BodyForm() (url.Values, error) {

	err := b.parse()
	if err != nil {
		return nil, err
	}
	return b.form, nil
}

//RawBody 获取raw数据
func (b *BodyRequestHandler) RawBody() ([]byte, error) {

	err := b.encode()
	if err != nil {
		return nil, err
	}
	return b.rawBody, nil

}

//Files 获取文件参数
func (b *BodyRequestHandler) Files() (map[string]*http_service.FileHeader, error) {

	err := b.parse()

	if err != nil {
		return nil, err
	}
	return b.files, nil
}

func (b *BodyRequestHandler) GetFile(key string) (file *http_service.FileHeader, has bool) {
	err := b.parse()

	if err != nil {
		return nil, false
	}
	file, has = b.files[key]
	return file, has
}

//parse 解析
func (b *BodyRequestHandler) parse() error {
	if b.isInit {
		return nil
	}
	contentType, _, _ := mime.ParseMediaType(b.contentType)
	switch contentType {
	case JSON:
		{
			e := json.Unmarshal(b.rawBody, &b.object)
			if e != nil {
				return e
			}
		}
	case AppLicationXML, TextXML:
		{
			e := xml.Unmarshal(b.rawBody, &b.object)
			if e != nil {
				return e
			}

		}

	case MultipartForm:
		{
			r, err := multipartReader(b.contentType, false, b.rawBody)
			if err != nil {
				return err
			}
			form, err := r.ReadForm(defaultMultipartMemory)
			if err != nil {
				return err
			}

			if b.form == nil {
				b.form = make(url.Values)
			}
			for k, v := range form.Value {
				b.form[k] = append(b.form[k], v...)
			}

			b.files = make(map[string]*http_service.FileHeader)
			for k, fs := range form.File {

				if len(fs) > 0 {
					file, err := fs[0].Open()
					if err != nil {
						return err
					}
					fileData, err := ioutil.ReadAll(file)
					if err != nil {
						return err
					}

					b.files[k] = &http_service.FileHeader{
						FileName: fs[0].Filename,
						Data:     fileData,
						Header:   fs[0].Header,
					}
				}
			}

			b.object = b.form
		}
	case FormData:
		{
			form, err := url.ParseQuery(string(b.rawBody))
			if err != nil {
				return err
			}
			if b.form == nil {
				b.form = form
			} else {
				for k, v := range form {
					b.form[k] = append(b.form[k], v...)
				}

			}
			b.object = b.form

		}
	}
	b.isInit = true
	return nil
}

//SetToForm 设置表单参数
func (b *BodyRequestHandler) SetToForm(key, value string) error {

	contentType, _, _ := mime.ParseMediaType(b.contentType)
	if contentType != FormData && contentType != MultipartForm {
		return ErrorNotForm
	}

	err := b.parse()
	if err != nil {
		return err
	}
	b.isWriteRaw = false

	if b.form == nil {
		b.form = make(url.Values)
	}
	b.form.Set(key, value)
	b.isWriteRaw = false

	return nil
}

//AddForm 新增表单参数
func (b *BodyRequestHandler) AddForm(key, value string) error {
	contentType, _, _ := mime.ParseMediaType(b.contentType)
	if contentType != FormData && contentType != MultipartForm {
		return ErrorNotForm
	}
	err := b.parse()
	if err != nil {
		return err
	}
	b.isWriteRaw = false

	if b.form == nil {
		b.form = make(url.Values)
	}
	b.form.Add(key, value)
	return nil
}

//AddFile 新增文件参数
func (b *BodyRequestHandler) AddFile(key string, file *http_service.FileHeader) error {

	contentType, _, _ := mime.ParseMediaType(b.contentType)
	if contentType != FormData && contentType != MultipartForm {
		return ErrorNotMultipart
	}
	err := b.parse()
	if err != nil {
		return err
	}
	b.isWriteRaw = false
	if file == nil && b.files != nil {
		delete(b.files, key)
		return nil
	}
	if b.files == nil {
		b.files = make(map[string]*http_service.FileHeader)
	}
	b.files[key] = file

	return nil
}

////Clone 克隆body
//func (b *BodyRequestHandler) Clone() *BodyRequestHandler {
//	rawBody, _ := b.RawBody()
//	return newBodyRequestHandler(b.contentType, rawBody)
//
//}

//BodyInterface 获取请求体对象
func (b *BodyRequestHandler) BodyInterface() (interface{}, error) {
	err := b.parse()
	if err != nil {
		return nil, err
	}

	return b.object, nil
}

//encode encode
func (b *BodyRequestHandler) encode() error {
	if b.isWriteRaw {
		return nil
	}

	contentType, _, _ := mime.ParseMediaType(b.contentType)
	if contentType != FormData && contentType != MultipartForm {
		b.isWriteRaw = true
		return nil
	}

	if len(b.files) > 0 {
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)

		for name, file := range b.files {
			part, err := writer.CreateFormFile(name, file.FileName)
			if err != nil {
				return err
			}
			_, err = part.Write(file.Data)
			if err != nil {
				return err
			}
		}

		for key, values := range b.form {
			temp := make(url.Values)
			temp[key] = values
			value := temp.Encode()
			err := writer.WriteField(key, value)
			if err != nil {
				return err
			}
		}
		err := writer.Close()
		if err != nil {
			return err
		}
		b.contentType = writer.FormDataContentType()
		b.rawBody = body.Bytes()
		b.isWriteRaw = true
	} else {
		if b.form != nil {
			b.rawBody = []byte(b.form.Encode())
		} else {
			b.rawBody = make([]byte, 0, 0)
		}
	}
	return nil
}

//SetForm 设置表单参数
func (b *BodyRequestHandler) SetForm(values url.Values) error {

	contentType, _, _ := mime.ParseMediaType(b.contentType)
	if contentType != FormData && contentType != MultipartForm {
		return ErrorNotForm
	}
	b.parse()
	b.form = values
	b.isWriteRaw = false

	return nil
}

//SetFile 设置文件参数
func (b *BodyRequestHandler) SetFile(files map[string]*http_service.FileHeader) error {

	contentType, _, _ := mime.ParseMediaType(b.contentType)
	if contentType != FormData && contentType != MultipartForm {
		return ErrorNotForm
	}
	b.parse()
	b.files = files
	// b.form = values
	b.isWriteRaw = false

	return nil
}

//SetRaw 设置raw数据
func (b *BodyRequestHandler) SetRaw(contentType string, body []byte) {
	b.rawBody, b.contentType, b.isInit, b.isWriteRaw = body, contentType, false, true
	return

}

////newBodyRequestHandler 创建body请求处理器
//func newBodyRequestHandler(contentType string, body []byte) *BodyRequestHandler {
//	b := new(BodyRequestHandler)
//	b.SetRaw(contentType, body)
//	return b
//}

func multipartReader(contentType string, allowMixed bool, raw []byte) (*multipart.Reader, error) {

	if contentType == "" {
		return nil, http.ErrNotMultipart
	}
	d, params, err := mime.ParseMediaType(contentType)
	if err != nil || !(d == "multipart/form-data" || allowMixed && d == "multipart/mixed") {
		return nil, http.ErrNotMultipart
	}
	boundary, ok := params["boundary"]
	if !ok {
		return nil, http.ErrMissingBoundary
	}
	body := ioutil.NopCloser(bytes.NewBuffer(raw))
	return multipart.NewReader(body, boundary), nil
}
