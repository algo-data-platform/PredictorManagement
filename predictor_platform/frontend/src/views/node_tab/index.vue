<template>
  <div class="app-container" style="margin: 5px 0;">
    <div id="app">
      <el-tag >all host:{{all_host}}</el-tag>
      <el-tag type="success">online host:{{online_host}}</el-tag>
      <el-tag type="danger">offline host:{{offline_host}}</el-tag>
    </div>
    <div style="margin-top: 15px;">
      <div id="filter-host">
        <el-input size="medium" style="width: 180px;" placeholder="输入搜索内容" v-model="search" prefix-icon="el-icon-search" maxlength="120">
        </el-input>
      </div>
    <div style="margin: 10px 0;">
    </div>
    <el-table
      :data="host_show_data"
      style="width: 100%"
      :default-sort = "{prop: 'host_ip', prop: 'service_name', order: 'descending'}"
      element-loading-text="Loading"
      border
      fit
      highlight-current-row>

      <el-table-column  prop="host_ip" align="center" label="Host" sortable width="125" height="250">
          <template slot-scope="scope">
            <el-button type="text">
              <a :href=" 'http://' + scope.row.host_ip + ':9528/get_service_model_info' "
                 target="_blank" class="buttonText">{{scope.row.host_ip}}
              </a>
            </el-button>
          </template>
      </el-table-column>

      <el-table-column prop="service_name" label="部署的服务" sortable width="210" align="center" fit="true"
        :filters="service_list"
        :filter-method="filterServiceName"
        filter-placement="bottom-end">
        <template slot-scope="scope">
          <el-tree :data="scope.row.names">
          </el-tree>
        </template>
      </el-table-column>

      <el-table-column prop="load_average" label="负载" sortable width="125" align="center" fit="true">
      </el-table-column>

      <el-table-column prop="core_num" align="center" label="核数" sortable width="100"
                       :filters="core_num_list"
                       :filter-method="filterCore"
                       filter-placement="bottom-end">
      </el-table-column>

      <el-table-column prop="node_status"  align="center" label="机器状态" sortable  width="120"
                       :filters="[{text: 'on', value: 'on'},{text: 'off', value: 'off'}]"
                       :filter-method="filterStatus"
                       filter-placement="bottom-end">
      </el-table-column>

      <el-table-column prop="memory_ratio" label="Avail/Total Memory" sortable width="180" align="center">
      </el-table-column>

      <el-table-column prop="disk_ratio" label="Avail/Total Disk" sortable width="180" align="center">
      </el-table-column>

      <el-table-column prop="cluster" label="机房" sortable width="100" align="center"
                       :filters="cluster_list"
                       :filter-method="filterCluster"
                       filter-placement="bottom-end">
      </el-table-column>
    </el-table>
    </div>
  </div>

</template>

<script>
import axios from 'axios'

export default {
  filters: {
    statusFilter(status) {
      const statusMap = {
        published: 'success',
        draft: 'gray',
        deleted: 'danger'
      }
      return statusMap[status]
    }
  },
  data() {
    return {
        all_host: 0,
        service_list: [],
        cluster_list:[],
        core_num_list:[],
        online_host: 0,
        offline_host: 0,
        host_ip: '',
        inputData: '',
        list: [],
        host_list: [],
        search: '',
        host_all_data: [],
        host_num: '',
        onChange: '',
        onSearch: '',
    };
  },
  created () {
      axios.get('/node_infos').then(response => {
          var res_data = response.data;
          this.list.unshift();
          let cluster_set = new Set();
          let service_set = new Set();
          let core_num_set = new Set();
          for (var index = 0; index < res_data.length; index++) {
              var load_average_all = res_data[index].ResourceInfo.LoadAverage.Average1 + "/"
                  + res_data[index].ResourceInfo.LoadAverage.Average5 + "/"
                  + res_data[index].ResourceInfo.LoadAverage.Average15;
              var node_is_available = res_data[index].ResourceInfo.NodeAvail;
              var status_info = res_data[index].StatusInfo;
              var cur_cluster = res_data[index].DataCenter;
              cluster_set.add(cur_cluster);
              var cur_core_num = res_data[index].ResourceInfo.CoreNum;
              core_num_set.add(cur_core_num);
              // service_name和所属models列表
              var service_to_models_list = [];
              var status_info_length = (status_info != null ? status_info.length : 0);
              for(var i = 0; i < status_info_length; i++) {
                  var service_name = status_info[i].service_name;
                  var model_records = status_info[i].model_records;
                  var model_list = [];
                  // models列表
                  for (var j = 0; j < model_records.length; j++) {
                      var cur_model = {
                          label : model_records[j].name
                      }
                      model_list.push(cur_model);
                  }
                  var service_to_models = {
                      label : service_name,
                      children : model_list
                  }
                  service_to_models_list.push(service_to_models);
                  service_set.add(status_info[i].service_name);
              }
              var item = {
                  host_ip: res_data[index].Host,
                  core_num: res_data[index].ResourceInfo.CoreNum,
                  memory_ratio : res_data[index].ResourceInfo.AvailMem + "/" +
                      res_data[index].ResourceInfo.TotalMem + " GB",
                  disk_ratio : res_data[index].ResourceInfo.AvailDisk + "/" +
                      res_data[index].ResourceInfo.TotalDisk + " GB",
                  load_average: load_average_all,
                  node_status: (node_is_available == '1') ? 'on' : 'off',
                  cluster: res_data[index].DataCenter,
                  names : service_to_models_list
              }
              if(node_is_available == "1") {
                  this.online_host++;
              }
              this.all_host++;
              this.list.push(item);
              this.host_list.push(res_data[index].Host);
          }
          this.offline_host = this.all_host - this.online_host;
          this.host_all_data = this.list;
          for(var each of cluster_set){
              var cluster_item = {
                  text: each,
                  value: each
              }
              this.cluster_list.push(cluster_item);
          }
          for(var each of service_set) {
              var service_item = {
                  text: each,
                  value: each
              }
              this.service_list.push(service_item);
          }
          for(var core of core_num_set) {
              var core_item = {
                  text: core,
                  value: core
              }
              this.core_num_list.push(core_item);
          }
      })
  },
  methods: {
      filterServiceName(value, row) {
          return row.names[0].label === value;
      },
      filterStatus(value, row) {
        return row.node_status === value;
      },
      filterCore(value, row) {
          return row.core_num === value;
      },
      filterCluster(value, row) {
          return row.cluster === value;
      },
  }, computed: {
        host_show_data:function () {
          var search = this.search;
          if(search) {
              return this.host_all_data.filter(function (host_infos) {
                  return Object.keys(host_infos).some(function (key) {
                      return String(host_infos[key]).toLocaleLowerCase().indexOf(search) > -1
                  })
              })
          }
          return this.host_all_data
      }
    },
}
</script>
