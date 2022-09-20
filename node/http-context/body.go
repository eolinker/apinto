package http_context

import (
	"bytes"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/valyala/fasthttp"

	"io/ioutil"
	"net/http"

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
	//form        url.Values
	//contentType string
	//files       map[string]*http_service.FileHeader
	//
	//isInit     bool
	//isWriteRaw bool
	//
	//object interface{}
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
	//contentType, _, _ := mime.ParseMediaType(b.contentType)
	//if contentType != FormData && contentType != MultipartForm {
	//	return ""
	//}
	//b.parse()
	//
	//if !b.isInit || b.form == nil {
	//	return ""
	//}
	//return b.form.Get(key)
}

//ContentType 获取contentType
func (b *BodyRequestHandler) ContentType() string {
	return string(b.request.Header.ContentType())
}

//BodyForm 获取表单参数
func (b *BodyRequestHandler) BodyForm() (url.Values, error) {
	args := b.request.PostArgs()

	return url.ParseQuery(args.String())
	//
	//err := b.parse()
	//if err != nil {
	//	return nil, err
	//}
	//return b.form, nil
}

//RawBody 获取raw数据
func (b *BodyRequestHandler) RawBody() ([]byte, error) {
	return b.request.Body(), nil
	//err := b.encode()
	//if err != nil {
	//	return nil, err
	//}
	//return b.rawBody(), nil

}

////Files 获取文件参数
//func (b *BodyRequestHandler) Files() (map[string]*http_service.FileHeader, error) {
//
//	err := b.parse()
//
//	if err != nil {
//		return nil, err
//	}
//	return b.files, nil
//}

func (b *BodyRequestHandler) GetFile(key string) ([]*multipart.FileHeader, bool) {
	multipartForm, err := b.request.MultipartForm()
	if err != nil {
		return nil, false
	}
	fl, has := multipartForm.File[key]

	return fl, has

}

//
////parse 解析
//func (b *BodyRequestHandler) parse() error {
//	if b.isInit {
//		return nil
//	}
//	contentType, _, _ := mime.ParseMediaType(b.contentType)
//	switch contentType {
//	case JSON:
//		{
//			e := json.Unmarshal(b.rawBody(), &b.object)
//			if e != nil {
//				return e
//			}
//		}
//	case AppLicationXML, TextXML:
//		{
//			e := xml.Unmarshal(b.rawBody(), &b.object)
//			if e != nil {
//				return e
//			}
//
//		}
//
//	case MultipartForm:
//		{
//			r, err := multipartReader(b.contentType, false, b.rawBody())
//			if err != nil {
//				return err
//			}
//			form, err := r.ReadForm(defaultMultipartMemory)
//			if err != nil {
//				return err
//			}
//
//			if b.form == nil {
//				b.form = make(url.Values)
//			}
//			for k, v := range form.Value {
//				b.form[k] = append(b.form[k], v...)
//			}
//
//			b.files = make(map[string]*http_service.FileHeader)
//			for k, fs := range form.File {
//
//				if len(fs) > 0 {
//					file, err := fs[0].Open()
//					if err != nil {
//						return err
//					}
//					fileData, err := ioutil.ReadAll(file)
//					if err != nil {
//						return err
//					}
//
//					b.files[k] = &http_service.FileHeader{
//						FileName: fs[0].Filename,
//						Data:     fileData,
//						Header:   fs[0].Header,
//					}
//				}
//			}
//
//			b.object = b.form
//		}
//	case FormData:
//		{
//			form, err := url.ParseQuery(string(b.rawBody()))
//			if err != nil {
//				return err
//			}
//			if b.form == nil {
//				b.form = form
//			} else {
//				for k, v := range form {
//					b.form[k] = append(b.form[k], v...)
//				}
//
//			}
//			b.object = b.form
//
//		}
//	}
//	b.isInit = true
//	return nil
//}

//SetToForm 设置表单参数
func (b *BodyRequestHandler) SetToForm(key, value string) error {
	b.request.PostArgs().Set(key, value)
	//contentType, _, _ := mime.ParseMediaType(b.contentType)
	//if contentType != FormData && contentType != MultipartForm {
	//	return ErrorNotForm
	//}
	//
	//err := b.parse()
	//if err != nil {
	//	return err
	//}
	//b.isWriteRaw = false
	//
	//if b.form == nil {
	//	b.form = make(url.Values)
	//}
	//b.form.Set(key, value)
	//b.isWriteRaw = false

	return nil
}

//AddForm 新增表单参数
func (b *BodyRequestHandler) AddForm(key, value string) error {

	contentType, _, _ := mime.ParseMediaType(string(b.request.Header.ContentType()))
	if contentType != FormData && contentType != MultipartForm {
		return ErrorNotForm
	}
	b.request.PostArgs().Add(key, value)
	return nil
	//err := b.parse()
	//if err != nil {
	//	return err
	//}
	//b.isWriteRaw = false
	//
	//if b.form == nil {
	//	b.form = make(url.Values)
	//}
	//b.form.Add(key, value)
	return nil
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
	b.SetFile(multipartForm.File)
	return nil
}

//SetFile 设置文件参数
func (b *BodyRequestHandler) SetFile(files map[string][]*multipart.FileHeader) error {

	multipartForm, err := b.request.MultipartForm()
	if err != nil {
		return err
	}
	multipartForm.File = files

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

////Clone 克隆body
//func (b *BodyRequestHandler) Clone() *BodyRequestHandler {
//	rawBody, _ := b.RawBody()
//	return newBodyRequestHandler(b.contentType, rawBody)
//
//}

////BodyInterface 获取请求体对象
//func (b *BodyRequestHandler) BodyInterface() (interface{}, error) {
//	err := b.parse()
//	if err != nil {
//		return nil, err
//	}
//
//	return b.object, nil
//}
//
////encode encode
//func (b *BodyRequestHandler) encode() error {
//	if b.isWriteRaw {
//		return nil
//	}
//
//	contentType, _, _ := mime.ParseMediaType(b.contentType)
//	if contentType != FormData && contentType != MultipartForm {
//		b.isWriteRaw = true
//		return nil
//	}
//
//	if len(b.files) > 0 {
//		body := new(bytes.Buffer)
//		writer := multipart.NewWriter(body)
//
//		for name, file := range b.files {
//			part, err := writer.CreateFormFile(name, file.FileName)
//			if err != nil {
//				return err
//			}
//			_, err = part.Write(file.Data)
//			if err != nil {
//				return err
//			}
//		}
//
//		for key, values := range b.form {
//			temp := make(url.Values)
//			temp[key] = values
//			value := temp.Encode()
//			err := writer.WriteField(key, value)
//			if err != nil {
//				return err
//			}
//		}
//		err := writer.Close()
//		if err != nil {
//			return err
//		}
//		b.contentType = writer.FormDataContentType()
//		b.request.SetBodyRaw(body.Bytes())
//		b.isWriteRaw = true
//	} else {
//		if b.form != nil {
//			b.request.SetBodyRaw([]byte(b.form.Encode()))
//		} else {
//			b.request.ResetBody()
//		}
//	}
//	return nil
//}

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
func (b *BodyRequestHandler) rawBody() []byte {
	return b.request.Body()
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
