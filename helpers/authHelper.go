package helpers

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 11)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}
func VerifyPassword(userPassword string, providedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	if err == nil {
		return true
	}
	return false
}
func CheckUserType(c *gin.Context, role string) (err error) {
	userType := c.GetString("role")
	err = nil
	if userType != role {
		err = errors.New("Unauthorized access")
		return err
	}
	return err
}

func MatchUserTypeToUid(c *gin.Context, userId string) (err error) {
	userType := c.GetString("role")
	uid := c.GetString("uid")
	err = nil
	if userType == "USER" && uid != userId {
		err = errors.New("Unauthorized access")
		return err
	}
	err = CheckUserType(c, userType)
	return err
}
