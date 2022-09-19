package request

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type maxBytesReader struct {
	ctx         *gin.Context
	body        io.ReadCloser
	remaining   int64
	wasAborted  bool
	sawEOF      bool
	responseObj any
}

var ErrRequestTooLarge = errors.New("HTTP request too large")

func (r *maxBytesReader) abortRequestEntityTooLarge() (int, error) {
	err := ErrRequestTooLarge

	if !r.wasAborted {
		r.wasAborted = true
		ctx := r.ctx
		_ = ctx.Error(err)
		ctx.Header("connection", "close")
		ctx.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, r.responseObj)
	}
	return 0, ErrRequestTooLarge
}

func (r *maxBytesReader) Read(bytes []byte) (int, error) {
	toRead := r.remaining
	if r.remaining == 0 {
		if r.sawEOF {
			return r.abortRequestEntityTooLarge()
		}
		// The underlying io.Reader may not return (0, io.EOF)
		// at EOF if the requested size is 0, so read 1 byte
		// instead. The io.Reader docs are a bit ambiguous
		// about the return value of Read when 0 bytes are
		// requested, and {bytes,strings}.Reader gets it wrong
		// too (it returns (0, nil) even at EOF).
		toRead = 1
	}
	if int64(len(bytes)) > toRead {
		bytes = bytes[:toRead]
	}
	data, err := r.body.Read(bytes)
	if errors.Is(err, io.EOF) {
		r.sawEOF = true
	}
	if r.remaining == 0 {
		// If we had zero bytes to read remaining (but hadn't seen EOF)
		// and we get a byte here, that means we went over our limit.
		if data > 0 {
			return r.abortRequestEntityTooLarge()
		}
		return 0, err
	}
	r.remaining -= int64(data)
	if r.remaining < 0 {
		r.remaining = 0
	}
	return data, err
}

func (r *maxBytesReader) Close() error {
	return r.body.Close()
}

// BodySizeLimiter returns a middleware that limits the size of request
// When a request is over the limit, the following will happen:
// * Error will be added to the context
// * Connection: close header will be set
// * Error 413 will be sent to the client (http.StatusRequestEntityTooLarge)
// * Current context will be aborted.
func BodySizeLimiter(limit int64, respondOnTooLarge any) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Request.Body = &maxBytesReader{
			ctx:         ctx,
			body:        ctx.Request.Body,
			remaining:   limit,
			wasAborted:  false,
			sawEOF:      false,
			responseObj: respondOnTooLarge,
		}
		ctx.Next()
	}
}
