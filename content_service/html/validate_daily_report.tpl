<html>
  <head><meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
    <title>一致性验证结果</title>
    <style type="text/css">
      <!--
      body{margin:0; font-family:Tahoma;Simsun;font-size:12px;}
      table{border:1px #000000 solid;}
      td{border:1px #000000 solid;}
      .title {font-size: 12px; COLOR: #FFFFFF; font-family: Tahoma;}
      .desc {font-size: 12px; COLOR: #000000; font-family: Tahoma;}
      -->
    </style>
  </head>
  <body><center>
    <h1>{{.Subject}} ({{.Date}})</h1>
    <h2>分模型验证情况总览</h2>
    <table cellpadding="0" cellspacing="0" border=0 style='border-collapse:collapse;border:none'>
    <tr bgcolor=#2561CF align=center class=title>
    {{range .SummaryHeader}}
      <td style="min-width:100px">{{.}}</td>
    {{end}}
    </tr>
    {{range .SummaryData}}
      <tr bgcolor=#FFFFFF class=desc>
      {{range $i,$e := .}}
        {{if eq $i 3}}
          {{if gt $e 0}}
            <td style="min-width:100px" height=30 align=center bgcolor="#FF0000">{{$e}}</td>
          {{else}}
            <td style="min-width:100px" height=30 align=center>{{$e}}</td>
          {{end}}
        {{else}}
          <td style="min-width:100px" height=30 align=center>{{$e}}</td>
        {{end}}
      {{end}}
      </tr>
    {{end}}
    </table><br>
    <h2>分模型验证情况明细(点击名称可展开)</h2>
    {{range .DetailResults}}
      <details>
      <summary><font size="4">{{.ModelName}}</font></summary>
        <table cellpadding="0" cellspacing="0" border=0 style='border-collapse:collapse;border:none'>
        <tr bgcolor=#2561CF align=center class=title>
        {{range .DetailHeader}}
          <td style="min-width:100px">{{.}}</td>
        {{end}}
        </tr>
        {{range .DetailData}}
          <tr bgcolor=#FFFFFF class=desc>
          {{range .}}
            <td style="min-width:100px" height=30 align=center>{{.}}</td>
          {{end}}
          </tr>
        {{end}}
        </table><br>
      </details>
    {{end}}
  </center></body>
</html>
