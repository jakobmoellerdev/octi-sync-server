package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unsafe"

	"go.jakob-moeller.cloud/octi-sync-server/service"

	"github.com/gin-gonic/gin"
)

// UserKey is the cookie name for user credential in basic auth.
const UserKey = "user"

// HeaderPrefix gets appended for every Auth Header.
const HeaderPrefix = "Basic "

// BasicAuthForRealm returns a Basic HTTP Authorization middleware. It takes as arguments a map[string]string where
// the key is the username and the value is the password, as well as the name of the Realm.
// If the realm is empty, "Authorization Required" will be used by default.
// (see http://tools.ietf.org/html/rfc2617#section-1.2)
func BasicAuthForRealm(accounts service.Accounts, realm string) gin.HandlerFunc {
	if realm == "" {
		realm = "Authorization Required"
	}
	realm = "Basic realm=" + strconv.Quote(realm)
	return func(context *gin.Context) {
		// Search user in the slice of allowed credentials
		user, err := accounts.FindHashed(context.Request.Context(),
			getHashedPassFromHeader(context.Request.Header.Get("Authorization")))
		if err != nil {
			// Credentials doesn't match, we return 401 and abort handlers chain.
			context.Header("WWW-Authenticate", realm)
			context.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// The user credentials was found, set user's id to key DeviceID in this context,
		// the user's id can be read later using
		// context.MustGet(auth.DeviceID).
		context.Set(UserKey, user)
	}
}

// BasicAuth returns a Basic HTTP Authorization middleware. It takes as argument a map[string]string where
// the key is the username and the value is the password.
func BasicAuth(accounts service.Accounts) gin.HandlerFunc {
	return BasicAuthForRealm(accounts, "")
}

func getHashedPassFromHeader(header string) string {
	encodedPass := strings.TrimLeft(header, HeaderPrefix)
	userPassBytes, _ := base64.StdEncoding.DecodeString(encodedPass)
	userPassString := bytesToString(userPassBytes)
	userPassArray := strings.Split(userPassString, ":")
	pass := userPassArray[1]
	hashedPass := fmt.Sprintf("%x", sha256.Sum256([]byte(pass)))
	return hashedPass
}

// bytesToString converts byte slice to string without a memory allocation.
func bytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
