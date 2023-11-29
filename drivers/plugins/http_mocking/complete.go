package http_mocking

import (
	"encoding/json"

	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/utils"
)

type complete struct {
	responseStatus  int
	contentType     string
	responseExample string
	responseSchema  map[string]interface{}
	responseHeader  map[string]string
}

func NewComplete(responseStatus int, contentType string, responseExample string, responseSchema map[string]interface{}, responseHeader map[string]string) *complete {
	return &complete{responseStatus: responseStatus, contentType: contentType, responseExample: responseExample, responseSchema: responseSchema, responseHeader: responseHeader}
}

func (c *complete) Complete(org eocontext.EoContext) error {
	ctx, err := http_context.Assert(org)
	if err != nil {
		return err
	}
	return c.writeHttp(ctx)
}

func (c *complete) writeHttp(ctx http_context.IHttpContext) error {
	ctx.Response().SetHeader("Content-Type", c.contentType)
	ctx.Response().SetStatus(c.responseStatus, "")

	for k, v := range c.responseHeader {
		ctx.Response().SetHeader(k, v)
	}

	if c.responseExample != "" {
		ctx.Response().SetBody([]byte(c.responseExample))
		return nil
	}

	schema := utils.JsonSchemaMockJsUnmarshal(c.responseSchema)
	bytes, err := json.Marshal(schema)
	if err != nil {
		log.Errorf("mocking complete err=%s", err.Error())
		return err
	}

	ctx.Response().SetBody(bytes)
	return nil
}

//func (c *complete) writeDubbo2(ctx dubbo2_context.IDubbo2Context) error {
//	if c.responseExample != "" {
//
//		var val interface{}
//		if err := json.Unmarshal([]byte(c.responseExample), &val); err != nil {
//			ctx.Response().SetBody(Dubbo2ErrorResult(err))
//			return err
//		}
//
//		ctx.Response().SetBody(getDubbo2Response(val, ctx.Proxy().Attachments()))
//		return nil
//	}
//
//	schema := jsonSchemaUnmarshal(c.responseSchema)
//	ctx.Response().SetBody(getDubbo2Response(schema, ctx.Proxy().Attachments()))
//	return nil
//}

//func (c *complete) writeGrpc(ctx grpc_context.IGrpcContext) error {
//	descriptor, err := c.descriptor.Descriptor().FindSymbol(fmt.Sprintf("%s.%s", ctx.Proxy().Service(), ctx.Proxy().Method()))
//	if err != nil {
//		return err
//	}
//	methodDesc := descriptor.GetFile().FindService(ctx.Proxy().Service()).FindMethodByName(ctx.Proxy().Method())
//
//	message := dynamic.NewMessage(methodDesc.GetOutputType())
//
//	fields := message.GetKnownFields()
//	for _, field := range fields {
//		switch field.GetType() {
//		case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE, descriptorpb.FieldDescriptorProto_TYPE_FLOAT: //float32
//			message.SetField(field, gofakeit.Float32())
//		case descriptorpb.FieldDescriptorProto_TYPE_INT64, descriptorpb.FieldDescriptorProto_TYPE_SINT64, descriptorpb.FieldDescriptorProto_TYPE_SFIXED64: //int64
//			message.SetField(field, gofakeit.Int64())
//		case descriptorpb.FieldDescriptorProto_TYPE_INT32, descriptorpb.FieldDescriptorProto_TYPE_SINT32, descriptorpb.FieldDescriptorProto_TYPE_SFIXED32: //int32
//			message.SetField(field, gofakeit.Int32())
//		case descriptorpb.FieldDescriptorProto_TYPE_UINT32, descriptorpb.FieldDescriptorProto_TYPE_FIXED32: //uint32
//			message.SetField(field, gofakeit.Uint32())
//		case descriptorpb.FieldDescriptorProto_TYPE_UINT64, descriptorpb.FieldDescriptorProto_TYPE_FIXED64: //uint64
//			message.SetField(field, gofakeit.Uint64())
//		case descriptorpb.FieldDescriptorProto_TYPE_BOOL: //bool
//			message.SetField(field, gofakeit.Bool())
//		case descriptorpb.FieldDescriptorProto_TYPE_STRING: //string
//			message.SetField(field, gofakeit.LetterN(5))
//		case descriptorpb.FieldDescriptorProto_TYPE_BYTES: //bytes
//			message.SetField(field, []byte(gofakeit.LetterN(5)))
//		}
//
//	}
//
//	ctx.Response().Write(message)
//	return nil
//}
