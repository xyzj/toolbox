package ginmiddleware

import (
	_ "embed"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed favicon.webp
var favicon []byte

//go:embed pages/404-cat.html
var page404cat []byte

//go:embed pages/404-code.html
var page404code []byte

//go:embed pages/404.html
var page404 []byte

//go:embed pages/500.html
var page500 []byte

//go:embed pages/403.html
var page403 []byte

//go:embed pages/helloworld.html
var pageHelloworld []byte

var templateEmpty = []byte(`<p><span style="color:hsl(0,0%,100%);"><strong>If you don't know what you're doing, just walk away...</strong></span></p>`)

// PageEmpty PPageEmptyage403
func PageEmpty(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write(templateEmpty)
}

// PageAbort PPageEmptyage403
func PageAbort(c *gin.Context) {
	c.AbortWithStatus(http.StatusGone)
}

// Page403 Page403
func Page403(c *gin.Context) {
	if c.Request.Method == "GET" {
		c.Header("Content-Type", "text/html")
		c.Writer.WriteHeader(http.StatusForbidden)
		c.Writer.Write(page403)
		return
	}
	c.String(http.StatusForbidden, "403 Forbidden")
}

// Page404 Page404
func Page404(c *gin.Context) {
	if c.Request.Method == "GET" {
		c.Header("Content-Type", "text/html")
		c.Writer.WriteHeader(http.StatusNotFound)
		c.Writer.Write(page404)
		return
	}
	c.String(http.StatusNotFound, "404 nothing here")
}

func Page404Rand(c *gin.Context) {
	s := pic404[rand.Intn(len(pic404))]
	if c.Request.Method == "GET" {
		c.Header("Content-Type", "text/html")
		c.Writer.WriteHeader(http.StatusNotFound)
		c.Writer.Write(s)
		return
	}
	c.String(http.StatusNotFound, "404 nothing here")
}

// Page404Code Page404
func Page404Code(c *gin.Context) {
	if c.Request.Method == "GET" {
		c.Header("Content-Type", "text/html")
		c.Writer.WriteHeader(http.StatusNotFound)
		c.Writer.Write(page404code)
		return
	}
	c.String(http.StatusNotFound, "404 nothing here")
}

// Page405 Page405
func Page405(c *gin.Context) {
	c.String(http.StatusMethodNotAllowed, "405 "+c.Request.Method+" is not allowed")
}

// PageDev PageDev
func PageDev(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.Writer.WriteHeader(http.StatusServiceUnavailable)
	c.Writer.Write(page500)
}

// PageDefault 健康检查
func PageDefault(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		if c.Request.RequestURI == "/" {
			c.Header("Content-Type", "text/html")
			c.Writer.WriteHeader(http.StatusOK)
			c.Writer.Write(pageHelloworld)
		} else {
			c.String(http.StatusOK, "ok")
		}
	case "POST":
		c.String(http.StatusOK, "ok")
	}
}
