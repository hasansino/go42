### Conventions

#### General Guidelines

- **Readability**: Code should be easy to read and understand by humans
- **Consistency**: Follow the established coding style of the project

#### Priority Order

1. Project-specific conventions in `/CONVENTIONS.md`
2. General guidelines listed above
3. Language or framework-specific best practices
4. Industry standards and widely accepted practices relevant to language, framework, or domain
5. If none of the above apply, use your best judgment

#### /CONVENTIONS.md content

{{ if gt (len .conventions.content) 0 -}}
{{ .conventions.content }}
{{ end -}}
