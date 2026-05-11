package controller

import (
	"io"
	"log"
	"net/http"
	"net/url"
)

// ImageProxyController sirve imágenes externas desde el backend para evitar
// problemas de CORS y hotlink protection en el browser.
// El frontend pide: GET /api/proxy/image?url=https://...
// El backend descarga la imagen y la sirve directamente al cliente.
type ImageProxyController struct{}

func NewImageProxyController() *ImageProxyController {
	return &ImageProxyController{}
}

var allowedHosts = map[string]bool{
	"static.wikia.nocookie.net": true,
	"upload.wikimedia.org":      true,
	"ui-avatars.com":            true,
	"api.dicebear.com":          true,
	"cdn.myanimelist.net":       true,
	"myanimelist.net":           true,
}

// ProxyImage godoc
// @Summary      Proxy de imagen externa
// @Description  Descarga una imagen desde una URL externa y la sirve al cliente, evitando CORS y hotlink protection
// @Tags         proxy
// @Param        url  query  string  true  "URL de la imagen a proxear"
// @Produce      image/png
// @Success      200
// @Failure      400  {object}  map[string]string
// @Failure      502  {object}  map[string]string
// @Router       /api/proxy/image [get]
func (c *ImageProxyController) ProxyImage(w http.ResponseWriter, r *http.Request) {
	rawURL := r.URL.Query().Get("url")
	if rawURL == "" {
		sendError(w, http.StatusBadRequest, "url query param requerido")
		return
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || !allowedHosts[parsed.Host] {
		sendError(w, http.StatusForbidden, "dominio no permitido")
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, rawURL, nil)
	if err != nil {
		sendError(w, http.StatusBadRequest, "error construyendo request")
		return
	}
	// Simular browser normal — sin Referer para evitar hotlink protection
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; OnePieceAPI/1.0)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("proxy image error fetching %s: %v", rawURL, err)
		sendError(w, http.StatusBadGateway, "error descargando imagen")
		return
	}
	defer resp.Body.Close()

	// Propagar content-type y cache headers
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Cache-Control", "public, max-age=86400") // cache 24h
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
