<template>
  <div class="app-container">
    <el-table
      :data="tableData"
      style="width: 100%"
      :default-sort = "{ prop: 'model_name', order: 'descending'}"
      element-loading-text="Loading"
      border
      fit
      highlight-current-row>

      <el-table-column  prop="model_name" align="center" label="模型名字" sortable  height="250" fit="true">
      </el-table-column>

      <el-table-column prop="model_channel" label="模型所属分类" sortable  align="center"
                       :filters="model_type_list"
                       :filter-method="filterCluster"
                       filter-placement="bottom-end" fit="true">
      </el-table-column>

      <el-table-column prop="mail_recipients" label="模型负责人邮箱" sortable width="330" align="center" >
        <template slot-scope="props">
            <div style="float:left;width:260px">
                <span  v-html="props.row.mail_recipients"/>
            </div>
            <div style="float:right;width:30px">
                <el-button @click="dialogUpdateMailRecipientsVisible = true, 
                updateFormModelTimeInfo.model_name=props.row.model_name,
                updateFormModelTimeInfo.mail_recipients=changeBrToDot(props.row.mail_recipients)" size="mini" type="primary" icon="el-icon-edit" circle>
                </el-button>
            </div>
            <div style="clear:both;height:0px;"></div>
        </template>
      </el-table-column>
      

      <el-table-column prop="update_average_time" label="7天平均更新时间" sortable  align="center" fit="true">
      </el-table-column>

      <el-table-column prop="latest_update_1" label="最近五次更新" sortable  align="center" fit="true">
        <template slot-scope="props">
            <span  v-html="props.row.latest_update_1"/>
        </template>
      </el-table-column>

    </el-table>

    <el-dialog title="修改数据" :visible.sync="dialogUpdateMailRecipientsVisible" :modal-append-to-body='false'>
        <el-form :model="updateFormModelTimeInfo" ref="updateFormModelTimeInfo">
            <el-form-item label="模型名称" :label-width="formLabelWidth">
                <el-input v-model="updateFormModelTimeInfo.model_name" auto-complete="off" readonly="readonly">
                </el-input>
            </el-form-item>
            <el-form-item label="模型负责人邮箱(多个邮箱用半角逗号分隔)" :label-width="formLabelWidth">
                <el-input v-model="updateFormModelTimeInfo.mail_recipients" auto-complete="off">
                </el-input>
            </el-form-item>
            <el-button @click="dialogUpdateMailRecipientsVisible = false">取 消</el-button>
            <el-button type="primary" @click="updateMailRecipients($event)">确 定</el-button>
        </el-form>
    </el-dialog>

  </div>
</template>

<script>
    import axios from 'axios'

    export default {
        data() {
            return {
                channelNames: [],
                model_type_list: [],
                channelToModels: {},
                modelNames: [],
                modelInfos: {},
                activeChannel: "1",
                node: '',
                start_time: '',
                status: '',
                qps: '',
                search: '',
                tableData: [],
                dialogUpdateMailRecipientsVisible: false,
                updateFormModelTimeInfo: {
                    model_name: '',
                    mail_recipients: '',
                },
                model_name: ''
            }
        },
        created: function() {
            this.getModels()
        },
        watch: {
            filterText(val) {
                this.$refs.tree2.filter(val)
            },
        },
        methods: {
            getModels: function() {
                axios.get('/mysql/show?table=models').then(async response => {
                    var res_data = response.data;
                    for (var i = 0; i < res_data.length; i++) {
                        var cur_model_data = res_data[i];
                        if (typeof (cur_model_data) == "undefined") {
                            continue;
                        }
                        var cur_model_name = cur_model_data.Name;
                        this.modelNames.push(cur_model_name);
                    }
                    var model_list="";
                    for (var i = 0; i < this.modelNames.length; i++) {
                        model_list += this.modelNames[i];
                        if (i != this.modelNames.length - 1) {
                            model_list += ","
                        }
                    }
                
                    var models_mail_recipients = [];
                    models_mail_recipients = await this.getRecipients(model_list);

                    var model_info_url = "/model_info/update_interval_week?model_list=";
                    model_info_url += model_list;
                    var model_list_info = [];
                    var model_channels = new Set();
                    axios.get(model_info_url).then(response => {
                        var resp_data = response.data;
                        for (var i = 0; i < resp_data.length; i++) {
                            var item;
                            var each_data = resp_data[i];
                            if ( each_data.LastestTimestampArray != null && typeof (each_data.ModelUpdateTimeWeekly) != "undefined") {
                                var update_interval_string;
                                var hour = parseInt(each_data.ModelUpdateTimeWeekly / 3600);
                                var minute = parseInt(each_data.ModelUpdateTimeWeekly % 3600 / 60);
                                if (hour == 0 && minute == 0) {
                                    update_interval_string = "7天内无更新"
                                } else {
                                    update_interval_string = hour + "小时" + minute + "分钟";
                                }
                                var update_timestamp_list = each_data.LastestTimestampArray;
                                if (update_timestamp_list != null) {
                                    item = {
                                        model_name: each_data.ModelName,
                                        update_average_time: update_interval_string,
                                        latest_update_1: this.spliceUpdateTime(update_timestamp_list),
                                        mail_recipients: this.changeDotToBr(models_mail_recipients[each_data.ModelName]),
                                        model_channel: each_data.ModelChannel
                                    }
                                }
                            } else {
                                item = {
                                    model_name: each_data.ModelName,
                                    update_average_time: '7天内无更新',
                                    latest_update_1: '',
                                    mail_recipients: this.changeDotToBr(models_mail_recipients[each_data.ModelName]),
                                    model_channel: each_data.ModelChannel
                                }
                            }
                            model_list_info.push(item);
                            model_channels.add(each_data.ModelChannel);
                        }
                        this.tableData = model_list_info;
                        for(let channel of model_channels) {
                            this.channelNames.push(channel);
                        }
                        this.modelTypeInit();


                        // 提示
                        this.$message({
                            message: '数据加载成功',
                            type: 'success',
                            customClass: 'login_alert',
                            duration: 2000
                        })
                    });
                });
            },
            getRecipients: async function(model_list) {
                var models_mail_recipients_url = "/model_info/models_mail_recipients?model_list=";
                models_mail_recipients_url += model_list;
                var models_mail_recipients = [];
                await axios.get(models_mail_recipients_url).then(response => {
                    models_mail_recipients = response.data;
                });
                return models_mail_recipients
            },
            modelTypeInit() {
                for(var i = 0; i < this.channelNames.length; i++) {
                    var item = {
                        text: this.channelNames[i],
                        value: this.channelNames[i]
                    };
                    this.model_type_list.push(item);
                }
            },
            spliceUpdateTime(update_timestamp_list) {
                var updateTimeStr = ""
                for (var i = 0; i <= 4; i++) {
                    if (typeof (update_timestamp_list[i]) == "undefined" || update_timestamp_list[i].length < 15) {
                        continue
                    }
                    updateTimeStr += this.formatTime(update_timestamp_list[i])
                    if (i < 4) {
                        updateTimeStr += "<br>"
                    }
                }
                return updateTimeStr
            },
            formatTime(time_stamp) {
                if (typeof (time_stamp) == "undefined" || time_stamp.length < 15) {
                    return "NA"
                }
                return time_stamp.substr(0,4) + "-" + time_stamp.substr(4,2) + "-"
                    + time_stamp.substr(6,2) + " " + time_stamp.substr(9,2) + ":"
                    + time_stamp.substr(11,2) + ":" + time_stamp.substr(13,2);
            },
            filterCluster(value, row) {
                return row.model_channel === value;
            },
            updateMailRecipients(event) {
                event.preventDefault();
                var update_mail_recipients_url = "/model_info/set_model_mail_recipients"
                axios.get(update_mail_recipients_url, {
                    params : {
                        model_name: this.updateFormModelTimeInfo.model_name,
                        mail_recipients: this.updateFormModelTimeInfo.mail_recipients
                    }
                }).then(function (response) {
                    console.log(response);
                }).catch(function (error) {
                    console.log(error);
                });
                this.dialogUpdateMailRecipientsVisible = false;
                location.reload();
                this.$route.push("/model_time_info/model_time_series_info")
            },
            changeDotToBr: function(modelsStr) {
                if(!(typeof modelsStr == "undefined" || modelsStr == null || modelsStr == "")){
                    modelsStr = modelsStr.replace(/,/g, '<br>');
                    console.log(modelsStr);
                }
                return modelsStr;
            },
            changeBrToDot: function(modelsStr) {
                if(!(typeof modelsStr == "undefined" || modelsStr == null || modelsStr == "")){
                    modelsStr = modelsStr.replace(/<br>/g, ',');
                    console.log("brtodot");
                    console.log(modelsStr);
                }
                return modelsStr;
            },
        }
    }
</script>
