    def {{ .Function }}(self, paramdict):
        '''
        SQL function parameters:
        {{- range .Args }}
        {{printf "%-22s" .Name }} : {{ .Typer }}
        {{- end }}
        Return type: {{ .Returns }}
        '''
        fields = ({{ range $i, $a := .Args -}}
            {{ if $i }}, {{ end }} 
            {{- .NameNoIn }}
        {{- end }})
        query = self._buildquery('{{ .Schema }}.{{ .Function }}', fields)
        {{- if .ResultsSingular }}
        results = self.query(query, fields, paramdict, one_row=True)
        {{- else }}
        results = self.query(query, fields, paramdict)
        {{- end }}
        return results
