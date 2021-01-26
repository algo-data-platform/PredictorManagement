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
    {{if eq .ClaimStatus "(待认领)"}}
      <hr align=center width=600 color=#987cb9 size=3>
      <a href="http://127.0.0.123:80/#/model_time_info/model_time_series_info"><font size="5">模型负责人请点击此链接认领模型</font></a>
      <hr align=center width=600 color=#987cb9 size=3>
    {{end}}
    <h1>一致性验证报告</h1>
    <h2>模型名称：{{.ModelName}}</h2>
    <h2>模型版本：{{.ModelVersion}}</h2>
    <h2>验证类型：{{.ValidateType}}</h2>
    {{if eq .ValidateConclusion "通过"}}
      <h2>验证结果：<font color="#00FF00">{{.ValidateConclusion}}</font></h2>
    {{else}}
      <h2>验证结果：<font color="#FF0000">{{.ValidateConclusion}}</font></h2>
    {{end}}
    {{if eq .ValidateConclusion "不通过，模型加载失败"}}
      <details>
      <summary><font size="4">Predictor含[{{.ModelName}}]、[{{.ModelVersion}}]关键字日志记录</font></summary>
      <div align=left style="background:grey; overflow:scroll; width:1000px; height:300px; border:3px solid #F00">
      <pre>{{.AlgoLog}}</pre>
      </div>
      </details>
      <a href={{.AlgoLogUrl}}><font size="4">Predictor全量日志</font></a>
    {{end}}
    {{if eq .ValidateType "样本验证"}}
      <h2>样本数量：{{.SampleCount}}</h2>
      <h2>通过样本：{{.SamplePassCount}}</h2>
      <h2>未通过样本：{{.SampleFailCount}}</h2>
      {{if gt .SampleFailCount 0}}
        <details>
        <summary><font size="4">未通过样本明细(点击可展开)</font></summary>
        <table cellpadding="0" cellspacing="0" border=0 style='border-collapse:collapse;border:none'>
        <tr bgcolor=#2561CF align=center class=title>
        {{range .Header}}
          <td style="min-width:100px">{{.}}</td>
        {{end}}
        </tr>
        {{range .Data}}
          <tr bgcolor=#FFFFFF class=desc>
          {{range .}}
            <td height=30 align=center style="min-width:100px">{{.}}</td>
          {{end}}
          </tr>
        {{end}}
        </table><br>
        </details>
      {{end}}
      <h2>无效样本：{{.SampleInvalidCount}}</h2>
      {{if gt .SampleInvalidCount 0}}
        <details>
        <summary><font size="4">无效样本明细(点击可展开)</font></summary>
          <table cellpadding="0" cellspacing="0" border=0 style='border-collapse:collapse;border:none'>
          <tr bgcolor=#2561CF align=center class=title>
          {{range .InvalidHeader}}
            <td style="min-width:100px">{{.}}</td>
          {{end}}
          </tr>
          {{range .InvalidData}}
            <tr bgcolor=#FFFFFF class=desc>
            {{range .}}
              <td height=30 align=center style="min-width:100px">{{.}}</td>
            {{end}}
            </tr>
          {{end}}
          </table><br>
        </details>
      {{end}}
    {{end}}
  </center></body>
</html>
