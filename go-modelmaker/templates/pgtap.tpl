SELECT * FROM {{ .Function }}(
    {{- range $i, $a := .Args }}
    {{if $i}},{{ end}}{{ printf "%s" $a.Name }} => 
    {{- end }}
)
