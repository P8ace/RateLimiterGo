# RateLimiterGo

```Go
// Middleware function singature

type MiddleWare func(http.Handlerfunc) http.Handlerfunc{}

func Middleware1() MiddleWare{
	return func(next http.Handlerfunc) http.Handlerfunc {
		return func(w http.ResponseWriter, r *http.Request){
			// Do something in the middle ware
			next()
		}
	}
}
	
}

````
