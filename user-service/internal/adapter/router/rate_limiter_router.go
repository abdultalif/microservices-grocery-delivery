package router

const (
	// Profile
	RateLimitProfileView    = 30
	RateLimitProfileUpdate  = 5
	RateLimitPasswordChange = 3
	RateLimitUploadAvatar   = 3

	// Auth
	RateLimitLogin               = 5
	RateLimitCreateAccount       = 10
	RateLimitVerifyAccount       = 5
	RateLimitForgotPassword      = 3
	RateLimitValidateForgotToken = 5
	RateLimitResetPassword       = 3
	RateLimitServiceToken        = 20

	// Oauth
	RateLimitOauthRegister         = 5
	RateLimitOauthRegisterCallback = 10
	RateLimitOauthLogin            = 5
	RateLimitOauthLoginCallback    = 10
	RateLimitOauthUnlink           = 5
	RateLimitOauthLink             = 5
	RateLimitOauthSetPassword      = 3
	RateLimitOauthLogout           = 20

	// Customer
	RateLimitCustomerViewAll        = 30
	RateLimitCustomerViewByID       = 20
	RateLimitCustomerCreate         = 10
	RateLimitCustomerUpdate         = 10
	RateLimitCustomerDelete         = 5
	RateLimitCustomerUpdateLocation = 100

	// Role
	RateLimitRoleViewAll  = 20
	RateLimitRoleViewByID = 20
	RateLimitRoleCreate   = 3
	RateLimitRoleUpdate   = 5
	RateLimitRoleDelete   = 3

	// Rate Limiter Windows (in seconds)
	RateLimitWindowOneMinute   = 60   // 1 minute
	RateLimitWindowFifteenMins = 900  // 15 minutes
	RateLimitWindowOneHour     = 3600 // 1 hour

	RateLimitMaxRequestsShort = 600  // 10 minutes
	RateLimitMaxRequestsLong  = 1800 // 30 minutes

)
