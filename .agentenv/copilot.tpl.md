# Project: {{ .project.name }}

<context>
  <language>{{ .project.language }}</language>
  {{ if gt (len .project.tags) 0 -}}
  <tags>
    {{- range .project.tags }}
    - {{ . }}
    {{- end }}
  </tags>
  {{- end }}
  {{ if gt (len .project.metadata) 0 -}}
  <metadata>
    {{- range $key, $value := .project.metadata }}
    <{{ $key }}>{{ $value }}</{{ $key }}>
    {{- end }}
  </metadata>
  {{- end }}
</context>

{{ .project.description }}

## Instructions

{{ .chunks }}

## Operational Modes

User can specify the operational mode by saying "Switch to [mode name] mode".

You are allowed to operate only in one of the following modes at any given time:

{{ .modes }}

## Workflows

User can specify the workflow by saying "Use [workflow name] workflow".

You are allowed to execute only one of the following workflows at any given time:

{{ .workflows }}

{{ if gt (len .analysis) 0 -}}
{{ .analysis.content }}
{{ end -}}
