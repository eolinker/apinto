package http_context

type IStatus interface {
	SetStatus(statusCode int)
	GetStatus() int
}

func (ctx *Context) SetStatus(statusCode int) {
	ctx.context.SetStatusCode(statusCode)
}

func (ctx *Context) GetStatus() int {
	return ctx.context.Response.StatusCode()
}
