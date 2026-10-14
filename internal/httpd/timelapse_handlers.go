package httpd

import (
	"bytes"
	"image"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	mjpeg "github.com/mattn/go-mjpeg"
	"github.com/rs/zerolog/log"
)


type timelapseResponse struct {
	Name        string `json:"name" binding:"required"`
}

func listTimelapseHandler(rootDir string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		dirs, err := os.ReadDir(rootDir)
		if err != nil {
			ctx.AbortWithError(500, err)
		}

		respItems := make([]timelapseResponse, len(dirs))
		for i, v := range dirs {
			respItems[i] = timelapseResponse{
				Name: v.Name(),
			}
		}
		ctx.JSON(http.StatusOK, respItems)
	}
}

func getTimelapse(rootDir string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		mac := ctx.Param("mac")
		log.Info().Str("mac", mac).Msg("making timelapse")

		macDir := filepath.Join(rootDir, mac)

		vidFiles, err := os.ReadDir(macDir)
		if err != nil {
			log.Error().Str("macDir", macDir).Err(err).Msg("failed to read timelapse dir")
			ctx.AbortWithError(500, err)
			return
		}


		stream := mjpeg.NewStream()
		go stream.ServeHTTP(ctx.Writer, ctx.Request)

		for _, v := range vidFiles {
			log.Info().Str("name", v.Name()).Msg("checking file")
			if strings.HasSuffix(v.Name(), ".jpeg") {
				log.Info().Str("name", v.Name()).Msg("adding file")
				contents, err := os.ReadFile(filepath.Join(macDir, v.Name()))
				if err != nil {
					log.Error().Str("macDir", macDir).Str("file", v.Name()).Err(err).Msg("failed to read file")
					continue
				}

				imgBuf := bytes.NewBuffer(contents)
				parsedImg, _, err := image.Decode(imgBuf)
				if err != nil {
					log.Error().Str("image", v.Name()).Err(err).Msg("failed to parse image")
					continue
				}

				newImg := imaging.FlipV(parsedImg)

				outBuf := bytes.NewBuffer([]byte{})

				err = imaging.Encode(outBuf, newImg, imaging.JPEG)
				if err != nil {
					log.Error().Str("img", v.Name()).Err(err).Msg("failed to re-encode flipped image")
					continue
				}

				err = stream.Update(outBuf.Bytes())
				if err != nil {
					log.Error().Str("macDir", macDir).Str("file", v.Name()).Err(err).Msg("failed to update stream")
					continue
				}
				time.Sleep(time.Millisecond*200)
			}
		}

		err = stream.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close stream")
		}
	}
}
