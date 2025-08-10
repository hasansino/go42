# Project: {{.Project}}

<context>
  <project>{{.Project}}</project>
  <language>{{.Language}}</language>
  {{- if .PRNumber}}
  <pr_number>{{.PRNumber}}</pr_number>
  {{- end}}
  {{- if .CommitSHA}}
  <commit_sha>{{.CommitSHA}}</commit_sha>
  {{- end}}
  {{- if .BuildURL}}
  <build_url>{{.BuildURL}}</build_url>
  {{- end}}
</context>
