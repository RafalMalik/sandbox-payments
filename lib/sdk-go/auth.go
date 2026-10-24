package sdk

import "strings"

func LooksLikeJWT(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	return parts[0] != "" && parts[1] != "" && parts[2] != ""
}

func AuthHeaders(token string, includeJSONContentType bool) map[string]string {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}
	if !LooksLikeJWT(token) {
		headers["X-Api-Key"] = token
	}
	headers["Accept"] = "application/json"
	if includeJSONContentType {
		headers["Content-Type"] = "application/json"
	}
	return headers
}
