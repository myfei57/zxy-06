package console

import (
	"html/template"
	"net/http"
)

var templates = template.Must(template.New("pages").Parse(`
{{define "layout"}}<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>{{.Title}}</title>
<style>
body{font-family:sans-serif;margin:24px;color:#222}
nav a{margin-right:12px}
table{border-collapse:collapse;margin-top:12px}
td,th{border:1px solid #ccc;padding:6px 10px;font-size:13px}
</style></head><body>
<nav><a href="/">总览</a><a href="/console/flights">航班</a><a href="/console/tasks">保障任务</a><a href="/console/resources">资源</a><a href="/console/audit">审计</a></nav>
<h2>{{.Title}}</h2>{{.Body}}</body></html>{{end}}
{{define "index"}}{{template "layout" .}}{{end}}
{{define "flights"}}{{template "layout" .}}{{end}}
{{define "tasks"}}{{template "layout" .}}{{end}}
{{define "resources"}}{{template "layout" .}}{{end}}
{{define "audit"}}{{template "layout" .}}{{end}}
`))

type pageData struct {
	Title string
	Body  template.HTML
}

func (a *API) render(w http.ResponseWriter, name, title string, body template.HTML) {
	data := pageData{Title: title, Body: body}
	if err := templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *API) IndexPage(w http.ResponseWriter, r *http.Request) {
	totalF, activeF, _ := flightCounts(a.state)
	totalT, activeT, _ := taskCounts(a.state)
	body := template.HTML(`<p>航班 ` + itoa(totalF) + ` 班（活跃 ` + itoa(activeF) + `），保障任务 ` + itoa(totalT) + ` 条（进行中 ` + itoa(activeT) + `）。</p>`)
	a.render(w, "index", "总览", body)
}

func (a *API) FlightsPage(w http.ResponseWriter, r *http.Request) {
	rows := ""
	for _, f := range a.state.Flights() {
		rows += `<tr><td>` + f.FlightNo + `</td><td>` + f.Status + `</td><td>` + f.UpdatedAt.Format("2006-01-02 15:04:05") + `</td></tr>`
	}
	body := template.HTML(`<table><tr><th>航班</th><th>状态</th><th>最新动态</th></tr>` + rows + `</table>`)
	a.render(w, "flights", "航班动态", body)
}

func (a *API) TasksPage(w http.ResponseWriter, r *http.Request) {
	rows := ""
	for _, t := range a.state.Tasks() {
		rows += `<tr><td>` + t.TaskType + `</td><td>` + t.Status + `</td><td>` + t.Owner + `</td></tr>`
	}
	body := template.HTML(`<table><tr><th>任务</th><th>状态</th><th>执行人</th></tr>` + rows + `</table>`)
	a.render(w, "tasks", "保障任务", body)
}

func (a *API) ResourcesPage(w http.ResponseWriter, r *http.Request) {
	rows := ""
	for _, rs := range a.state.Resources() {
		rows += `<tr><td>` + rs.Name + `</td><td>` + rs.Kind + `</td><td>` + rs.Status + `</td></tr>`
	}
	body := template.HTML(`<table><tr><th>资源</th><th>类型</th><th>状态</th></tr>` + rows + `</table>`)
	a.render(w, "resources", "资源占用", body)
}

func (a *API) AuditPage(w http.ResponseWriter, r *http.Request) {
	rows := ""
	for _, e := range a.state.RecentAudit(50) {
		rows += `<tr><td>` + e.Action + `</td><td>` + e.Target + `</td><td>` + e.Detail + `</td><td>` + e.At.Format("2006-01-02 15:04:05") + `</td></tr>`
	}
	body := template.HTML(`<table><tr><th>动作</th><th>对象</th><th>详情</th><th>时间</th></tr>` + rows + `</table>`)
	a.render(w, "audit", "审计日志", body)
}

func itoa(v int) string  { return formatInt(int64(v)) }
func itoa64(v int64) string { return formatInt(v) }
func ftoa(v float64) string { return formatFloat(v) }
