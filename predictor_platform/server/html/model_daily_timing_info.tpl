<html>
  <head><meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
    <title>模型时效性结果--日期({{.Date}})</title>
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
  <body>
    <center>
    <h1>模型时效性报告--日报</h1>
    <h2>模型时效性一览表</h2>
    <table cellpadding="0" cellspacing="0" border=0 style='border-collapse:collapse;border:none'>
    <tr bgcolor=#2561CF align=center class=title>
    {{range .TableHeader}}
      <td style="min-width:100px">{{.}}</td>
    {{end}}
    </tr>
    {{range .ModelData}}
      <tr bgcolor=#FFFFFF class=desc>
      {{range . }}
        <td style="min-width:100px" height=30 align=center>{{.}}</td>
      {{end}}
      </tr>
    {{end}}
    </table><br>
    </center>
  </body>
</html>