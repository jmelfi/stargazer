{{- $wl := .WithLicense -}}
{{- $ws := .WithStars -}}
{{- $wb := .WithBtt -}}
{{- $a := .Anchors -}}
{{- $s := .Stars -}}
# Awesome Starred Repos List

{{ .Credits.Text }}{{ .Credits.Link }}  
Total starred repositories: `{{ .Total }}`

{{- if .WithToc }}
## Contents
{{ range $key := .Keys }}
  - [{{ $key }}](#{{ with (index $a $key) }}{{ . }}{{ end }})
{{- end }}
{{- end }}


{{ range $key := .Keys }}
## {{ $key }}
| Name  | Description {{ if $wl }} | License {{ end }}{{ if $ws }} | Stars {{ end }} |
| ----- | -----{{ if $wl }} | :---:{{ end }}{{ if $ws }} |----:{{ end }} |
{{- with (index $s $key) }}{{ range . }}
| [{{- .NameWithOwner -}}]({{- .Url -}}) | {{ .Description }} {{ if .Archived }}(*archived*){{ end }} {{ if $wl }} | {{ with .License}}{{ . }}{{ else }}-{{ end }}{{ end }} {{ if $ws }}| ⭐️{{ .Stars }}{{ end }} |
{{- end }}
{{- end }}
{{- if $wb }} 

**[⬆ back to top](#contents)**{{ end }}
{{- end }}
