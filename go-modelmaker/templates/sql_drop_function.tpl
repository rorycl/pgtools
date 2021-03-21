DROP FUNCTION {{ .Schema }}.{{ .Function }} (
	{{- range $i, $a := .Args -}}
        {{- if $i }}, {{ end }}
        {{- .Typer }}
	{{- end }});
