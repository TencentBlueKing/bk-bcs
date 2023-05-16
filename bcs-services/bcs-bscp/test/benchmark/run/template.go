package run

import (
	"html/template"
	"io"
)

var tp *defaultTemplate

type defaultTemplate struct {
	engine *template.Template
}

func init() {
	t := template.New("template")
	t = template.Must(t.Parse(htmlTemplate))

	tp = &defaultTemplate{
		engine: t,
	}
}

func (t defaultTemplate) render(wr io.Writer, results []metricsData) {
	t.engine.Execute(wr, map[string][]metricsData{
		"Results": results,
	})
}

// htmlTemplate test results file statistical report html template.
const htmlTemplate = `
<html>
	<head>
		<title>{{.Name}}</title>
		<meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>
			<style type="text/css">
				table.gridtable {
					font-family: verdana,arial,sans-serif;
					font-size:12px;
					color:#333333;
					border-width: 1px;
					border-color: #666666;
					border-collapse: collapse;
				}
				table.gridtable th {
					border-width: 1px;
					padding: 8px;
					border-style: solid;
					border-color: #666666;
					background-color: #dedede;
				}
				table.gridtable td {
					border-width: 1px;
					padding: 8px;
					border-style: solid;
					border-color: #666666;
					background-color: #ffffff;
				}
			</style>
	</head>
	<body>
		<table class="gridtable">
			<thead>
				<tr>
					<td>Title</td>
					<td>QPS</td>
					<td>AverageDuration</td>
					<td>Percent95Duration</td>
					<td>MaxDuration</td>
					<td>MinDuration</td>
					<td>MedianDuration</td>
					<td>Percent85Duration</td>
					<td>TotalRequest</td>
					<td>SucceedRequest</td>
					<td>FailedRequest</td>
					<td>OnTheFlyRequest</td>
					<td>SustainSeconds</td>
					<td>Concurrent</td>
				</tr>
			</thead>
		
			<tbody>
				{{ with .Results }}
				{{- range . }}
				<tr>
					<td>{{.Title}}</td>
					<td style='color:blue;'>{{.Metrics.QPS}}</td>
					<td style='color:blue;'>{{.Metrics.AverageDuration}}ms</td>
					<td>{{.Metrics.Percent95Duration}}ms</td>
					<td>{{.Metrics.MaxDuration}}ms</td>
					<td>{{.Metrics.MinDuration}}ms</td>
					<td>{{.Metrics.MedianDuration}}ms</td>
					<td>{{.Metrics.Percent85Duration}}ms</td>
					<td>{{.Metrics.TotalRequest}}</td>
					<td>{{.Metrics.SucceedRequest}}</td>
					<td>
						{{ if eq .Metrics.FailedRequest 0 }}
							<p>0</p>
						{{ else }}
							<p style='color:red;'>{{.Metrics.FailedRequest}}</p>
						{{ end }}
					</td>
					<td>{{.Metrics.OnTheFlyRequest}}</td>
					<td >{{.Metrics.SustainSeconds}}</td>
					<td>{{.Metrics.Concurrent}}</td>
				</tr>
				{{- end }}
				{{ end }}
			</tbody>
		</table>
		<br>
</body>
</html>
`
