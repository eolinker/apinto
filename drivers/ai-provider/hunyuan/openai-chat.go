package hunyuan

import (
	"encoding/json"
	ai_convert "github.com/eolinker/apinto/ai-convert"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeOpenAIChat, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return ai_convert.NewOpenAIChat(provider, c.APIKey, c.BaseUrl, mt, 0, nil, errorCallback)
	})

}

func errorCallback(ctx http_service.IHttpContext, body []byte) {
	/*
		腾讯混元大模型错误码对照表

		| 错误码                           | 说明                                                         |
		|---------------------------------|--------------------------------------------------------------|
		| ActionOffline                    | 接口已下线。                                                 |
		| AuthFailure.InvalidAuthorization  | 请求头部的 Authorization 不符合腾讯云标准。                     |
		| AuthFailure.InvalidSecretId      | 密钥非法（不是云 API 密钥类型）。                             |
		| AuthFailure.MFAFailure           | MFA 错误。                                                   |
		| AuthFailure.SecretIdNotFound      | 密钥不存在。请在 控制台 检查密钥是否已被删除或者禁用，如状态正常，请检查密钥是否填写正确，注意前后不得有空格。 |
		| AuthFailure.SignatureExpire       | 签名过期。Timestamp 和服务器时间相差不得超过五分钟，请检查本地时间是否和标准时间同步。 |
		| AuthFailure.SignatureFailure      | 签名错误。签名计算错误，请对照调用方式中的签名方法文档检查签名计算过程。 |
		| AuthFailure.TokenFailure          | token 错误。                                                 |
		| AuthFailure.UnauthorizedOperation | 请求未授权。请参考 CAM 文档对鉴权的说明。                       |
		| DryRunOperation                  | DryRun 操作，代表请求将会是成功的，只是多传了 DryRun 参数。   |
		| FailedOperation                  | 操作失败。                                                   |
		| InternalError                    | 内部错误。                                                   |
		| InvalidAction                    | 接口不存在。                                                 |
		| InvalidParameter                 | 参数错误（包括参数格式、类型等错误）。                         |
		| InvalidParameterValue            | 参数取值错误。                                               |
		| InvalidRequest                   | 请求 body 的 multipart 格式错误。                             |
		| IpInBlacklist                    | IP地址在黑名单中。                                           |
		| IpNotInWhitelist                 | IP地址不在白名单中。                                         |
		| LimitExceeded                    | 超过配额限制。                                               |
		| MissingParameter                 | 缺少参数。                                                   |
		| NoSuchProduct                     | 产品不存在。                                                 |
		| NoSuchVersion                    | 接口版本不存在。                                             |
		| RequestLimitExceeded             | 请求的次数超过了频率限制。                                     |
		| RequestLimitExceeded.GlobalRegionUinLimitExceeded | 主账号超过频率限制。                         |
		| RequestLimitExceeded.IPLimitExceeded | IP限频。                                                   |
		| RequestLimitExceeded.UinLimitExceeded | 主账号限频。                                               |
		| RequestSizeLimitExceeded          | 请求包超过限制大小。                                         |
		| ResourceInUse                    | 资源被占用。                                                 |
		| ResourceInsufficient              | 资源不足。                                                   |
		| ResourceNotFound                  | 资源不存在。                                                 |
		| ResourceUnavailable               | 资源不可用。                                                 |
		| ResponseSizeLimitExceeded         | 返回包超过限制大小。                                         |
		| ServiceUnavailable                | 当前服务暂时不可用。                                         |
		| UnauthorizedOperation             | 未授权操作。                                                 |
		| UnknownParameter                  | 未知参数错误，用户多传未定义的参数会导致错误。                 |
		| UnsupportedOperation              | 操作不支持。                                                 |
		| UnsupportedProtocol               | http(s) 请求协议错误，只支持 GET 和 POST 请求。                 |
		| UnsupportedRegion                 | 接口不支持所传地域。                                         |

		业务错误码

		| 错误码                           | 说明                                                         |
		|---------------------------------|--------------------------------------------------------------|
		| FailedOperation.ConsoleServerError | 控制台服务异常。                                             |
		| FailedOperation.DownloadError     | LogoUrl 或 LogoImage 有误，水印图下载失败。                     |
		| FailedOperation.EngineRequestTimeout | 引擎层请求超时；请稍后重试。                                 |
		| FailedOperation.EngineServerError  | 引擎层内部错误；请稍后重试。                                   |
		| FailedOperation.EngineServerLimitExceeded | 引擎层请求超过限额；请稍后重试。                             |
		| FailedOperation.FreeResourcePackExhausted | 免费资源包余量已用尽，请购买资源包或开通后付费。             |
		| FailedOperation.GenerateImageFailed | 图片包含敏感内容                                             |
		| FailedOperation.ImageDecodeFailed  | 水印图解码失败                                               |
		| FailedOperation.ImageDownloadError | LogoUrl 或 LogoImage 有误，水印图下载失败。                     |
		| FailedOperation.PartnerAccountUnSupport | 合作伙伴账号不允许开通，请更换账号。                           |
		| FailedOperation.ResourcePackExhausted | 资源包余量已用尽，请购买资源包或开通后付费。                     |
		| FailedOperation.ServiceNotActivated | 服务未开通，请前往控制台申请试用。                             |
		| FailedOperation.ServiceStop        | 用户主动停服。                                               |
		| FailedOperation.ServiceStopArrears  | 欠费停服。                                                   |
		| FailedOperation.SetPayModeExceed    | 后付费设置次数超过每月限制。                                   |
		| FailedOperation.UserUnAuthError     | 用户未实名，请先进行实名认证。                                 |
		| InvalidParameter.InvalidParameter  | 参数不合法。                                                 |
		| InvalidParameterValue.Model        | 模型不存在。                                                 |
		| InvalidParameterValue.ParameterValueError | 参数字段或者值有误                                     |
		| OperationDenied.ImageIllegalDetected | 图片可能包含敏感信息，请重试                                   |
		| OperationDenied.TextIllegalDetected  | 文本包含违法违规信息，审核不通过。                             |
		| ResourceInsufficient.ChargeResourceExhaust | 计费资源已耗尽。                                         |
		| ResourceUnavailable.InArrears       | 账号已欠费。                                                 |
		| ResourceUnavailable.LowBalance       | 余额不足。                                                   |
		| ResourceUnavailable.NotExist         | 计费状态未知，请确认是否已在控制台开通服务。                     |
		| ResourceUnavailable.StopUsing        | 账号已停服。                                                 |
	*/
	var data ai_convert.Response
	err := json.Unmarshal(body, &data)
	if err != nil {
		log.Errorf("unmarshal body error: %v, body: %s", err, string(body))
		return
	}

	switch data.Error.Code {
	case "AuthFailure.InvalidAuthorization", "AuthFailure.InvalidSecretId", "AuthFailure.SecretIdNotFound", "AuthFailure.SignatureFailure", "AuthFailure.TokenFailure", "AuthFailure.UnauthorizedOperation", "AuthFailure.SignatureExpire", "AuthFailure.MFAFailure":
		ai_convert.SetAIStatusInvalid(ctx)
	case "RequestLimitExceeded", "RequestLimitExceeded.GlobalRegionUinLimitExceeded", "RequestLimitExceeded.IPLimitExceeded", "RequestLimitExceeded.UinLimitExceeded", "RequestSizeLimitExceeded", "ResponseSizeLimitExceeded", "ResourceInUse", "ResourceInsufficient", "ResourceNotFound", "ResourceUnavailable":
		ai_convert.SetAIStatusExceeded(ctx)
	case "LimitExceeded", "FailedOperation.ServiceStop", "FailedOperation.ServiceStopArrears", "FailedOperation.SetPayModeExceed", "ResourceInsufficient.ChargeResourceExhaust", "ResourceUnavailable.InArrears", "ResourceUnavailable.LowBalance", "ResourceUnavailable.NotExist", "ResourceUnavailable.StopUsing", "FailedOperation.FreeResourcePackExhausted":
		ai_convert.SetAIStatusQuotaExhausted(ctx)
	default:
		ai_convert.SetAIStatusInvalidRequest(ctx)
	}
}
