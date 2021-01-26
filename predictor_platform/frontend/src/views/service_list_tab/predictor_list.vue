<template>
  <div class="app-container">
    <el-collapse :data="serviceNames">
      <div v-for="(service, index) in serviceNames">
        <el-collapse-item :title="service" :name="index">
          <div id="app">
            <div class="clsmsg">
              <el-tag>部署机器:{{service_to_host[service].length}}</el-tag>
            </div>
          </div>

            <el-table :data="service_to_host[service].filter(data => !search || data.node.toLowerCase().includes(search.toLowerCase())
                || data.start_time.toLowerCase().includes(search.toLowerCase()))">
              <el-table-column prop="node" label="Host" sortable>
              </el-table-column>

              <el-table-column prop="start_time" label="运行时间" sortable>
              </el-table-column>

              <el-table-column prop="status" label="服务状态" sortable
                               :filters="[{text: 'on', value: 'on'},{text: 'off', value: 'off'}]"
                               :filter-method="filterStatus"
                               filter-placement="bottom-end">
              </el-table-column>

              <el-table-column prop="qps" label="服务qps" sortable>
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
    </el-collapse>
  </div>
</template>

<script>
    import axios from 'axios'

export default {
  data() {
    return {
      serviceNames: [],
      service_to_host_count: 0,
      service_to_host: {},
      tableData: [],
      node: '',
      start_time: '',
      status: '',
      qps: '',
      search: '',
      defaultProps: {
        children: 'children',
        label: 'label',
      }
    }
  },
    created () {
        axios.get('/node_infos').then(response => {
                var res_data = response.data;
                var service_name_set = new Set();
                for (var index = 0; index < res_data.length; index++) {
                    var status_info = res_data[index].StatusInfo;
                    var status_info_length = (status_info != null ? status_info.length : 0);
                    for (var i = 0; i < status_info_length; i++) {
                        service_name_set.add(status_info[i].service_name);
                        var item = {
                            service_name : status_info[i].service_name,
                            node: res_data[index].Host,
                            start_time: '',
                            status: '',
                            qps: '',
                        }
                        this.tableData.push(item);
                    }
                }
                // 根据service name，update tableData中的数据源
                for (var service_name of service_name_set) {
                    this.serviceNames.push(service_name);
                }
                for (var index = 0; index < this.serviceNames.length; index++) {
                    var cur_service_name = this.serviceNames[index];
                    var node_list = [];
                    for (var j = 0; j < this.tableData.length; j++) {
                        if (cur_service_name == this.tableData[j].service_name) {
                            node_list.push(this.tableData[j]);
                        }
                    }
                    this.service_to_host[cur_service_name] = node_list;
                    if(this.service_to_host[cur_service_name] != "undefined" && this.service_to_host[cur_service_name] != null) {
                        this.service_to_host_count = this.service_to_host[cur_service_name].length;
                    }
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
    }
  }
}
</script>
