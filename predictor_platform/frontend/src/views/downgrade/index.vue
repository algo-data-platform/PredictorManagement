<template>
  <div class="app-container" style="margin: 5px 0;">
    <div style="margin-top: 15px;">
      <el-row :gutter="20">
        <el-col :span="18">
          <div class="filter-item el-input el-input--medium" style="width: 200px;">
            <el-input  placeholder="输入搜索内容" type="medium" v-model="search" prefix-icon="el-icon-search" maxlength="120">
            </el-input>
          </div>
          <el-button @click="refresh()" size="medium" type="primary"> <i class="el-icon-refresh" />&nbsp;&nbsp;刷新
        </el-col>
        <el-col :span="6">
          <div align="right">
           
          </div>
        </el-col>
      </el-row>
      
      <div style="margin: 10px 0;">
      </div>
      <el-table
        :data="service_show_data"
        style="width: 100%"
        :default-sort = "{ prop: 'id', order: 'descending'}"
        element-loading-text="Loading"
        border
        fit
        highlight-current-row>

        <el-table-column  prop="id" align="center" label="ID" sortable width="70px" >
        </el-table-column>

        <el-table-column prop="name" label="服务名" sortable fit="true" align="center">
        </el-table-column>

        <!--<el-table-column prop="desc" label="描述" sortable fit="true" align="center">
        </el-table-column>-->

        <el-table-column prop="prom_percent" label="降级百分比（来自prometheus）" sortable align="center">
           <template slot-scope="props">
            <el-progress  :percentage="props.row.prom_percent"></el-progress>
           </template>
        </el-table-column>
        <el-table-column  prop="set_service_percent" align="center" label="设置降级百分比" width="300px" sortable  >
          <template slot-scope="props">
            <div class="block">
              <el-slider
                v-model="props.row.set_service_percent"
                show-input>
              </el-slider>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="operator" label="操作" fit="true" align="center">
              <template slot-scope="props">
                <el-button  @click="startDowngrade(props.row)" size="mini" type="primary" style="margin-bottom:5px"><svg-icon icon-class="play-one" />&nbsp;&nbsp;开启
                </el-button>
               
                <el-button  @click="resetDowngrade(props.row)" size="mini" type="success"> <svg-icon icon-class="redo" />&nbsp;&nbsp;重置
                </el-button>
              </template>
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
        inputData: '',
        list: [],
        host_list: [],
        search: '',
        service_all_data: [],

        select_host: '',
        select_model: '',
        all_host_ip:[],
        all_models: [],
        stressForm : {
          host: '',
          models : '',
          qps: '',
          moreModelObject: [{
            model: '',
            qps: ''
          }],
        },
        onChange: '',
        onSearch: '',
    };
  },
  created () {
      this.loadList()
  },
  methods: {
    loadList: function() {
      axios.get('/mysql/show?table=services').then(response => {
          var _this = this
          var data_list = response.data;
          _this.service_all_data = [];
          for (var i = 0; i < data_list.length; i++) {
              var item = {
                  id: data_list[i].ID,
                  name: data_list[i].Name,
                  desc: data_list[i].Desc,
                  prom_percent: "0",
                  set_service_percent: 20,
              }
              _this.service_all_data.push(item);
          }
          // 获取prom_percent
          _this.loadPromPercent();
      })
    },
    loadPromPercent: function() {
      axios.get('/downgrade/get_prometheus_downgrade_percent').then(response => {
          var _this = this
          var res_data = response.data;
          if (res_data.code != 0) {
            _this.$message({
                type: 'error',
                message: '获取prometheus降级百分比失败,' + response.data.msg
            });
            return ;
          } else {
            _this.$message({
                type: 'success',
                message: '获取prometheus降级百分比成功'
            });
          }
          for (var i=0; i < _this.service_all_data.length; i++) {
            var service_data = _this.service_all_data[i]
            if (res_data.data.hasOwnProperty(service_data.name)){
              _this.service_all_data[i].prom_percent=res_data.data[service_data.name]
            }
          }
      });
    },
    
    startDowngrade: function(row) {
      this.$confirm('确定要开启对（'+ row.name +'）服务开启'+ row.set_service_percent +'%降级吗？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
      }).then(() => {
        axios.get('/downgrade/set_by_service',{
          params : {
              sid: row.id,
              percent: row.set_service_percent,
          }
        }).then(response => {
            var _this = this
            var res_data = response.data;
            if (res_data.code != 0) {
              _this.$message({
                  type: 'error',
                  message: response.data.msg
              });
              return ;
            } else {
              _this.$message({
                  type: 'success',
                  message: response.data.msg
              });
            }
           
        })
      });
      
    },
    resetDowngrade: function(row) {
        this.$confirm('确定要对（'+ row.name +'）服务重置降级吗？', '提示', {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
        }).then(() => {
          axios.get('/downgrade/reset_by_service',{
            params : {
                sid: row.id
            }
          }).then(response => {
              var _this = this
              var res_data = response.data;
              if (res_data.code != 0) {
                _this.$message({
                    type: 'error',
                    message: response.data.msg
                });
                return ;
              } else {
                row.set_service_percent = 0;
                _this.$message({
                    type: 'success',
                    message: response.data.msg
                });
              }
              
          })
        });
        
      },
      refresh: function() {
        this.loadPromPercent()
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
