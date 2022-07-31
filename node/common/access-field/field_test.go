package access_field

//
//import (
//	"fmt"
//	"testing"
//)
//
//func TestFields_ToMap(t *testing.T) {
//	type fields struct {
//		RequestId         string
//		Msec              int64
//		TimeLocal         string
//		TimeIso8601       string
//		Timestamp         int64
//		Timing            int64
//		service           string
//		ServiceTitle      string
//		Version           string
//		Api               string
//		ApiTitle          string
//		ApiPath           string
//		RequestUri        string
//		Host              string
//		GatewayIp         string
//		Cluster           string
//		ClusterName       string
//		StatusCode        int
//		RequestMsg        string
//		RequestMsgSize    string
//		RequestHeader     string
//		ResponseMsg       string
//		ResponseMsgSize   int
//		ResponseHeader    string
//		Proxys            []*Proxy
//		RemoteAddr        string
//		HTTPXForwardedFor string
//		HTTPReferer       string
//		HTTPUserAgent     string
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		want   map[string]interface{}
//	}{
//		{
//			name:   "test",
//			fields: fields{},
//			want:   nil,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			f := &Fields{
//				RequestID:         tt.fields.RequestId,
//				Msec:              tt.fields.Msec,
//				TimeLocal:         tt.fields.TimeLocal,
//				TimeIso8601:       tt.fields.TimeIso8601,
//				Timestamp:         tt.fields.Timestamp,
//				Timing:            tt.fields.Timing,
//				service:           tt.fields.service,
//				ServiceTitle:      tt.fields.ServiceTitle,
//				Version:           tt.fields.Version,
//				Api:               tt.fields.Api,
//				ApiTitle:          tt.fields.ApiTitle,
//				ApiPath:           tt.fields.ApiPath,
//				RequestUri:        tt.fields.RequestUri,
//				Host:              tt.fields.Host,
//				GatewayIp:         tt.fields.GatewayIp,
//				Cluster:           tt.fields.Cluster,
//				ClusterName:       tt.fields.ClusterName,
//				StatusCode:        tt.fields.StatusCode,
//				RequestMsg:        tt.fields.RequestMsg,
//				RequestMsgSize:    tt.fields.RequestMsgSize,
//				RequestHeader:     tt.fields.RequestHeader,
//				ResponseMsg:       tt.fields.ResponseMsg,
//				ResponseMsgSize:   tt.fields.ResponseMsgSize,
//				ResponseHeader:    tt.fields.ResponseHeader,
//				Proxys:            tt.fields.Proxys,
//				RemoteAddr:        tt.fields.RemoteAddr,
//				HTTPXForwardedFor: tt.fields.HTTPXForwardedFor,
//				HTTPReferer:       tt.fields.HTTPReferer,
//				HTTPUserAgent:     tt.fields.HTTPUserAgent,
//			}
//			 got := f.ToMap()
//			 log.Debug(got)
//
//		})
//	}
//}
