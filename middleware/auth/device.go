package auth

import (
	"net/http"

	"go.jakob-moeller.cloud/octi-sync-server/service"

	"github.com/gin-gonic/gin"
)

// DeviceID is the cookie name for user credential in basic auth.
const DeviceID = "device"

// DeviceIDHeader holds Device Authentication.
const DeviceIDHeader = "X-Device-ID"

// DeviceAuth returns a Basic HTTP Authorization middleware. It takes as arguments a map[string]string where
// the key is the username and the value is the password, as well as the name of the Realm.
// If the realm is empty, "Authorization Required" will be used by default.
// (see http://tools.ietf.org/html/rfc2617#section-1.2)
func DeviceAuth(devices service.Devices) gin.HandlerFunc {
	return func(context *gin.Context) {
		account, found := context.Get(UserKey)
		if !found {
			// Account not found, we return 401 and abort handlers chain.
			context.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		deviceID := context.GetHeader(DeviceIDHeader)
		if deviceID == "" {
			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"msg": "this endpoint has to be called with the " + DeviceIDHeader + " Header!",
			})
			return
		}

		device, err := devices.FindByDeviceID(
			context.Request.Context(),
			account.(service.Account), deviceID)
		if err != nil {
			// Device not found for Account, we return 401 and abort handlers chain.
			context.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// The account credentials was found, set account's id to key DeviceID in this context,
		// the account's id can be read later using
		// context.MustGet(auth.DeviceID).
		context.Set(DeviceID, device)
	}
}
