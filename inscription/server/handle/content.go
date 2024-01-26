package handle

import (
	"bytes"
	"github.com/andybalholm/brotli"
	"github.com/gin-gonic/gin"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/internal/util"
	"github.com/kataras/iris/v12/context"
	"net/http"
)

// Content is a handler function for handling content requests.
// It validates the request parameters and calls the doContent function.
func (h *Handler) Content(ctx *gin.Context) {
	inscriptionId := ctx.Param("inscriptionId")
	if inscriptionId == "" {
		ctx.Status(http.StatusBadRequest)
		return
	}
	if err := h.doContent(ctx, inscriptionId); err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
}

// doContent is a helper function for handling content requests.
// It retrieves the content of a specific inscription and returns it in the response.
func (h *Handler) doContent(ctx *gin.Context, inscriptionId string) error {
	inscription, err := h.DB().GetInscriptionById(inscriptionId)
	if err != nil {
		return err
	}
	if inscription.Id == 0 {
		ctx.Status(http.StatusNotFound)
		return nil
	}

	// Set cache control headers
	ctx.Header(context.CacheControlHeaderKey, "public, max-age=1209600, immutable")

	// Set content type
	contentType := constants.ContentTypeOctetStream
	if inscription.ContentType != "" {
		contentType = constants.ContentType(inscription.ContentType)
	}

	// Handle content encoding
	if inscription.ContentEncoding != "" {
		acceptEncoding := util.ParseAcceptEncoding(ctx.Request.Header.Get(context.AcceptEncodingHeaderKey))
		if acceptEncoding.IsAccept(inscription.ContentEncoding) {
			ctx.Header(context.ContentEncodingHeaderKey, inscription.ContentEncoding)
		} else if inscription.ContentEncoding == "br" {
			if len(inscription.Body) == 0 {
				ctx.Status(http.StatusOK)
				return nil
			}
			decompressed := make([]byte, 0)
			if _, err := brotli.NewReader(bytes.NewReader(inscription.Body)).Read(decompressed); err != nil {
				return err
			}
			ctx.Data(http.StatusOK, string(contentType), decompressed)
		} else {
			ctx.Status(http.StatusNotAcceptable)
			return nil
		}
	}

	// If there is nobody, return an OK status
	if len(inscription.Body) == 0 {
		ctx.Status(http.StatusOK)
		return nil
	}

	// Return the body with the appropriate content type
	ctx.Data(http.StatusOK, string(contentType), inscription.Body)
	return nil
}
