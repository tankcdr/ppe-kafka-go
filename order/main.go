package main

import "github.com/gin-gonic/gin"

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.String(200, "ok")
	})
	return r
}

func main() {
	r := setupRouter()
	r.Run(":8080")
}
