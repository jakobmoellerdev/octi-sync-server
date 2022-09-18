package request

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

type maxBytesReader struct {
	ctx         *gin.Context
	body        io.ReadCloser
	remaining   int64
	wasAborted  bool
	sawEOF      bool
	responseObj any
}

func (r *maxBytesReader) abortRequestEntityTooLarge() (n int, err error) {
	n, err = 0, fmt.Errorf("HTTP request too large")

	if !r.wasAborted {
		r.wasAborted = true
		ctx := r.ctx
		_ = ctx.Error(err)
		ctx.Header("connection", "close")
		ctx.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, r.responseObj)
	}
	return
}

func (r *maxBytesReader) Read(p []byte) (n int, err error) {
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
	if int64(len(p)) > toRead {
		p = p[:toRead]
	}
	n, err = r.body.Read(p)
	if err == io.EOF {
		r.sawEOF = true
	}
	if r.remaining == 0 {
		// If we had zero bytes to read remaining (but hadn't seen EOF)
		// and we get a byte here, that means we went over our limit.
		if n > 0 {
			return r.abortRequestEntityTooLarge()
		}
		return 0, err
	}
	r.remaining -= int64(n)
	if r.remaining < 0 {
		r.remaining = 0
	}
	return n, err
}

func (r *maxBytesReader) Close() error {
	return r.body.Close()
}

// BodySizeLimiter returns a middleware that limits the size of request
// When a request is over the limit, the following will happen:
// * Error will be added to the context
// * Connection: close header will be set
// * Error 413 will be sent to the client (http.StatusRequestEntityTooLarge)
// * Current context will be aborted
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
