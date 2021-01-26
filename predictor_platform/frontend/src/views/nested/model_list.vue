<script>
    export default {
        data() {
        },
        methods: {
            handleChange(val) {
                console.log(val);
            }
        }
    }
</script>

<template>
  <div class="app-container">
    <el-collapse :data="modelNames">
      <div style="font-size:30px" v-for="(model_name, index) in modelNames">
        <el-collapse-item :title="model_name">
          <div style="text-indent:1em;" v-for="(service_name, index1) in modelToService[model_name]">
            <el-collapse-item :title="service_name" style="font-size:20px">
              <div id="app">
                <div class="clsmsg">
                  <el-tag effect="dark">部署机器:{{serviceToHostInfo[model_name][service_name].length}}</el-tag>
                </div>
              </div>
              <el-table :data="serviceToHostInfo[model_name][service_name].filter(data => !search || data.node.toLowerCase().includes(search.toLowerCase())
                || data.timestamp.toLowerCase().includes(search.toLowerCase()) || data.status.toLowerCase().includes(search.toLowerCase())
                || data.path.toLowerCase().includes(search.toLowerCase()) || data.md5.toLowerCase().includes(search.toLowerCase()))">

                <el-table-column prop="node" label="Host" sortable>
                </el-table-column>

                <el-table-column prop="timestamp" label="timestamp" sortable>
                </el-table-column>

                <el-table-column prop="status" label="模型状态" sortable
                  :filters="[{text: 'loaded', value: 'loaded'},{text: 'failed', value: 'failed'}, {text: 'loading', value: 'loading'}]"
                  :filter-method="filterStatus"
                  filter-placement="bottom-end">
                </el-table-column>

                <el-table-column prop="path" label="模型路径" sortable>
                </el-table-column>

                <el-table-column prop="md5" label="md5值" sortable>
                </el-table-column>

                <el-table-column align="left">
                  <template slot="header" slot-scope="scope">
                    <el-input size="medium" style="width: 280px;" placeholder="输入搜索内容" v-model="search"  maxlength="280">
                    </el-input>
                  </template>
                </el-table-column>

              </el-table>
            </el-collapse-item>
          </div>
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
                serviceToHostInfo: {},
                modelToService: {},
                model_to_host_count: 0,
                service: '',
                search: '',
                tableData: [],
                cols: [
                    { prop: 'node', label: 'Host'},
                    { prop: 'timestamp', label: 'timestamp'},
                    { prop: 'status', label: '模型状态'},
                    { prop: 'path', label: '模型路径'},
                    { prop: 'md5', label: 'md5值'},
                ],
                defaultProps: {
                    children: 'children',
                    label: 'label',
                },
            }
        },
        created () {
            axios.get('/node_infos').then(response => {
                var res_data = response.data;
                var modelSet = new Set();
                for (var i = 0; i < res_data.length; i++) {
                    var cur_status_info = res_data[i].StatusInfo;
                    var cur_status_info_length = (cur_status_info != null ? cur_status_info.length : 0);
                    for (var j = 0; j < cur_status_info_length; j++) {
                        var cur_models = cur_status_info[j].model_records;
                        cur_models.forEach(function (model) {
                            modelSet.add(model.name);
                        });
                    }
                }
                for (var each_model of modelSet) {
                    this.modelNames.push(each_model);
                }
                // 遍历当前所有的模型，绑定模型和service name
                var table_set = new Set();
                var model_to_service = {};
                this.modelNames.forEach(function (each_model) {
                    // 以node为单位来分析数据
                    var service_names = new Set();
                    for(var i = 0; i < res_data.length; i++) {
                        var status_info_list = res_data[i].StatusInfo;
                        if (status_info_list == null) {
                            continue;
                        }
                        status_info_list.forEach(function (each_status_info) {
                            var cur_service_name = each_status_info.service_name;
                            var cur_model_records = each_status_info.model_records;
                            for (var index = 0; index < cur_model_records.length; index++) {
                                if (cur_model_records[index].name == each_model && typeof(cur_service_name) != "undefined") {
                                    service_names.add(cur_service_name);
                                    // var item
                                    var item = {
                                        model: each_model,
                                        service_name: cur_service_name,
                                        node: res_data[i].Host,
                                        timestamp: cur_model_records[index].timestamp,
                                        succtime: cur_model_records[index].success_time,
                                        status: cur_model_records[index].state,
                                        path: cur_model_records[index].configName,
                                        md5: cur_model_records[index].md5,
                                    };
                                    table_set.add(item);
                                }
                            }
                            if (service_names.size != 0) {
                                model_to_service[each_model] = service_names;
                            }
                        })
                    }
                })
                //this.modelToService = model_to_service;
                for (var key in model_to_service) {
                    var service_name_set = model_to_service[key];
                    var service_list = [];
                    for (var service_name of service_name_set) {
                        service_list.push(service_name);
                    }
                    this.modelToService[key] = service_list;
                }
                for (var model of this.modelNames) {
                    var service_list = this.modelToService[model];
                    for (var service of service_list) {
                        var model_service_to_host = new Set();
                        table_set.forEach(
                            function (item) {
                                var time_format = item.succtime.split('_');
                                var year_month_str = time_format[0];
                                var minute_second_str = time_format[1];
                                var succ_time_str = "";
                                if (year_month_str.length != 0 && minute_second_str.length != 0) {
                                    succ_time_str += year_month_str.substr(0,4) + "/" + year_month_str.substr(4,2) + "/" + year_month_str.substr(6,2) + " ";
                                    succ_time_str += minute_second_str.substr(0,2) + ":" + minute_second_str.substr(2,2) + ":" + minute_second_str.substr(4,2);
                                }                                 
                                if (model == item.model && service == item.service_name) {
                                    var host_status = {
                                        node: item.node,
                                        timestamp: item.timestamp,
                                        status: item.status == "loaded" ? "loaded at " + succ_time_str : item.status,
                                        path: item.path,
                                        md5: item.md5,
                                    };
                                    model_service_to_host.add(host_status);
                                }
                            }
                        );
                        var model_service_host = [];
                        for(var each of model_service_to_host) {
                            model_service_host.push(each);
                        }
                        if (typeof this.serviceToHostInfo[model] == "undefined") {
                            this.serviceToHostInfo[model] = {
                                service_name : model_service_host,
                            }
                        }
                        this.serviceToHostInfo[model][service] = model_service_host;
                        if(this.serviceToHostInfo[model][service] != "undefined" && this.serviceToHostInfo[model][service] != null) {
                            this.model_to_host_count = this.serviceToHostInfo[model][service].length;
                        }
                    } //第二个for循环
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
                if (row.status.indexOf(value) > -1)
                    return true
            }
        }
    }
</script>
