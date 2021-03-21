    def {{ .Schema }}_{{ .Function }}(self, {{ range $i, $a := .Args -}}
        {{- if $i }}, {{ end }}
        {{- $a.Name }}{{ if .TypeDefaulted }}=None{{ end }}{{ end -}}
        ):
        """
        Namespace: {{ .Schema }} 
        Input parameters:
        {{- range .Args }}
            PgArgument: Name: {{ .Name }}, Type: {{ .Typer }}, Default:
            {{- if .TypeDefaulted }} {{ .Default }}{{ else }} None{{ end }}
        {{- end }}

        Returns:
            type: {{ .Returns }}
        """
        return self.callproc('{{ .Schema }}.{{ .Function }}', (
        {{- range .Args }}
            {{ .Name -}}, {{- end -}}
        ))

