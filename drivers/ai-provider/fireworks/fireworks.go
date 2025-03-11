package fireworks

import (
	"encoding/json"
	"fmt"
	"net/url"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

func init() {
	ai_convert.RegisterConverterCreateFunc("fireworks", Create)
}

type Config struct {
	APIKey  string `json:"fireworks_api_key"`
	BaseUrl string `json:"base_url"`
}

// checkConfig validates the provided configuration.
// It ensures the required fields are set and checks the validity of the Base URL if provided.
//
// Parameters:
//   - v: An interface{} expected to be a pointer to a Config struct.
//
// Returns:
//   - *Config: The validated configuration cast to *Config.
//   - error: An error if the validation fails, or nil if it succeeds.
func checkConfig(conf *Config) error {
	// Check if the APIKey is provided. It is a required field.
	if conf.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}
	if conf.BaseUrl != "" {
		u, err := url.Parse(conf.BaseUrl)
		if err != nil {
			// Return an error if the Base URL cannot be parsed.
			return fmt.Errorf("base url is invalid")
		}
		// Ensure the parsed URL contains both a scheme and a host.
		if u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("base url is invalid")
		}
	}
	return nil
}

func Create(cfg string) (ai_convert.IConverter, error) {
	var conf Config
	err := json.Unmarshal([]byte(cfg), &conf)
	if err != nil {
		return nil, err
	}
	err = checkConfig(&conf)
	if err != nil {
		return nil, err
	}

	return ai_convert.NewOpenAIConvert(conf.APIKey, conf.BaseUrl, 0, nil, errorCallback)
}

func errorCallback(ctx http_service.IHttpContext, body []byte) {
	// 400	Bad Request	Invalid input or malformed request.	Review the request parameters and ensure they match the expected format.
	// 401	Unauthorized	Invalid API key or insufficient permissions.	Verify your API key and ensure it has the correct permissions.
	// 402	Payment Required	User’s account is not on a paid plan or has exceeded usage limits.	Check your billing status and ensure your payment method is up to date. Upgrade your plan if necessary.
	// 403	Forbidden	The model name may be incorrect, or the model does not exist. This error is also returned to avoid leaking information about model availability.	Verify the model name on the Fireworks site and ensure it exists. Double-check the spelling of the model name in your request.
	// 404	Not Found	The API endpoint is incorrect, or the resource path is invalid (e.g., a user tried accessing /v1/foobar instead of a valid endpoint).	Verify the URL path in your request and ensure you are using the correct API endpoint as per the documentation.
	// 405	Method Not Allowed	Using an unsupported HTTP method (e.g., using GET instead of POST).	Check the API documentation for the correct HTTP method to use for the request.
	// 408	Request Timeout	The request took too long to complete, possibly due to server overload or network issues.	Retry the request after a brief wait. Consider increasing the timeout value if applicable.
	// 412	Precondition Failed	This error occurs when attempting to invoke a LoRA model that failed to load. The final validation of the model happens during inference, not at upload time.	Check the body of the request for a detailed error message. Ensure the LoRA model was uploaded correctly and is compatible. Contact support if the issue persists.
	// 413	Payload Too Large	Input data exceeds the allowed size limit.	Reduce the size of the input payload (e.g., by trimming large text or image data).
	// 429	Over Quota	The user has reached the API rate limit.	Wait for the quota to reset or upgrade your plan for a higher rate limit.
	// 500	Internal Server Error	This indicates a server-side code bug and is unlikely to resolve on its own.	Contact Fireworks support immediately, as this error typically requires intervention from the engineering team.
	// 502	Bad Gateway	The server received an invalid response from an upstream server.	Wait and retry the request. If the error persists, it may indicate a server outage.
	// 503	Service Unavailable	The service is down for maintenance or experiencing issues.	Retry the request after some time. Check for any maintenance announcements.
	// 504	Gateway Timeout	The server did not receive a response in time from an upstream server.	Wait briefly and retry the request. Consider using a shorter input prompt if applicable.
	// 520	Unknown Error	An unexpected error occurred with no clear explanation.	Retry the request. If the issue persists, contact support for further assistance.
	switch ctx.Response().StatusCode() {
	case 400:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 401, 403:
		// Handle the Invalid API key error. 官方返回状态码与文档不一致，应该返回401，实际返回403
		ai_convert.SetAIStatusInvalid(ctx)
	case 402:
		// Handle the insufficient quota error.
		ai_convert.SetAIStatusQuotaExhausted(ctx)
	case 429:
		// Handle the rate limit error.
		ai_convert.SetAIStatusExceeded(ctx)
	default:
		ai_convert.SetAIStatusInvalidRequest(ctx)
	}
}
