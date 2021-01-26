<template>
  <div class="app-container">
    <el-collapse v-model="activeModel" :data="modelNames"  @change="clickModelTab" accordion>
      <div v-for="(model, index) in modelNames">
        <el-collapse-item :title="model" :name="index">
          <div id="app">
            <div class="clsmsg">
              <el-select  v-model="activeNum" placeholder="5" @change="selectModelNum">
                <el-option
                  v-for="num in options"
                  :key="num.value"
                  :label="num.label"
                  :value="num.value">
                </el-option>
              </el-select>
              <el-button @click="refresh()" size="medium" type="primary"> <i class="el-icon-refresh" />&nbsp;&nbsp;刷新
              </el-button>
            </div>
          </div>
          <el-table :data="modelInfos[model]">
            <el-table-column prop="timestamp" label="timestamp" sortable>
            </el-table-column>

            <el-table-column prop="status" label="status" sortable align="center">
                <template slot-scope="props">
                    <el-tag v-if="props.row.is_new_version && props.row.status!='' && props.row.percent == 100 " type="success">{{props.row.status}}</el-tag>
                    <el-tag v-if="props.row.is_new_version && props.row.status!='' && props.row.percent != 100 " type="warning">loading</el-tag>
                    <el-tag v-if="!props.row.is_new_version && props.row.status!='' && props.row.percent != 100 " type="warning">{{props.row.status}}</el-tag>
                </template>    
            </el-table-column>
            <el-table-column prop="percent" label="progress" sortable align="center">
                <template slot-scope="props">
                    <el-progress v-if="props.row.percent>0" :percentage="props.row.percent"></el-progress>
                </template>    
            </el-table-column>

            <el-table-column prop="md5" label="md5" sortable>
            </el-table-column>

            <el-table-column prop="is_locked" label="is_locked" sortable="">
            </el-table-column>

            <el-table-column prop="desc" label="desc" sortable>
            </el-table-column>

            <el-table-column prop="createdAt" label="createdAt" sortable>
            </el-table-column>

            <el-table-column prop="updatedAt" label="updatedAt" sortable>
            </el-table-column>

          </el-table>
        </el-collapse-item>
      </div>
    </el-collapse>
  </div>
</template>

<script>
    import axios from 'axios'

    export default {
        data() {
            return {
                modelNames: [],
                modelInfos: {},
                service_to_host: {},
                tableData: [],
                activeModel: 'llearner_lr_v1',
                activeNum: '',
                selectModel: '',
                selectNum: '',
                options: [{
                    value: '5',
                    label: '5'
                }, {
                    value: '10',
                    label: '10',
                }, {
                    value: 'all',
                    label: 'all'
                }],
                node: '',
                start_time: '',
                status: '',
                percent: 0,
                qps: '',
                search: '',
                defaultProps: {
                    children: 'children',
                    label: 'label',
                }
            }
        },
        created () {
            axios.get('/mysql/show?table=models').then(response => {
                var res_data = response.data;
                var modelSet = new Set();
                for (var i = 0; i < res_data.length; i++) {
                    var cur_model_data = res_data[i];
                    if (typeof (cur_model_data) == "undefined") {
                        continue;
                    }
                    var cur_model_name = cur_model_data.Name;
                    modelSet.add(cur_model_name);
                }
                for (var model_name of modelSet) {
                    this.modelNames.push(model_name);
                }
            })
        },
        watch: {
            filterText(val) {
                this.$refs.tree2.filter(val)
            }
        },
        methods: {
            filterNode(value, data) {
                if (!value) return true
                return data.label.indexOf(value) !== -1
            },
            filterStatus(value, row) {
                return row.node_status === value;
            },

            clickModelTab(index) {
                var click_model_name = this.modelNames[index];
                if( click_model_name != null) {
                    this.$message({
                        message: '数据加载中....',
                        type: 'success',
                        customClass: 'login_alert',
                        duration: 2000
                    })
                }
                console.log("handleChange val:", click_model_name);
                this.selectModel = click_model_name;
                // default num=5
                var model_info_url = "/model_info/model_history?modelname=" + click_model_name
                	+ "&number=5";
                var node_info = [];
                axios.get(model_info_url).then(response => {
                    response.data.sort(function (x, y) {
                        if (x.Timestamp < y.Timestamp) {
                            return 1;
                        }else {
                            return -1;
                        }
                    })
                    var is_new_version_flag = false;
                    for (var i = 0; i < response.data.length; i++) {
                        var cur_info = response.data[i];
                        var is_new_version = false;
                        if (!is_new_version_flag && cur_info.Status == "loaded") {
                            is_new_version = true;
                            is_new_version_flag = true;
                        }
                        var item = {
                            timestamp: cur_info.Timestamp,
                            status: cur_info.Status,
                            is_new_version : is_new_version,
                            percent: cur_info.Percent,
                            md5: cur_info.Md5,
                            is_locked: cur_info.IsLocked,
                            desc: cur_info.Desc,
                            createdAt: cur_info.CreatedAt,
                            updatedAt: cur_info.UpdatedAt
                        };
                        node_info.push(item);
                    }
                })
                this.modelInfos[click_model_name] = node_info;
            },
            refresh: function(){
                var model_num = 5;
                if (this.selectNum != '') {
                    model_num = this.selectNum;
                }
                this.selectModelNum(model_num);
            },
            selectModelNum(model_num) {
                if( this.selectModel != null) {
                    this.$message({
                        message: '数据加载中....',
                        type: 'success',
                        customClass: 'login_alert',
                        duration: 2000
                    })
                }
                this.selectNum = model_num;
                this.activeNum = model_num;
                window.localStorage.setItem("selectModel", this.selectModel);
                window.localStorage.setItem("selectNum", this.selectNum);
                console.log("model name is:", this.selectModel);
                if (this.selectModel.length != 0) {
                    var model_info_url = "/model_info/model_history?modelname=" + this.selectModel
                        + "&number=" + model_num;
                    var node_info = [];
                    axios.get(model_info_url).then(response => {
                        response.data.sort( function (x, y) {
                            if (x.Timestamp < y.Timestamp) {
                                return 1;
                            } else {
                                return -1;
                            }
                        })
                        var is_new_version_flag = false;
                        for (var i = 0; i < response.data.length; i++) {
                            var cur_info = response.data[i];
                            var is_new_version = false;
                            if (!is_new_version_flag && cur_info.Status == "loaded") {
                                is_new_version = true;
                                is_new_version_flag = true;
                            }
                            var cur_info = response.data[i];
                            var item = {
                                timestamp: cur_info.Timestamp,
                                status: cur_info.Status,
                                is_new_version: is_new_version,
                                percent: cur_info.Percent,
                                md5: cur_info.Md5,
                                is_locked: cur_info.IsLocked,
                                desc: cur_info.Desc,
                                createdAt: cur_info.CreatedAt,
                                updatedAt: cur_info.UpdatedAt
                            };
                            node_info.push(item);
                            this.modelInfos[this.selectModel].splice(i, 1, item);
                        }
                    })
                }
            }
        }
    }
</script>
