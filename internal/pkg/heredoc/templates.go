package heredoc

import "text/template"

// addTemplates adds the custom templates to the given template.
func addTemplates(t *template.Template) error {
	_, err := t.Parse(`
	{{- define "mdCodeOrBold" -}}
		{{- if IsMD -}}
			{{ Code . }}
		{{- else -}}
			{{ Bold . }}
		{{- end}}
	{{- end}}`)
	if err != nil {
		return err
	}

	return nil
}
