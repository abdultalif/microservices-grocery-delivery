package router

const (

	// Cart Customer Rate Limiter
	RateLimitAddToCart      = 20
	RateLimitGetCart        = 60
	RateLimitRemoveFromCart = 20
	RateLimitRemoveAllCart  = 5

	// Category Public Rate Limit
	RateLimitGetCategoriesHome = 60
	RateLimitGetCategoriesShop = 60

	// Category Admin Rate Limit
	RateLimitGetAllCategories  = 30
	RateLimitGetCategoryByID   = 30
	RateLimitGetCategoryBySlug = 30
	RateLimitCreateCategory    = 10
	RateLimitUpdateCategory    = 10
	RateLimitDeleteCategory    = 5

	// Product Public Rate Limit
	RateLimitGetProductsShop      = 60
	RateLimitGetProductsHome      = 60
	RateLimitGetProductDetailHome = 60

	// Product Admin Rate Limit
	RateLimitGetAllProducts   = 30
	RateLimitGetProductByID   = 30
	RateLimitCreateProduct    = 10
	RateLimitUpdateProduct    = 10
	RateLimitDeleteProduct    = 5
	RateLimitUploadProductImg = 10

	// Rate Limiter Windows (in seconds)
	RateLimitWindowOneMinute   = 60   // 1 minute
	RateLimitWindowFifteenMins = 900  // 15 minutes
	RateLimitWindowOneHour     = 3600 // 1 hour

	RateLimitMaxRequestsShort = 600  // 10 minutes
	RateLimitMaxRequestsLong  = 1800 // 30 minutes

)
