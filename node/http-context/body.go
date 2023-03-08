package http_context

import (
	"bytes"
	"io"
	"strings"

	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/valyala/fasthttp"

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

// BodyRequestHandler body请求处理器
type BodyRequestHandler struct {
	request *fasthttp.Request

	formdata *multipart.Form
}

func (b *BodyRequestHandler) MultipartForm() (*multipart.Form, error) {
	if b.formdata != nil {
		return b.formdata, nil
	}
	if !strings.Contains(b.ContentType(), MultipartForm) {
		return nil, ErrorNotMultipart
	}
	form, err := b.request.MultipartForm()
	if err != nil {
		return nil, err
	}

	b.formdata = &multipart.Form{
		Value: form.Value,
		File:  form.File,
	}
	b.resetFile()
	return form, nil
}
func (b *BodyRequestHandler) Files() (map[string][]*multipart.FileHeader, error) {
	form, err := b.MultipartForm()
	if err != nil {
		return nil, err
	}
	return form.File, nil
}
func (b *BodyRequestHandler) reset(request *fasthttp.Request) {
	b.request = request
	b.formdata = nil
}
func NewBodyRequestHandler(request *fasthttp.Request) *BodyRequestHandler {
	return &BodyRequestHandler{request: request}
}

// GetForm 获取表单参数
func (b *BodyRequestHandler) GetForm(key string) string {
	contentType, _, _ := mime.ParseMediaType(b.ContentType())

	switch contentType {
	case FormData:
		args := b.request.PostArgs()
		if args == nil {
			return ""
		}
		return string(args.Peek(key))
	case MultipartForm:
		form, err := b.MultipartForm()
		if err != nil {
			return ""
		}
		vs := form.Value[key]
		if len(vs) > 0 {
			return vs[0]
		}
		return ""

	}
	return ""
}

// ContentType 获取contentType
func (b *BodyRequestHandler) ContentType() string {
	return string(b.request.Header.ContentType())
}

// BodyForm 获取表单参数
func (b *BodyRequestHandler) BodyForm() (url.Values, error) {

	contentType, _, _ := mime.ParseMediaType(string(b.request.Header.ContentType()))
	switch contentType {
	case FormData:
		return url.ParseQuery(string(b.request.Body()))
	case MultipartForm:
		multipartForm, err := b.MultipartForm()
		if err != nil {
			return nil, err
		}
		return multipartForm.Value, nil
	default:
		return nil, ErrorNotForm
	}

}

// RawBody 获取raw数据
func (b *BodyRequestHandler) RawBody() ([]byte, error) {
	return b.request.Body(), nil
}

func (b *BodyRequestHandler) GetFile(key string) ([]*multipart.FileHeader, bool) {
	multipartForm, err := b.MultipartForm()
	if err != nil {
		return nil, false
	}
	fl, has := multipartForm.File[key]

	return fl, has

}

func (b *BodyRequestHandler) SetToForm(key, value string) error {
	contentType, _, _ := mime.ParseMediaType(string(b.request.Header.ContentType()))
	switch contentType {
	case FormData:
		b.request.PostArgs().Set(key, value)
		b.request.SetBodyRaw(b.request.PostArgs().QueryString())
		return nil
	case MultipartForm:
		multipartForm, err := b.MultipartForm()
		if err != nil {
			return err
		}
		multipartForm.Value[key] = []string{value}
		return b.resetFile()
	default:
		return ErrorNotForm
	}
}

// AddForm 新增表单参数
func (b *BodyRequestHandler) AddForm(key, value string) error {

	contentType, _, _ := mime.ParseMediaType(string(b.request.Header.ContentType()))
	switch contentType {
	case FormData:
		b.request.PostArgs().Add(key, value)
		b.request.SetBody(b.request.PostArgs().QueryString())
		return nil
	case MultipartForm:
		multipartForm, err := b.MultipartForm()
		if err != nil {
			return err
		}
		multipartForm.Value[key] = append(multipartForm.Value[key], value)
		return b.resetFile()
	default:
		return ErrorNotForm
	}
}

// AddFile 新增文件参数
func (b *BodyRequestHandler) AddFile(key string, file *multipart.FileHeader) error {

	contentType, _, _ := mime.ParseMediaType(b.ContentType())
	if contentType != FormData && contentType != MultipartForm {
		return ErrorNotMultipart
	}
	multipartForm, err := b.MultipartForm()
	if err != nil {
		return err
	}
	multipartForm.File[key] = append(multipartForm.File[key], file)

	return b.resetFile()
}

// SetFile 设置文件参数
func (b *BodyRequestHandler) SetFile(files map[string][]*multipart.FileHeader) error {

	multipartForm, err := b.MultipartForm()
	if err != nil {
		return err
	}
	multipartForm.File = files

	return b.resetFile()
}

func (b *BodyRequestHandler) resetFile() error {
	multipartForm := b.formdata
	if multipartForm == nil {
		return nil
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

			data, err := io.ReadAll(fio)
			if err != nil {
				fio.Close()
				return err
			}
			_, err = part.Write(data)
			if err != nil {
				fio.Close()
				return err
			}
			fio.Close()
		}
	}

	for key, values := range multipartForm.Value {
		//temp := make(url.Values)
		//temp[key] = values
		//value := temp.Encode()
		for _, value := range values {
			err := writer.WriteField(key, value)
			if err != nil {
				return err
			}
		}

	}
	err := writer.Close()
	if err != nil {
		return err
	}
	b.request.Header.SetContentType(writer.FormDataContentType())
	b.request.SetBodyRaw(body.Bytes())
	return nil
}

// SetForm 设置表单参数
func (b *BodyRequestHandler) SetForm(values url.Values) error {

	contentType, _, _ := mime.ParseMediaType(b.ContentType())
	if contentType != FormData && contentType != MultipartForm {
		return ErrorNotForm
	}
	switch contentType {
	case FormData:
		b.request.SetBodyString(values.Encode())
	case MultipartForm:
		multipartForm, err := b.MultipartForm()
		if err != nil {
			return err
		}
		multipartForm.Value = values
		return b.resetFile()
	}

	return ErrorNotForm
}

// SetRaw 设置raw数据
func (b *BodyRequestHandler) SetRaw(contentType string, body []byte) {
	b.request.SetBodyRaw(body)
	b.request.Header.SetContentType(contentType)

}
