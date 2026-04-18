package httpapi

import (
	"net/http"
	"os"
)

const swaggerUIHTML = `<!DOCTYPE html>
<html>
<head>
  <title>Keklik API</title>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css">
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
<script>
  SwaggerUIBundle({
    url: "/swagger/swagger.yaml",
    dom_id: '#swagger-ui',
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
    layout: "BaseLayout"
  })
</script>
</body>
</html>`

func swaggerUIHandler(w http.ResponseWriter, r *http.Request, enabled bool) {
	if !enabled {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(swaggerUIHTML))
}

func swaggerYAMLHandler(w http.ResponseWriter, r *http.Request, enabled bool) {
	if !enabled {
		http.NotFound(w, r)
		return
	}
	data, err := os.ReadFile("docs/swagger.yaml")
	if err != nil {
		http.Error(w, "spec not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/yaml")
	_, _ = w.Write(data)
}
