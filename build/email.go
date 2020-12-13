package build

import (
	"awesomeProject/git"
	"bytes"
	"html/template"
	"log"
	"strings"

	"gopkg.in/gomail.v2"
)

// WorkOrderError struct used for work order errors
type WorkOrderError struct {
	Service     string
	Environment string
	Error       error
}

// WorkOrderCommitInfo struct used for work order commit info
type WorkOrderCommitInfo struct {
	Service     string
	Environment string
	CommitInfo  []git.CommitInfo
}

// getSuccessfulBuildTemplate Func
func getSuccessfulBuildTemplate(environment string, stories []string, commitInfo []WorkOrderCommitInfo) string {

	const tpl = `
		<!DOCTYPE html>
		<html>
		<head>
		  <meta charset="UTF-8">
			<style>
			table, th, td {
		    border: 1px solid Gainsboro;
		    border-collapse: collapse;
		}
		th {
		padding-top: 10px;
		padding-bottom: 10px;
		background-color: Azure;
		}
		</style>

		</head>
		<body>
		<div><img src="cid:tw.gif" width="300" height="64"></div>
		<h2>The following stories have been released to {{.Environment}}.</h2>
		<table style="width:50%">
		<tr>
			<th>Ticket#</th>
		</tr>
		{{range .Stories}}
		<tr>
			<td align="center">{{.}}</td>
		</tr>{{end}}
		</table>
		<div style="padding-bottom:30px;"></div>
		{{range .WorkOrderCommitInfo}}
		<h2>Release details for {{.Service}} in {{.Environment}}.</h2>
		<div>
		<table style="width:100%">
			<tr>
		    <th>Description</th>
		    <th>Commit</th>
		    <th>Author</th>
				<th>Date</th>
			</tr>
		{{range .CommitInfo}}
			<tr>
				<td>{{.Description}}</td>
				<td>{{.Commit}}</td>
				<td>{{.Author}}</td>
				<td>{{.Date}}</td>
			</tr>{{end}}
		</table>
		<div style="padding-bottom:20px;"></div>
		</div>
		{{end}}
		<div style="padding-top:15px; padding-bottom:20px">Contact the <a href="mailto:TW_Services_Group@trnswrks.com">Service Team</a> with any questions.</div>
		</body>
		</html>`

	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	t, err := template.New("releaseNotification").Parse(tpl)
	check(err)
	data := struct {
		Environment         string
		Stories             []string
		WorkOrderCommitInfo []WorkOrderCommitInfo
	}{
		Environment:         strings.ToUpper(environment),
		Stories:             stories,
		WorkOrderCommitInfo: commitInfo,
	}

	bytes := bytes.NewBufferString("")
	err = t.Execute(bytes, data)
	check(err)
	return string(bytes.Bytes())
}

// getFailedBuildTemplate Func
func getFailedBuildTemplate(errors []WorkOrderError) string {
	const tpl = `
		<!DOCTYPE html>
		<html>
		<head>
		  <meta charset="UTF-8">
			<style>
			table, th, td {
		    border: 1px solid Gainsboro;
		    border-collapse: collapse;
		}
		th {
		text-align: left;
		padding-top: 10px;
		padding-bottom: 10px;
		background-color: Azure;
		}
		</style>

		</head>
		<body>
	<div><img src="test.gif" width="300" height="64"></div>
		{{range .}}
		<h2>The build for {{.Service}} in {{.Environment}} has failed.</h2>
		<div>
		<table style="width:100%">
		<th>Errors</th>
			<tr>
			{{if .Error}}
				<td>{{.Error}}</td>
			{{else}}
				<td>none</td>
			{{end}}
			</tr>
		</table>
		</div>{{end}}
		<div style="padding-top:15px; padding-bottom:20px"></div>
		</body>
		</html>`

	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	t, err := template.New("releaseNotification").Parse(tpl)
	check(err)

	bytes := bytes.NewBufferString("")
	err = t.Execute(bytes, errors)
	check(err)

	return string(bytes.Bytes())
}

// SendSucccessEmail func to send email on build success
func sendSucccessEmail(environment string, stories []string, commitInfo []WorkOrderCommitInfo) error {

	m := gomail.NewMessage()

	m.SetHeader("From", "services@test.com")
	m.SetHeader("To", "testers@test.com")
	m.SetHeader("Cc", "group@test.com", "team@test.com")
	m.SetHeader("Subject", strings.ToUpper(environment)+" Release Notes!")
	m.SetBody("text/html", getSuccessfulBuildTemplate(environment, stories, commitInfo))

	d := gomail.NewDialer("localhost", 25, "", "")
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

// SendFailureEmail func to send email on build failure
func sendFailureEmail(environment string, errors []WorkOrderError) error {

	m := gomail.NewMessage()

	m.SetHeader("From", "services@test.com")
	m.SetHeader("To", "servicesp@test.com")
	m.SetHeader("Cc", "team@test.com")
	m.SetHeader("Subject", strings.ToUpper(environment)+" Build Failed!")
	m.SetBody("text/html", getFailedBuildTemplate(errors))

	d := gomail.NewDialer("localhost", 25, "", "")
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
