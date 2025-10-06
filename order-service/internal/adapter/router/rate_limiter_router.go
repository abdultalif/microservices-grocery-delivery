package router

const (

	// Rate Limiter Limit Order Customer
	RateLimitCreateOrder       = 10
	RateLimitGetAllCustomer    = 60
	RateLimitGetDetailCustomer = 30
	RateLimitGetOrderByCode    = 30

	// Rate Limiter Limit Order Admin
	RateLimitGetAllAdmin       = 60
	RateLimitDeleteOrderByID   = 10
	RateLimitGetOrderByIDAdmin = 60
	RateLimitUpdateOrderStatus = 40

	// Rate Limiter Windows (in seconds)
	RateLimitWindowOneMinute   = 60   // 1 minute
	RateLimitWindowFifteenMins = 900  // 15 minutes
	RateLimitWindowOneHour     = 3600 // 1 hour

	RateLimitMaxRequestsShort = 600  // 10 minutes
	RateLimitMaxRequestsLong  = 1800 // 30 minutes

)
