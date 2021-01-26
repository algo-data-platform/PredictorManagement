<template>
  <div class="app-container" style="margin: 5px 0;">
    <div style="margin-top: 15px;">
      <el-row :gutter="20">
        <el-col :span="18">
          <div class="filter-item el-input el-input--medium" style="width: 200px;">
            <el-input  placeholder="输入搜索内容" type="medium" v-model="search" prefix-icon="el-icon-search" maxlength="120">
            </el-input>
          </div>
          
        </el-col>
        <el-col :span="6">
          <div align="right">
            <el-button type="primary" icon="el-icon-tree" @click="begin_migrate()">集群调整</el-button>
          </div>
        </el-col>
      </el-row>
      
      <div style="margin: 10px 0;">
      </div>
      <el-table
        :data="service_show_data"
        style="width: 100%"
        :default-sort = "{ prop: 'sid', order: 'descending'}"
        element-loading-text="Loading"
        border
        fit
        highlight-current-row>

        <el-table-column  prop="sid" align="center" label="ID" sortable width="70px" >
        </el-table-column>

        <el-table-column prop="service_name" label="服务名" sortable fit="true" align="center">
        </el-table-column>

        <el-table-column prop="host_num" label="总数" sortable fit="true" align="center">
          <template slot-scope="props">
           {{props.row.host_num}}
          </template>
        </el-table-column>

        <el-table-column prop="idc_host_nums" label="IDC统计" sortable align="center">
           <template slot-scope="props">
                <table style="width:100%;text-align:center;">
                    <tr  v-for="item in props.row.idc_host_nums">
                        <td style="width:50%;text-align:center;">{{item.idc}}</td>
                        <td style="text-align:center;"><el-tag type="success">{{item.host_num}}</el-tag></td>
                    </tr>
                </table>
           </template>
        </el-table-column>
        <el-table-column prop="cpu_host_nums" label="CPU核数统计" sortable align="center">
           <template slot-scope="props">
                <table style="width:100%;text-align:center;">
                    <tr  v-for="item in props.row.cpu_host_nums">
                        <td style="width:50%;text-align:center;">{{item.core_num}}</td>
                        <td style="text-align:center;"><el-tag type="success">{{item.host_num}}</el-tag></td>
                    </tr>
                </table>
           </template>
        </el-table-column>
        <el-table-column prop="idc_host_nums" label="Mem统计" sortable align="center">
           <template slot-scope="props">
                <table style="width:100%;text-align:center;">
                    <tr  v-for="item in props.row.mem_host_nums">
                        <td style="width:50%;text-align:center;">{{item.total_mem}}</td>
                        <td style="text-align:center;"><el-tag type="success">{{item.host_num}}</el-tag></td>
                    </tr>
                </table>
           </template>
        </el-table-column>
        </el-table>

        <el-dialog title="集群调整" :visible.sync="dialogMigrateVisible" width="60%">
            <el-steps :active="active" finish-status="success" >
              <el-step title="选择机器"></el-step>
              <el-step title="预览"></el-step>
            </el-steps>
            <el-form :model="migrateForm" ref="migrateForm" v-show="step1Visible">
                <el-form-item label="From Service:" >
                    <el-select placeholder="请选择From Service" filterable v-model="migrateForm.from_service" @change="doSelectFromService" style="width:100%">
                    <el-option
                        v-for="item in from_service_list"
                        :key="item.value"
                        :label="item.label"
                        :value="item.value">
                    </el-option>
                    </el-select>
                </el-form-item>
                <el-form-item label="To Service:" >
                    <el-select placeholder="请选择To Service" filterable v-model="migrateForm.to_service" @change="doSelectToService" style="width:100%">
                    <el-option
                        v-for="item in to_service_list"
                        :key="item.value"
                        :label="item.label"
                        :value="item.value"
                        :disabled="item.disabled">
                    </el-option>
                    </el-select>
                </el-form-item>
                <el-row>
                  <el-col :span="8">
                    <el-form-item label="迁移台数:" >
                        <el-input-number v-model="migrateForm.num" :min="0" :max="1000" label="填写要迁移的机器数量"></el-input-number>
                    </el-form-item>
                  </el-col>
                  <el-col :span="16">
                  </el-col>
                </el-row>
            </el-form>
            <div v-show="step2Visible">
              <el-tag effect="dark" style="margin-top:10px;margin-bottom:10px;">Found: {{preview_hosts_count}}</el-tag>
              <div class="infinite-list-wrapper" style="overflow:auto;height:350px;">
                <el-table
                  :data="previewHosts"
                  style="width: 100%"
                  :default-sort = "{ prop: 'hsid', order: 'descending'}"
                  element-loading-text="Loading"
                  border
                  fit
                  highlight-current-row
                  >

                  <el-table-column  prop="hid" align="center" label="hid" sortable width="70px" >
                  </el-table-column>
                  <el-table-column  prop="ip" align="center" label="IP" sortable fit="true" >
                  </el-table-column>

                  <el-table-column prop="service_names" label="From Service"  fit="true" align="center">
                  </el-table-column>
                  <el-table-column prop="to_service_name" label="To Service"  fit="true" align="center">
                    {{select_to_service_name}}
                  </el-table-column>
                  <el-table-column prop="CoreNum" label="核数" sortable fit="true" align="center">
                  </el-table-column>
                  <el-table-column prop="TotalMem" label="内存" sortable fit="true" align="center">
                  </el-table-column>
                  <el-table-column prop="operator" label="操作" fit="true" align="center">
                    <template slot-scope="props">
                      <el-button  @click="deleteHosts(props.$index)" size="mini" type="danger"> 删除
                      </el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </div>
            <div slot="footer" class="dialog-footer">
                <el-button @click="dialogMigrateVisible = false">取 消</el-button>
                <el-button type="primary" v-if="active==0 || active==1" @click="preview($event)">下一步 && 预览</el-button>
                <el-button type="primary" v-if="active==2" @click="lastStep($event)">上一步</el-button>
                <el-button type="primary" v-if="active==2" @click="doMigrate($event)" :disabled="disabledMigrate">确定迁移</el-button>
            </div>
        </el-dialog>
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
        inputData: '',
        list: [],
        host_list: [],
        search: '',
        service_all_data: [],

        select_from_service: '',
        select_to_service: '',
        from_service_list: [],
        to_service_list: [],
        dialogMigrateVisible: false,
        
        migrateForm : {
          from_service: '',
          to_service : '',
          num: '',
        },
        onChange: '',
        onSearch: '',
        active: 0,
        step1Visible: true,
        step2Visible: false,
        previewHosts:[],
        disabledMigrate: false,
        preview_hosts_count: 0,
        service_count_map: {},
        to_service_map: {},
        select_to_service_name: '',
    };
  },
  created () {
      this.loadList()
      
  },
  methods: {
    loadList: function() {
      axios.get('/migrate/service_stats').then(response => {
        var _this = this
        var res_data = response.data;
        if (res_data.code != 0) {
            _this.$message({
                type: 'error',
                message: response.data.msg
            });
            return ;
        } else {
            var memBuckets = new Array();
            for (var i=0; i<=20; i++) {
              memBuckets.push(16*i)
            }
            var halfSearch = function(target) {
              var low = 0;
              var high = memBuckets.length -1;
              while(low <= high) {
                var mid= parseInt((low + high) / 2);
                if (memBuckets[mid] > target) {
                  high = mid - 1;
                } else if (memBuckets[mid] < target) {
                  low=mid + 1;
                } else {
                  return mid
                }
              }
              if(target>memBuckets[high]){
                return high + 1;
              } else {
                return high;
              }
            }
            var data_list = res_data.data;
            _this.service_all_data = [];
            for (var i = 0; i < data_list.length; i++) {
                // 内存分桶
                var mem_bucket_map = new Map();
                for(var idx=0;idx < data_list[i].mem_host_nums.length;idx++){
                  var total_mem = data_list[i].mem_host_nums[idx].total_mem;
                  var host_num = data_list[i].mem_host_nums[idx].host_num;
                  // ex= halfSearch(total_mem)
                  var ex = parseInt(total_mem/16);
                  var mod = total_mem % 16;
                  if(mod > 0){
                    ex = ex + 1
                  }
                  if(mem_bucket_map.has(ex)){
                    mem_bucket_map.set(ex,mem_bucket_map.get(ex) + host_num)
                  } else {
                    mem_bucket_map.set(ex,host_num)
                  }
                }
                var mem_buckets = [];
                for(var bucket of mem_bucket_map) {
                    var ex=bucket[0];
                    var num = bucket[1];
                    var mem_bucket_item = {
                      total_mem : ex * 16 +"G",
                      host_num : num,
                    }
                    mem_buckets.push(mem_bucket_item);
                }
                
                var item = {
                    sid: data_list[i].sid,
                    service_name : data_list[i].service_name,
                    host_num: data_list[i].host_num,
                    idc_host_nums: data_list[i].idc_host_nums,
                    cpu_host_nums: data_list[i].cpu_host_nums,
                    mem_host_nums: mem_buckets
                }
                _this.service_all_data.push(item);
                _this.service_count_map[data_list[i].sid]=data_list[i].host_num
            }
            _this.loadFromService()
            //_this.loadToService()
        }
         
      })
    },
    loadFromService: function() {
      var _this = this
      axios.get('/migrate/get_from_services').then(response => {
          var response_data = response.data;
          var data_list = response_data.data;
          _this.from_service_list = [];
          _this.to_service_list = [];
          for (var i = 0; i < data_list.length; i++) {
              var sids = data_list[i].sids.join(",")
              var names = ""
              for(var idx=0; idx<data_list[i].sids.length; idx++) {
                names += data_list[i].names[idx]
                if (idx < data_list[i].sids.length -1) {
                  names += " + "
                }
              }
              names += " (" + data_list[i].host_num + "台)"
              _this.from_service_list.push({value: sids, label: names});

              _this.to_service_list.push({value: sids, label: names, disabled: false});
              var to_services = data_list[i].names.join(",")
              _this.to_service_map[sids]=to_services
          }
      });
    },
    loadToService: function() {
      var _this = this
      axios.get('/migrate/get_to_services').then(response => {
          var response_data = response.data;
          var data_list = response_data.data;
          _this.to_service_list = [];
          for (var i = 0; i < data_list.length; i++) {
              var sids = data_list[i].sids.join(",")
              var to_services = data_list[i].names.join(",")
              var names = ""
              for(var idx=0; idx<data_list[i].sids.length; idx++) {
                names += data_list[i].names[idx] + " (" + _this.service_count_map[data_list[i].sids[idx]] + "台)"
                if (idx < data_list[i].sids.length -1) {
                  names += " + "
                }
              }
              _this.to_service_list.push({value: sids, label: names, disabled: false});
              _this.to_service_map[sids]=to_services
          }
      });
    },
    begin_migrate: function() {
        this.dialogMigrateVisible=true;
        this.clearDialog()
    },
    doSelectFromService: function(from_service) {
        this.select_from_service = from_service
        for(var i=0; i<this.to_service_list.length; i++){
          if(this.to_service_list[i].value == from_service) {
            this.to_service_list[i].disabled=true
          } else {
            this.to_service_list[i].disabled=false
          }
        }
    },
    doSelectToService: function(to_service){
        this.select_to_service = to_service
        if(this.to_service_map.hasOwnProperty(to_service)){
          this.select_to_service_name = this.to_service_map[to_service]
        } else {
          this.select_to_service_name = '';
        }
    },
    lastStep: function(event) {
      event.preventDefault();
      this.active = 1;
      this.step1Visible=true
      this.step2Visible=false
    },
    preview: function(event) {
      var _this = this;
      event.preventDefault();
      var from_service = this.migrateForm.from_service;
      var to_service = this.migrateForm.to_service;
      var num = this.migrateForm.num;
      axios.get('/migrate/preview', {
          params : {
              from_service : from_service,
              to_service: to_service,
              num: num,
          }
      }).then(function (response) {
          if (response.data.code == 0) {
            // todo 写入预览机器列表
            var data_list=response.data.data
            _this.previewHosts = [];
            for (var i = 0; i < data_list.length; i++) {
                var item = {
                    hsid: data_list[i].hsid,
                    hid: data_list[i].hid,
                    ip: data_list[i].ip,
                    service_names : data_list[i].service_names.join(","),
                    idc : data_list[i].idc,
                    CoreNum : data_list[i].resource_info.CoreNum,
                    TotalMem : data_list[i].resource_info.TotalMem + "G",
                }
                _this.previewHosts.push(item);
            }
            if(_this.previewHosts.length==0) {
              _this.disabledMigrate=true
            } else {
              _this.disabledMigrate=false
            } 
            _this.preview_hosts_count=_this.previewHosts.length
            
            // 进入下一步
            _this.active = 2
            _this.step1Visible=false
            _this.step2Visible=true
            _this.$message({
              type: 'success',
              message: '获取预览机器成功'
            });
          } else {
            _this.active = 0
            _this.$message({
              type: 'error',
              message: '预览失败, ' + response.data.msg
            });
            return;
          }
      }).catch(function (error) {
          _this.active = 0
          _this.$message({
              type: 'error',
              message: '预览失败, 网络异常'
          });
          return ;
      });
      
    },
    doMigrate: function(event) {
      var _this = this;
      event.preventDefault();
      var from_service = this.migrateForm.from_service;
      var to_service = this.migrateForm.to_service;
      var to_migrate_hids_list = [];
      for (var i = 0; i < this.previewHosts.length; i++) {
        to_migrate_hids_list.push(this.previewHosts[i].hid)
      }
      var to_migrate_hids = to_migrate_hids_list.join(",")
      axios.get('/migrate/do_migrate', {
          params : {
              from_service : from_service,
              to_service: to_service,
              to_migrate_hids: to_migrate_hids,
          }
      }).then(function (response) {
          if (response.data.code == 0) {
            _this.previewHosts = [];
           
            _this.dialogMigrateVisible=false
            _this.loadList();
            _this.$message({
              type: 'success',
              message: response.data.msg
            });
          } else {
            _this.$message({
              type: 'error',
              message: '迁移失败, ' + response.data.msg
            });
            return;
          }
      }).catch(function (error) {
          _this.$message({
              type: 'error',
              message: '迁移失败, 网络异常'
          });
          return ;
      });
      
    },
    clearDialog: function() {
      this.active = 0
      this.step1Visible=true
      this.step2Visible=false
      this.migrateForm = {
        from_service:"",
        to_service:"",
        num:0,
      }
      this.previewHosts = []
    },
    deleteHosts: function(index) {
      this.previewHosts.splice(index, 1);
      if(this.previewHosts.length==0) {
        this.disabledMigrate=true
      } else {
        this.disabledMigrate=false
      }
    }
  }, computed: {
        service_show_data:function () {
          var search = this.search;
          if(search) {
              return this.service_all_data.filter(function (service_infos) {
                  return Object.keys(service_infos).some(function (key) {
                      return String(service_infos[key]).toLocaleLowerCase().indexOf(search) > -1
                  })
              })
          }
          return this.service_all_data
      }
    },
}
</script>
