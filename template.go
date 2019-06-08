package main

var (
	indexTemplate = `<!DOCTYPE html>
				<html lang="en">
				<head>
					<meta charset="UTF-8">
					<title>V3 File Durability Status</title>
				</head>
				<body>
						<h1>V3 File Availability & Durability Status</h1>
						<h3>Current Percentage: {{.Percentage}}%</h3>
						<h4>Filename: {{.File}}</h4>
						<h4>Size: {{.Size}} Bytes</h4>
						<h4>Check Interval: {{.Interval}}</h4>
						<table>
							<tr>
								<th>Check Nr.</th>
								<th>Timestamp</th>
								<th>Success Status</th>
								<th>Error</th>
							</tr>
								{{ range $Key, $Stat := .List }}<tr>
									<td>{{$Key}}</td>
									<td>{{$Stat.Time.Format "Jan 02, 2006 15:04:05 UTC" }}</td>
									<td>{{$Stat.Success}}</td>
									{{if $Stat.Error}}<td>{{$Stat.Error}}</td>{{end}}
								</tr>{{end}}
						</table>
				</body>
				</html>`
)
