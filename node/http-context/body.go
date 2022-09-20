package http_context

import (
	"bytes"

	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/valyala/fasthttp"

	"io/ioutil"

	"mime"
	"mime/multipart"
	"net/url"
)

const defaultMultipartMemory = 32 << 20 // 32 MB
var (
	_ http_context.IBodyDataWriter = (*BodyRequestHandler)(nil)
)

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
	request *fasthttp.Request
}

func (b *BodyRequestHandler) Files() (map[string][]*multipart.FileHeader, error) {
	form, err := b.request.MultipartForm()
	if err != nil {
		return nil, err
	}
	return form.File, nil
}

func NewBodyRequestHandler(request *fasthttp.Request) *BodyRequestHandler {
	return &BodyRequestHandler{request: request}
}

//GetForm 获取表单参数
func (b *BodyRequestHandler) GetForm(key string) string {
	args := b.request.PostArgs()
	if args == nil {
		return ""
	}
	return string(args.Peek(key))
}

//ContentType 获取contentType
func (b *BodyRequestHandler) ContentType() string {
	return string(b.request.Header.ContentType())
}

//BodyForm 获取表单参数
func (b *BodyRequestHandler) BodyForm() (url.Values, error) {
	args := b.request.PostArgs()

	return url.ParseQuery(args.String())
}

//RawBody 获取raw数据
func (b *BodyRequestHandler) RawBody() ([]byte, error) {
	return b.request.Body(), nil
}

func (b *BodyRequestHandler) GetFile(key string) ([]*multipart.FileHeader, bool) {
	multipartForm, err := b.request.MultipartForm()
	if err != nil {
		return nil, false
	}
	fl, has := multipartForm.File[key]

	return fl, has

}

//SetToForm 设置表单参数
func (b *BodyRequestHandler) SetToForm(key, value string) error {
	contentType, _, _ := mime.ParseMediaType(string(b.request.Header.ContentType()))
	switch contentType {
	case FormData:
		b.request.PostArgs().Set(key, value)
		b.request.SetBody(b.request.PostArgs().QueryString())
		return nil
	case MultipartForm:
		multipartForm, err := b.request.MultipartForm()
		if err != nil {
			return err
		}
		multipartForm.Value[key] = []string{value}
		return b.resetFile()
	default:
		return ErrorNotForm
	}
}

//AddForm 新增表单参数
func (b *BodyRequestHandler) AddForm(key, value string) error {

	contentType, _, _ := mime.ParseMediaType(string(b.request.Header.ContentType()))
	switch contentType {
	case FormData:
		b.request.PostArgs().Add(key, value)
		b.request.SetBody(b.request.PostArgs().QueryString())
		return nil
	case MultipartForm:
		multipartForm, err := b.request.MultipartForm()
		if err != nil {
			return err
		}
		multipartForm.Value[key] = append(multipartForm.Value[key], value)
		return b.resetFile()
	default:
		return ErrorNotForm
	}
}

//AddFile 新增文件参数
func (b *BodyRequestHandler) AddFile(key string, file *multipart.FileHeader) error {

	contentType, _, _ := mime.ParseMediaType(b.ContentType())
	if contentType != FormData && contentType != MultipartForm {
		return ErrorNotMultipart
	}
	multipartForm, err := b.request.MultipartForm()
	if err != nil {
		return err
	}
	multipartForm.File[key] = append(multipartForm.File[key], file)

	return b.resetFile()
}

//SetFile 设置文件参数
func (b *BodyRequestHandler) SetFile(files map[string][]*multipart.FileHeader) error {

	multipartForm, err := b.request.MultipartForm()
	if err != nil {
		return err
	}
	multipartForm.File = files

	return b.resetFile()
}

func (b *BodyRequestHandler) resetFile() error {
	multipartForm, err := b.request.MultipartForm()
	if err != nil {
		return err
	}
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for name, fs := range multipartForm.File {
		for _, f := range fs {
			fio, err := f.Open()
			if err != nil {
				return err
			}

			part, err := writer.CreateFormFile(name, f.Filename)
			if err != nil {
				fio.Close()
				return err
			}
			data, err := ioutil.ReadAll(fio)
			if err != nil {
				fio.Close()
				return err
			}
			_, err = part.Write(data)
			if err != nil {
				return err
			}
		}
	}

	for key, values := range multipartForm.Value {
		temp := make(url.Values)
		temp[key] = values
		value := temp.Encode()
		err := writer.WriteField(key, value)
		if err != nil {
			return err
		}
	}
	err = writer.Close()
	if err != nil {
		return err
	}
	b.request.Header.SetContentType(writer.FormDataContentType())
	b.request.SetBodyRaw(body.Bytes())
	return nil
}

//SetForm 设置表单参数
func (b *BodyRequestHandler) SetForm(values url.Values) error {

	contentType, _, _ := mime.ParseMediaType(b.ContentType())
	if contentType != FormData && contentType != MultipartForm {
		return ErrorNotForm
	}
	switch contentType {
	case FormData:
		b.request.PostArgs().Parse(values.Encode())
	}

	return nil
}

//SetRaw 设置raw数据
func (b *BodyRequestHandler) SetRaw(contentType string, body []byte) {
	b.request.SetBodyRaw(body)
	b.request.SetBody(body)
	b.request.Header.SetContentType(contentType)
	return

}
