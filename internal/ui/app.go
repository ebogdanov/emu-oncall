package ui

// NOTE: This section is in R&D state

import (
	"strings"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/ebogdanov/emu-oncall/internal/config"
	"github.com/rs/zerolog"

	"image/png"
	"net/http"
)

type App struct {
	cfg *config.App
	l   zerolog.Logger
}

func New(cfg *config.App, logger zerolog.Logger) *App {
	return &App{cfg: cfg, l: logger.With().Str("component", "app").Logger()}
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := []byte("Not implemented =(")

	if strings.Contains(r.RequestURI, "/qrcode") {
		a.QrCode(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp)
}

// QrCode generates code which should be recognized by OnCall App
func (a *App) QrCode(w http.ResponseWriter, r *http.Request) {
	// Set the content for the QR code
	hostName := "http://" + r.Host
	if a.cfg.Hostname != "" {
		hostName = a.cfg.Hostname
	}

	content := `{"token":"18af30". "oncall_api_url":"` + hostName + `/oncall"}`

	qrCode, _ := qr.Encode(content, qr.L, qr.Auto)
	qrCode, _ = barcode.Scale(qrCode, 256, 256)

	err := png.Encode(w, qrCode)
	if err != nil {
		a.l.Error().Err(err).Msg("failed to generate QR code")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	// Set the content type to image/png
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
}
