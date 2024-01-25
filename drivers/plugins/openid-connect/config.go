package openid_connect

type Config struct {
	Anonymous                    interface{}   `json:"anonymous"`
	Audience                     []interface{} `json:"audience"`
	AudienceClaim                []interface{} `json:"audience_claim"`
	AudienceRequired             []interface{} `json:"audience_required"`
	AuthMethods                  []string      `json:"auth_methods"`
	AuthenticatedGroupsClaim     []string      `json:"authenticated_groups_claim"`
	AuthorizationCookieDomain    interface{}   `json:"authorization_cookie_domain"`
	AuthorizationCookieHttponly  bool          `json:"authorization_cookie_httponly"`
	AuthorizationCookieLifetime  int           `json:"authorization_cookie_lifetime"`
	AuthorizationCookieName      string        `json:"authorization_cookie_name"`
	AuthorizationCookiePath      string        `json:"authorization_cookie_path"`
	AuthorizationCookieSamesite  string        `json:"authorization_cookie_samesite"`
	AuthorizationCookieSecure    interface{}   `json:"authorization_cookie_secure"`
	AuthorizationEndpoint        interface{}   `json:"authorization_endpoint"`
	AuthorizationQueryArgsClient interface{}   `json:"authorization_query_args_client"`
	AuthorizationQueryArgsNames  interface{}   `json:"authorization_query_args_names"`
	AuthorizationQueryArgsValues interface{}   `json:"authorization_query_args_values"`
	BearerTokenCookieName        interface{}   `json:"bearer_token_cookie_name"`
	BearerTokenParamType         []string      `json:"bearer_token_param_type"`
	CacheIntrospection           bool          `json:"cache_introspection"`
	CacheTokenExchange           bool          `json:"cache_token_exchange"`
	CacheTokens                  bool          `json:"cache_tokens"`
	CacheTokensSalt              interface{}   `json:"cache_tokens_salt"`
	CacheTtl                     int           `json:"cache_ttl"`
	CacheTtlMax                  interface{}   `json:"cache_ttl_max"`
	CacheTtlMin                  interface{}   `json:"cache_ttl_min"`
	CacheTtlNeg                  interface{}   `json:"cache_ttl_neg"`
	CacheTtlResurrect            interface{}   `json:"cache_ttl_resurrect"`
	CacheUserInfo                bool          `json:"cache_user_info"`
	ClientAlg                    interface{}   `json:"client_alg"`
	ClientArg                    string        `json:"client_arg"`
	ClientAuth                   interface{}   `json:"client_auth"`
	ClientCredentialsParamType   []string      `json:"client_credentials_param_type"`
	ClientId                     []string      `json:"client_id"`
	ClientJwk                    interface{}   `json:"client_jwk"`
	ClientSecret                 []string      `json:"client_secret"`
	ConsumerBy                   []string      `json:"consumer_by"`
	ConsumerClaim                interface{}   `json:"consumer_claim"`
	ConsumerOptional             bool          `json:"consumer_optional"`
	CredentialClaim              []string      `json:"credential_claim"`
	DiscoveryHeadersNames        interface{}   `json:"discovery_headers_names"`
	DiscoveryHeadersValues       interface{}   `json:"discovery_headers_values"`
	DisplayErrors                bool          `json:"display_errors"`
	Domains                      interface{}   `json:"domains"`

	EnableHsSignatures      bool          `json:"enable_hs_signatures"`
	EndSessionEndpoint      interface{}   `json:"end_session_endpoint"`
	ExtraJwksUris           interface{}   `json:"extra_jwks_uris"`
	ForbiddenDestroySession bool          `json:"forbidden_destroy_session"`
	ForbiddenErrorMessage   string        `json:"forbidden_error_message"`
	ForbiddenRedirectUri    interface{}   `json:"forbidden_redirect_uri"`
	GroupsClaim             []string      `json:"groups_claim"`
	GroupsRequired          interface{}   `json:"groups_required"`
	HideCredentials         bool          `json:"hide_credentials"`
	HttpProxy               interface{}   `json:"http_proxy"`
	HttpProxyAuthorization  interface{}   `json:"http_proxy_authorization"`
	HttpVersion             float64       `json:"http_version"`
	HttpsProxy              interface{}   `json:"https_proxy"`
	HttpsProxyAuthorization interface{}   `json:"https_proxy_authorization"`
	IdTokenParamName        interface{}   `json:"id_token_param_name"`
	IdTokenParamType        []string      `json:"id_token_param_type"`
	IgnoreSignature         []interface{} `json:"ignore_signature"`
	IntrospectJwtTokens     bool          `json:"introspect_jwt_tokens"`

	Issuer                       string        `json:"issuer"`
	IssuersAllowed               interface{}   `json:"issuers_allowed"`
	JwtSessionClaim              string        `json:"jwt_session_claim"`
	JwtSessionCookie             interface{}   `json:"jwt_session_cookie"`
	Keepalive                    bool          `json:"keepalive"`
	Leeway                       int           `json:"leeway"`
	LoginAction                  string        `json:"login_action"`
	LoginMethods                 []string      `json:"login_methods"`
	LoginRedirectMode            string        `json:"login_redirect_mode"`
	LoginRedirectUri             interface{}   `json:"login_redirect_uri"`
	LoginTokens                  []string      `json:"login_tokens"`
	LogoutMethods                []string      `json:"logout_methods"`
	LogoutPostArg                interface{}   `json:"logout_post_arg"`
	LogoutQueryArg               interface{}   `json:"logout_query_arg"`
	LogoutRedirectUri            interface{}   `json:"logout_redirect_uri"`
	LogoutRevoke                 bool          `json:"logout_revoke"`
	LogoutRevokeAccessToken      bool          `json:"logout_revoke_access_token"`
	LogoutRevokeRefreshToken     bool          `json:"logout_revoke_refresh_token"`
	LogoutUriSuffix              interface{}   `json:"logout_uri_suffix"`
	MaxAge                       interface{}   `json:"max_age"`
	NoProxy                      interface{}   `json:"no_proxy"`
	PasswordParamType            []string      `json:"password_param_type"`
	PreserveQueryArgs            bool          `json:"preserve_query_args"`
	RedirectUri                  interface{}   `json:"redirect_uri"`
	RediscoveryLifetime          int           `json:"rediscovery_lifetime"`
	RefreshTokenParamName        interface{}   `json:"refresh_token_param_name"`
	RefreshTokenParamType        []string      `json:"refresh_token_param_type"`
	RefreshTokens                bool          `json:"refresh_tokens"`
	ResponseMode                 string        `json:"response_mode"`
	ResponseType                 []string      `json:"response_type"`
	Reverify                     bool          `json:"reverify"`
	RevocationEndpoint           interface{}   `json:"revocation_endpoint"`
	RevocationEndpointAuthMethod interface{}   `json:"revocation_endpoint_auth_method"`
	RolesClaim                   []string      `json:"roles_claim"`
	RolesRequired                interface{}   `json:"roles_required"`
	RunOnPreflight               bool          `json:"run_on_preflight"`
	Scopes                       []string      `json:"scopes"`
	ScopesClaim                  []interface{} `json:"scopes_claim"`
	ScopesRequired               []interface{} `json:"scopes_required"`
	SearchUserInfo               bool          `json:"search_user_info"`
	SessionCompressor            string        `json:"session_compressor"`
	SessionCookieDomain          interface{}   `json:"session_cookie_domain"`

	SessionMemcacheHost   string      `json:"session_memcache_host"`
	SessionMemcachePort   int         `json:"session_memcache_port"`
	SessionMemcachePrefix string      `json:"session_memcache_prefix"`
	SessionMemcacheSocket interface{} `json:"session_memcache_socket"`

	SessionSecret                interface{}  `json:"session_secret"`
	SessionStorage               string       `json:"session_storage"`
	SessionStrategy              string       `json:"session_strategy"`
	SslVerify                    bool         `json:"ssl_verify"`
	Timeout                      int          `json:"timeout"`
	TokenEndpoint                interface{}  `json:"token_endpoint"`
	TokenEndpointAuthMethod      interface{}  `json:"token_endpoint_auth_method"`
	TokenExchangeEndpoint        interface{}  `json:"token_exchange_endpoint"`
	TokenHeadersClient           interface{}  `json:"token_headers_client"`
	TokenHeadersGrants           interface{}  `json:"token_headers_grants"`
	TokenHeadersNames            interface{}  `json:"token_headers_names"`
	TokenHeadersPrefix           interface{}  `json:"token_headers_prefix"`
	TokenHeadersReplay           interface{}  `json:"token_headers_replay"`
	TokenHeadersValues           interface{}  `json:"token_headers_values"`
	TokenPostArgsClient          interface{}  `json:"token_post_args_client"`
	TokenPostArgsNames           interface{}  `json:"token_post_args_names"`
	TokenPostArgsValues          interface{}  `json:"token_post_args_values"`
	UnauthorizedErrorMessage     string       `json:"unauthorized_error_message"`
	UnauthorizedRedirectUri      interface{}  `json:"unauthorized_redirect_uri"`
	UnexpectedRedirectUri        interface{}  `json:"unexpected_redirect_uri"`
	UpstreamAccessTokenHeader    string       `json:"upstream_access_token_header"`
	UpstreamAccessTokenJwkHeader interface{}  `json:"upstream_access_token_jwk_header"`
	UpstreamHeadersClaims        interface{}  `json:"upstream_headers_claims"`
	UpstreamHeadersNames         interface{}  `json:"upstream_headers_names"`
	UpstreamIdTokenHeader        interface{}  `json:"upstream_id_token_header"`
	UpstreamIdTokenJwkHeader     interface{}  `json:"upstream_id_token_jwk_header"`
	UpstreamIntrospectionHeader  interface{}  `json:"upstream_introspection_header"`
	UpstreamRefreshTokenHeader   interface{}  `json:"upstream_refresh_token_header"`
	UpstreamSessionIdHeader      interface{}  `json:"upstream_session_id_header"`
	UpstreamUserInfoHeader       interface{}  `json:"upstream_user_info_header"`
	UserinfoEndpoint             interface{}  `json:"userinfo_endpoint"`
	VerifyClaims                 bool         `json:"verify_claims"`
	VerifyNonce                  bool         `json:"verify_nonce"`
	VerifyParameters             bool         `json:"verify_parameters"`
	VerifySignature              bool         `json:"verify_signature"`
	SessionRedis                 SessionRedis `json:"session_redis"`
}

type SessionRedis struct {
	SessionRedisAuth                   interface{} `json:"session_redis_auth"`
	SessionRedisClusterMaxredirections interface{} `json:"session_redis_cluster_maxredirections"`
	SessionRedisClusterNodes           interface{} `json:"session_redis_cluster_nodes"`
	SessionRedisConnectTimeout         interface{} `json:"session_redis_connect_timeout"`
	SessionRedisHost                   string      `json:"session_redis_host"`
	SessionRedisPort                   int         `json:"session_redis_port"`
	SessionRedisPrefix                 string      `json:"session_redis_prefix"`
	SessionRedisReadTimeout            interface{} `json:"session_redis_read_timeout"`
	SessionRedisSendTimeout            interface{} `json:"session_redis_send_timeout"`
	SessionRedisServerName             interface{} `json:"session_redis_server_name"`
	SessionRedisSocket                 interface{} `json:"session_redis_socket"`
	SessionRedisSsl                    bool        `json:"session_redis_ssl"`
	SessionRedisSslVerify              bool        `json:"session_redis_ssl_verify"`
}

type SessionCookie struct {
	SessionCookieHttponly bool        `json:"session_cookie_httponly"`
	SessionCookieIdletime interface{} `json:"session_cookie_idletime"`
	SessionCookieLifetime int         `json:"session_cookie_lifetime"`
	SessionCookieMaxsize  int         `json:"session_cookie_maxsize"`
	SessionCookieName     string      `json:"session_cookie_name"`
	SessionCookiePath     string      `json:"session_cookie_path"`
	SessionCookieRenew    int         `json:"session_cookie_renew"`
	SessionCookieSamesite string      `json:"session_cookie_samesite"`
	SessionCookieSecure   interface{} `json:"session_cookie_secure"`
}

type Introspection struct {
	IntrospectionEndpoint           interface{} `json:"introspection_endpoint"`
	IntrospectionEndpointAuthMethod interface{} `json:"introspection_endpoint_auth_method"`
	IntrospectionHeadersClient      interface{} `json:"introspection_headers_client"`
	IntrospectionHeadersNames       interface{} `json:"introspection_headers_names"`
	IntrospectionHeadersValues      interface{} `json:"introspection_headers_values"`
	IntrospectionHint               string      `json:"introspection_hint"`
	IntrospectionPostArgsClient     interface{} `json:"introspection_post_args_client"`
	IntrospectionPostArgsNames      interface{} `json:"introspection_post_args_names"`
	IntrospectionPostArgsValues     interface{} `json:"introspection_post_args_values"`
}

type DownStream struct {
}

type Upstream struct {
	DownstreamAccessTokenHeader    interface{} `json:"downstream_access_token_header"`
	DownstreamAccessTokenJwkHeader interface{} `json:"downstream_access_token_jwk_header"`
	DownstreamHeadersClaims        interface{} `json:"downstream_headers_claims"`
	DownstreamHeadersNames         interface{} `json:"downstream_headers_names"`
	DownstreamIdTokenHeader        interface{} `json:"downstream_id_token_header"`
	DownstreamIdTokenJwkHeader     interface{} `json:"downstream_id_token_jwk_header"`
	DownstreamIntrospectionHeader  interface{} `json:"downstream_introspection_header"`
	DownstreamRefreshTokenHeader   interface{} `json:"downstream_refresh_token_header"`
	DownstreamSessionIdHeader      interface{} `json:"downstream_session_id_header"`
	DownstreamUserInfoHeader       interface{} `json:"downstream_user_info_header"`
}
