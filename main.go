package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	rtctokenbuilder "github.com/AgoraIO-Community/go-tokenbuilder/rtctokenbuilder"
	rtmtokenbuilder "github.com/AgoraIO-Community/go-tokenbuilder/rtmtokenbuilder"
	"github.com/gin-gonic/gin"
)

const RoleRtmUser = "RTM_USER_ROLE"

var appID = "97e7d0cfd7694624933991bb2fe241bd"          
var appCertificate = "1d3f90ccf18c419291ee6446842b2ba4" 

func main() {
	api := gin.Default()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	api.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	api.Use(nocache())
	api.GET("rtc/:channelName/:role/:tokentype/:uid/", getRtcToken)
	api.GET("rtm/:uid/", getRtmToken)
	api.GET("rte/:channelName/:role/:tokentype/:uid/", getBothTokens)
	api.Run(":" + port) 
}

func nocache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "private, no-cache, no-store, must-revalidate")
		c.Header("Expires", "-1")
		c.Header("Pragma", "no-cache")
		c.Header("Access-Control-Allow-Origin", "*")
	}
}

// Rest of your code remains the same...

func getRtcToken(c *gin.Context) {
	log.Printf("rtc token\n")
	channelName, tokentype, uidStr, role, expireTimestamp, err := parseRtcParams(c)

	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(400, gin.H{
			"message": "Error Generating RTC token: " + err.Error(),
			"status":  400,
		})
		return
	}

	rtcToken, tokenErr := generateRtcToken(channelName, uidStr, tokentype, role, expireTimestamp)

	if tokenErr != nil {
		log.Println(tokenErr)
		c.Error(tokenErr)
		errMsg := "Error Generating RTC token - " + tokenErr.Error()
		c.AbortWithStatusJSON(400, gin.H{
			"status": 400,
			"error":  errMsg,
		})
	} else {
		log.Println("RTC Token generated")
		c.JSON(200, gin.H{
			"rtcToken": rtcToken,
		})
	}
}

func getRtmToken(c *gin.Context) {
	log.Printf("rtm token\n")
	uidStr, expireTimestamp, err := parseRtmParams(c)

	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(400, gin.H{
			"message": "Error Generating RTM token: " + err.Error(),
			"status":  400,
		})
		return
	}

	rtmToken, tokenErr := rtmtokenbuilder.BuildToken(appID, appCertificate, uidStr, expireTimestamp, "Rtm_User")

	if tokenErr != nil {
		log.Println(tokenErr)
		c.Error(tokenErr)
		errMsg := "Error Generating RTM token: " + tokenErr.Error()
		c.AbortWithStatusJSON(400, gin.H{
			"error":  errMsg,
			"status": 400,
		})
	} else {
		log.Println("RTM Token generated")
		c.JSON(200, gin.H{
			"rtmToken": rtmToken,
		})
	}
}

func getBothTokens(c *gin.Context) {
	log.Printf("dual token\n")
	channelName, tokentype, uidStr, role, expireTimestamp, rtcParamErr := parseRtcParams(c)

	if rtcParamErr != nil {
		c.Error(rtcParamErr)
		c.AbortWithStatusJSON(400, gin.H{
			"message": "Error Generating RTC token: " + rtcParamErr.Error(),
			"status":  400,
		})
		return
	}

	rtcToken, rtcTokenErr := generateRtcToken(channelName, uidStr, tokentype, role, expireTimestamp)
	rtmToken, rtmTokenErr := rtmtokenbuilder.BuildToken(appID, appCertificate, uidStr, expireTimestamp, RoleRtmUser)

	if rtcTokenErr != nil {
		log.Println(rtcTokenErr)
		c.Error(rtcTokenErr)
		errMsg := "Error Generating RTC token - " + rtcTokenErr.Error()
		c.AbortWithStatusJSON(400, gin.H{
			"status": 400,
			"error":  errMsg,
		})
	} else if rtmTokenErr != nil {
		log.Println(rtmTokenErr)
		c.Error(rtmTokenErr)
		errMsg := "Error Generating RTM token - " + rtmTokenErr.Error()
		c.AbortWithStatusJSON(400, gin.H{
			"status": 400,
			"error":  errMsg,
		})
	} else {
		log.Println("Tokens generated")
		c.JSON(200, gin.H{
			"rtcToken": rtcToken,
			"rtmToken": rtmToken,
		})
	}
}

func parseRtcParams(c *gin.Context) (channelName, tokentype, uidStr string, role rtctokenbuilder.Role, expireTimestamp uint32, err error) {
	channelName = c.Param("channelName")
	roleStr := c.Param("role")
	tokentype = c.Param("tokentype")
	uidStr = c.Param("uid")
	expireTime := c.DefaultQuery("expiry", "3600")

	if roleStr == "publisher" {
		role = rtctokenbuilder.RolePublisher
	} else {
		role = rtctokenbuilder.RoleSubscriber
	}

	expireTime64, parseErr := strconv.ParseUint(expireTime, 10, 64)
	if parseErr != nil {
		err = fmt.Errorf("failed to parse expireTime: %s, causing error: %s", expireTime, parseErr)
	}

	expireTimeInSeconds := uint32(expireTime64)
	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp = currentTimestamp + expireTimeInSeconds

	return channelName, tokentype, uidStr, role, expireTimestamp, err
}

func parseRtmParams(c *gin.Context) (uidStr string, expireTimestamp uint32, err error) {
	uidStr = c.Param("uid")
	expireTime := c.DefaultQuery("expiry", "3600")

	expireTime64, parseErr := strconv.ParseUint(expireTime, 10, 64)
	if parseErr != nil {
		err = fmt.Errorf("failed to parse expireTime: %s, causing error: %s", expireTime, parseErr)
	}

	expireTimeInSeconds := uint32(expireTime64)
	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp = currentTimestamp + expireTimeInSeconds

	return uidStr, expireTimestamp, err
}

func generateRtcToken(channelName, uidStr, tokentype string, role rtctokenbuilder.Role, expireTimestamp uint32) (rtcToken string, err error) {
	log.Printf(appID, appCertificate)
	if tokentype == "userAccount" {
		log.Printf("Building Token with userAccount: %s\n", uidStr)
		rtcToken, err = rtctokenbuilder.BuildTokenWithAccount(appID, appCertificate, channelName, uidStr, role, expireTimestamp)

		return rtcToken, err
	} else if tokentype == "uid" {
		uid64, parseErr := strconv.ParseUint(uidStr, 10, 64)
		if parseErr != nil {
			err = fmt.Errorf("failed to parse uidStr: %s, to uint causing error: %s", uidStr, parseErr)
			return "", err
		}

		uid := uint32(uid64)
		log.Printf("Building Token with uid: %d\n", uid)
		rtcToken, err = rtctokenbuilder.BuildTokenWithUid(appID, appCertificate, channelName, uid, role, expireTimestamp)

		return rtcToken, err
	} else {
		err = fmt.Errorf("failed to generate RTC token for Unknown Tokentype: %s", tokentype)
		log.Println(err)
		return "", err
	}
}
