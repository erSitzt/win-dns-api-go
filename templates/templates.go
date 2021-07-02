package templates

import "text/template"

var ZoneTemplate = template.Must(template.New("").Parse(
	`$ORIGIN {{.ZoneName}}.
{{range .AllRecords -}}
{{.Name | printf "%-30s"}} {{.TTL | printf "%6d"}}  {{.Type | printf "%-10s"}} {{.Value}}
{{end -}}
`))

var ZoneListTemplate = template.Must(template.New("").Parse(
	`{{range .AllZones -}}
{{.Name | printf "%-50s"}} {{.Type | printf "%-15s"}} {{.Storage | printf "%-15s"}} {{.Properties}} 
{{end -}}
`))
