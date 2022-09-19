package routes

import (
	controllers "v8hid/Go-Mongo-Auth/controllers"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("users/signup", controllers.Singup())
	incomingRoutes.POST("users/login", controllers.Login())
}
