package ai

import "strings"

// FinalizeImageResponse ensures the client can display the image.
// Mini App frontend reads imageUrl / image_url (not image_base64).
func FinalizeImageResponse(r ImageResponse) ImageResponse {
	b64 := strings.TrimSpace(r.ImageBase64)
	if strings.TrimSpace(r.ImageURL) == "" && b64 != "" {
		r.ImageURL = base64DataURL(b64)
	}
	return r
}

func base64DataURL(b64 string) string {
	if strings.HasPrefix(b64, "data:") {
		return b64
	}
	mime := "image/png"
	if strings.HasPrefix(b64, "/9j/") {
		mime = "image/jpeg"
	}
	return "data:" + mime + ";base64," + b64
}
