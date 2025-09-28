package server

import "github.com/gin-gonic/gin"

func (app *Application) HealthHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "healthy - all systems operational",
	})
}
