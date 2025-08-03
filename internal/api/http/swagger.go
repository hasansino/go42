package http

type swaggerTemplateData struct {
	SpecURLs  map[string]string
	DarkTheme bool
}

const swaggerTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <meta name="description" content="SwaggerUI" />
    <Title>SwaggerUI</Title>
	<link rel="icon" href="/static/favicon.svg" type="image/svg+xml" sizes="any">
    <link rel="stylesheet" href="/static/swagger/swagger-ui.css" />
	{{ if .DarkTheme }}
		<link rel="stylesheet" href="/static/swagger/dark.min.css" />
		<link rel="stylesheet" href="/static/swagger/one-dark.min.css" />
	{{ end }}
</head>
<body>
<div id="swagger-ui"></div>
<script src="/static/swagger/swagger-ui-bundle.js" crossorigin></script>
<script src="/static/swagger/swagger-ui-standalone-preset.js" crossorigin></script>
<script>
    window.onload = () => {
        window.ui = SwaggerUIBundle({
			urls: [
                  {{range $name, $url := .SpecURLs}}
                  {
                      url: "{{$url}}",
                      name: "{{$name}}"
                  },
                  {{end}}
            ],
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
			displayRequestDuration: true,
			syntaxHighlight: {
				activated: true,
				theme: "obsidian"
			},
			withCredentials: true,
			persistAuthorization: true,
			defaultModelsExpandDepth: -1
        });
    };
</script>
</body>
</html>
`
