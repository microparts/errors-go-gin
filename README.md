errors-go-gin
-------------

errors for gin

```go
package main

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
	"github.com/microparts/errors-go-gin"
	
	"github.com/pkg/errors"
)

func main() {
	ginErrors.InitValidator()
	r := gin.New()
	r.GET("/", func(c*gin.Context) {c.JSON(http.StatusOK,`{"status":"ok"}`)})
	r.GET("/err", func(c *gin.Context) { ginErrors.Response(c, errors.New("error")) })
	_ = r.Run(":8080")
}
```