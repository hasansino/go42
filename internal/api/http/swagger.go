package http

const swaggerUIVersion = "5.11.0"

// nolint:unused
const (
	// https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.12.0/swagger-ui.css
	swaggerCDNjsdelivr = "https://cdn.jsdelivr.net/npm/swagger-ui-dist@"
	// https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.12.0/swagger-ui.css
	swaggerCDNcdnjs = "https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/"
	// https://unpkg.com/swagger-ui-dist@5.12.0/swagger-ui.css
	swaggerCDNunpkg = "https://unpkg.com/swagger-ui-dist@"
)

type swaggerTemplateData struct {
	SpecURL string
	CDN     string
	Version string
}

const swaggerTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <meta name="description" content="SwaggerUI" />
    <Title>SwaggerUI</Title>
    <link rel="stylesheet" href="{{.CDN}}{{.Version}}/swagger-ui.css" />
</head>
<body>
<div id="swagger-ui"></div>
<script src="{{.CDN}}{{.Version}}/swagger-ui-bundle.js" crossorigin></script>
<script src="{{.CDN}}{{.Version}}/swagger-ui-standalone-preset.js" crossorigin></script>
<script>
    window.onload = () => {
        window.ui = SwaggerUIBundle({
            url: '{{.SpecURL}}',
            dom_id: '#swagger-ui',
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIStandalonePreset
            ],
            plugins: [
                SwaggerUIBundle.plugins.DownloadUrl
            ],
            layout: "StandaloneLayout",
            deepLinking: true,
            filter: true,
        });
    };
</script>
</body>
</html>
`
