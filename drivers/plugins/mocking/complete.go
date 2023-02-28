package mocking

import (
	"dubbo.apache.org/dubbo-go/v3/protocol"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo/impl"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	grpc_descriptor "github.com/eolinker/apinto/grpc-descriptor"
	"github.com/eolinker/eosc/eocontext"
	dubbo2_context "github.com/eolinker/eosc/eocontext/dubbo2-context"
	grpc_context "github.com/eolinker/eosc/eocontext/grpc-context"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/protobuf/types/descriptorpb"
)

type complete struct {
	responseStatus  int
	contentType     string
	responseExample string
	responseSchema  map[string]interface{}
	descriptor      grpc_descriptor.IDescriptor
}

func NewComplete(responseStatus int, contentType string, responseExample string, responseSchema map[string]interface{}, descriptor grpc_descriptor.IDescriptor) *complete {
	return &complete{responseStatus: responseStatus, contentType: contentType, responseExample: responseExample, responseSchema: responseSchema, descriptor: descriptor}
}

func (c *complete) Complete(org eocontext.EoContext) error {
	switch ctx := org.(type) {
	case dubbo2_context.IDubbo2Context:
		return c.writeDubbo2(ctx)
	case http_context.IHttpContext:
		return c.writeHttp(ctx)
	case grpc_context.IGrpcContext:
		return c.writeGrpc(ctx)
	}

	return errors.New("eoContext unrealized")
}

func (c *complete) writeHttp(ctx http_context.IHttpContext) error {
	ctx.Response().SetHeader("Content-Type", c.contentType)
	ctx.Response().SetStatus(c.responseStatus, "")

	if c.responseExample != "" {
		ctx.Response().SetBody([]byte(c.responseExample))
		return nil
	}

	schema := jsonSchemaUnmarshal(c.responseSchema)
	bytes, err := json.Marshal(schema)
	if err != nil {
		log.Errorf("mocking complete err=%s", err.Error())
		return err
	}

	ctx.Response().SetBody(bytes)
	return nil
}

func (c *complete) writeDubbo2(ctx dubbo2_context.IDubbo2Context) error {
	if c.responseExample != "" {

		var val interface{}
		if err := json.Unmarshal([]byte(c.responseExample), &val); err != nil {
			ctx.Response().SetBody(Dubbo2ErrorResult(err))
			return err
		}

		ctx.Response().SetBody(getDubbo2Response(val, ctx.Proxy().Attachments()))
		return nil
	}

	schema := jsonSchemaUnmarshal(c.responseSchema)
	ctx.Response().SetBody(getDubbo2Response(schema, ctx.Proxy().Attachments()))
	return nil
}

func (c *complete) writeGrpc(ctx grpc_context.IGrpcContext) error {
	descriptor, err := c.descriptor.Descriptor().FindSymbol(fmt.Sprintf("%s.%s", ctx.Proxy().Service(), ctx.Proxy().Method()))
	if err != nil {
		return err
	}
	methodDesc := descriptor.GetFile().FindService(ctx.Proxy().Service()).FindMethodByName(ctx.Proxy().Method())

	message := dynamic.NewMessage(methodDesc.GetOutputType())

	fields := message.GetKnownFields()
	for _, field := range fields {
		switch field.GetType() {
		case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE, descriptorpb.FieldDescriptorProto_TYPE_FLOAT: //float32
			message.SetField(field, gofakeit.Float32())
		case descriptorpb.FieldDescriptorProto_TYPE_INT64, descriptorpb.FieldDescriptorProto_TYPE_SINT64, descriptorpb.FieldDescriptorProto_TYPE_SFIXED64: //int64
			message.SetField(field, gofakeit.Int64())
		case descriptorpb.FieldDescriptorProto_TYPE_INT32, descriptorpb.FieldDescriptorProto_TYPE_SINT32, descriptorpb.FieldDescriptorProto_TYPE_SFIXED32: //int32
			message.SetField(field, gofakeit.Int32())
		case descriptorpb.FieldDescriptorProto_TYPE_UINT32, descriptorpb.FieldDescriptorProto_TYPE_FIXED32: //uint32
			message.SetField(field, gofakeit.Uint32())
		case descriptorpb.FieldDescriptorProto_TYPE_UINT64, descriptorpb.FieldDescriptorProto_TYPE_FIXED64: //uint64
			message.SetField(field, gofakeit.Uint64())
		case descriptorpb.FieldDescriptorProto_TYPE_BOOL: //bool
			message.SetField(field, gofakeit.Bool())
		case descriptorpb.FieldDescriptorProto_TYPE_STRING: //string
			message.SetField(field, gofakeit.LetterN(5))
		case descriptorpb.FieldDescriptorProto_TYPE_BYTES: //bytes
			message.SetField(field, []byte(gofakeit.LetterN(5)))
		}

	}

	ctx.Response().Write(message)
	return nil
}

func jsonSchemaUnmarshal(properties interface{}) interface{} {
	propertiesMap, ok := properties.(map[string]interface{})
	if !ok {
		return "son schema 格式错误"
	}
	if val, ok := propertiesMap["example"]; ok {
		return val
	} else {
		if t, tOk := propertiesMap["type"].(string); tOk {
			switch t {
			case "string":
				return gofakeit.LetterN(10)
			case "number":
				return gofakeit.Float64()
			case "integer":
				return gofakeit.Int64()
			case "boolean":
				return gofakeit.Bool()
			case "object":
				propertiesMaps, pOk := propertiesMap["properties"].(map[string]interface{})
				if !pOk {
					return "json schema 格式错误"
				}
				resultMap := make(map[string]interface{})
				for key, vProperties := range propertiesMaps {
					resultMap[key] = jsonSchemaUnmarshal(vProperties)
				}
				return resultMap
			case "array":
				items, iOk := propertiesMap["items"].(map[string]interface{})
				if !iOk {
					return "json schema 格式错误"
				}
				resultList := make([]interface{}, 0)
				resultList = append(resultList, jsonSchemaUnmarshal(items))
				return resultList
			}
		}
		return "json schema 格式错误"
	}
}

func getDubbo2Response(obj interface{}, attachments map[string]interface{}) protocol.RPCResult {
	payload := impl.NewResponsePayload(obj, nil, attachments)
	return protocol.RPCResult{
		Attrs: payload.Attachments,
		Err:   payload.Exception,
		Rest:  payload.RspObj,
	}
}

func Dubbo2ErrorResult(err error) protocol.RPCResult {
	payload := impl.NewResponsePayload(nil, err, nil)
	return protocol.RPCResult{
		Attrs: payload.Attachments,
		Err:   payload.Exception,
		Rest:  payload.RspObj,
	}
}
