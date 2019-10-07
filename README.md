# ginerrors

Smart generating error code and response for gin mux based on passed error.

## Usage

```go
package main

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
	"github.com/microparts/errors-go-gin"
	
	"github.com/pkg/errors"
)

func main() {
	r := gin.New()

	r.GET("/", func(c*gin.Context) {c.JSON(http.StatusOK,`{"status":"ok"}`)})
	r.GET("/err", func(c *gin.Context) { ginerrors.Response(c, errors.New("error")) })
	_ = r.Run(":8080")
}
```

## Linter

Lint code with [golangci-lint](https://github.com/golangci/golangci-lint) and 
[custom config](https://github.com/microparts/docker-golang/blob/master/lint/.golangci.yml) for it: 

    make lint

## Testing

Test code with race checking and generation coverage profile:

    make test
