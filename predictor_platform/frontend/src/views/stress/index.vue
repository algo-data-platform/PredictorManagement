<template>
  <div class="app-container" style="margin: 5px 0;">
    <div style="margin-top: 15px;">
      <el-row :gutter="20">
        <el-col :span="18">
          <div class="filter-item el-input el-input--medium" style="width: 200px;">
            <el-input  placeholder="输入搜索内容" type="medium" v-model="search" prefix-icon="el-icon-search" maxlength="120">
            </el-input>
          </div>
    
          <label style="font-size: 14px;color: #606266;line-height: 40px;padding: 0 12px 0 0;margin-left:20px">压测状态：</label>
          <el-select placeholder="压测状态" v-model="search_is_enable" @change="doSetEnable" >
              <el-option
              v-for="item in status_list"
              :key="item.value"
              :label="item.label"
              :value="item.value">
            </el-option>
          </el-select>
          
        </el-col>
        <el-col :span="6">
          <div align="right">
            <el-button type="primary" icon="el-icon-edit" @click="dialogAddStressVisible = true">创建压测任务</el-button>
          </div>
        </el-col>
      </el-row>
      
      <div style="margin: 10px 0;">
      </div>
      <el-table
        :data="stress_show_data"
        style="width: 100%"
        :default-sort = "{ prop: 'id', order: 'descending'}"
        element-loading-text="Loading"
        border
        fit
        highlight-current-row>

        <el-table-column  prop="id" align="center" label="ID" sortable fit="true" >
        </el-table-column>

        <el-table-column prop="ip" label="IP" sortable fit="true" align="center">
        </el-table-column>

        <el-table-column prop="model_names" label="压测模型 -> QPS" sortable width="500px" align="center">
          <template slot-scope="props">
            <span  v-html="props.row.model_names_qps_table"/>
          </template>
        </el-table-column>
        <el-table-column  prop="is_enable" align="center" label="状态" sortable fit="true" >
          <template slot-scope="props">
            <el-tag v-if="props.row.is_enable == 1 " type="success">已开启</el-tag>
            <el-tag v-if="props.row.is_enable == 0 " type="info">已停止</el-tag>
          </template>
        </el-table-column>
        <el-table-column  prop="create_time" align="center" label="创建时间" sortable fit="true" >
        </el-table-column>
        <el-table-column prop="operator" label="操作" fit="true" align="center">
              <template slot-scope="props">
                <el-button v-if="props.row.is_enable == 1 " @click="editStress(props.row)" size="mini" type="primary" icon="el-icon-edit" style="margin-bottom:5px">QPS
                </el-button>
                <br/>
                <el-button v-if="props.row.is_enable == 1 " @click="disableStress(props.row.id)" size="mini" type="danger"> <svg-icon icon-class="stop" />&nbsp;&nbsp;停止
                </el-button>
                <el-button v-if="props.row.is_enable == 0 " @click="reEnableStress(props.row.id)" size="mini" type="success"> <svg-icon icon-class="redo" />&nbsp;&nbsp;重新开始
                </el-button>
              </template>
            </el-table-column>
      </el-table>
      <el-dialog title="创建压测任务" :visible.sync="dialogAddStressVisible">
        <el-form :model="stressForm" ref="stressForm">
          <el-form-item label="压测机器:" >
            <el-select placeholder="请选择" filterable v-model="stressForm.host" @change="doSelectHost" style="width:100%">
              <el-option
                v-for="item in all_host_ip"
                :key="item.value"
                :label="item.label"
                :value="item.value">
              </el-option>
            </el-select>
          </el-form-item>
          <div class="moreRules">
            <div class="moreRulesIn" v-for="(item, index) in stressForm.moreModelObject" :key="item.key">
              <el-row>
                <el-col :span="16" style="padding-right:10px">
                  <el-form-item class="rules" label="选择压测模型:" :prop="'moreModelObject.' + index +'.model'" >
                    <el-select placeholder="请选择model" filterable v-model="item.model" @change="doSelectModel" style="width:100%">
                      <el-option
                        v-for="item in all_models"
                        :key="item.value"
                        :label="item.label"
                        :value="item.value">
                      </el-option>
                    </el-select>
                  </el-form-item>
                </el-col>
                <el-col :span="5">
                  <el-form-item class="rules" label="qps:" :prop="'moreModelObject.'+ index +'.qps'" >
                    <el-input v-model="item.qps" placeholder="请输入压测qps" :disabled="isReadonly" class="el-select_box"></el-input>
                  </el-form-item>
                </el-col>
                <el-col :span="3">
                    <el-button  v-if="index > 0" type="primary" @click="deleteModel(item, index)" :disabled="isReadonly" style="margin-top:40px;margin-left:28px"><i class="el-icon-minus" /></el-button>
                </el-col>
              </el-row>
            </div>
          </div>
        
          <el-form-item v-show="!isRead">
            <el-button type="text"  @click="addModel" :disabled="isReadonly"><i class="el-icon-plus"/></i>添加压测模型</el-button>
          </el-form-item>
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button @click="dialogAddStressVisible = false">取 消</el-button>
          <el-button type="primary" @click="submitStressData($event)">确 定</el-button>
        </div>
      </el-dialog>
      <el-dialog title="编辑压测任务" :visible.sync="dialogEditStressVisible">
        <el-form :model="stressForm" ref="stressForm">
          <el-form-item label="压测机器:" >
            <el-select placeholder="请选择" disabled v-model="editStressForm.host" @change="doSelectHost" style="width:100%" readonly="readonly">
              <el-option
                v-for="item in all_host_ip"
                :key="item.value"
                :label="item.label"
                :value="item.value">
              </el-option>
            </el-select>
          </el-form-item>
          <div class="moreRules">
            <div class="moreRulesIn" v-for="(item, index) in editStressForm.moreModelObject" :key="item.key">
              <el-row>
                <el-col :span="16" style="padding-right:10px">
                  <el-form-item class="rules" label="选择压测模型:" :prop="'moreModelObject.' + index +'.model'" >
                    <el-select placeholder="请选择model" disabled v-model="item.model" @change="doSelectModel" style="width:100%" readonly="readonly">
                      <el-option
                        v-for="item in all_models"
                        :key="item.value"
                        :label="item.label"
                        :value="item.value">
                      </el-option>
                    </el-select>
                  </el-form-item>
                </el-col>
                <el-col :span="5">
                  <el-form-item class="rules" label="qps:" :prop="'moreModelObject.'+ index +'.qps'" >
                    <el-input v-model="item.qps" placeholder="请输入压测qps" :disabled="isReadonly" class="el-select_box"></el-input>
                  </el-form-item>
                </el-col>
                <el-col :span="3">
                </el-col>
              </el-row>
            </div>
          </div>
        
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button @click="dialogEditStressVisible = false">取 消</el-button>
          <el-button type="primary" @click="submitEditStressData($event)">确 定</el-button>
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
        stress_all_data: [],
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
        editStressForm : {
          stress_id: '',
          host: '',
          models : '',
          qps: '',
          moreModelObject: [{
            model: '',
            qps: ''
          }],
        },
        dialogAddStressVisible: false,
        dialogEditStressVisible: false,
        search_is_enable : '1',
        status_list: [
          {
            value: '1',
            label: '已开启'
          },{
            value: '0',
            label: '已停止'
          },{
            value: '2',
            label: '全部'
          },
        ],
        host_num: '',
        onChange: '',
        onSearch: '',
    };
  },
  created () {
      this.loadList(),
      this.loadHostList(),
      this.loadModels()
  },
  methods: {
    doSetEnable: function(is_enable){
      this.search_is_enable = is_enable;
      this.loadList()
    },
    loadList: function() {
      axios.get('/stress/list?is_enable='+ this.search_is_enable ).then(response => {
          var _this = this
          var res_data = response.data;
          if (res_data.code != 0) {
            _this.$message({
                type: 'error',
                message: '加载压测任务失败,' + response.data.msg
            });
            return ;
          }
          var list = res_data.data
          for (var i = 0; i < list.length; i++) {
            var model_names_qps_table = '<table style="width:100%;text-align:center;">';
            var item = list[i];
            for (var midx =0;midx < list[i].model_names.length; midx++ ) {
              var row = '<tr><td style="width:80%;text-align:center;">'+ list[i].model_names[midx] +'</td><td style="text-align:center;">'+ list[i].qps[midx] +'</td></tr>';
              model_names_qps_table += row;
            }
            model_names_qps_table += '</table>';
            list[i]["model_names_qps_table"] =model_names_qps_table;
          }
          _this.stress_all_data = list;
      })
    },
    loadHostList: function() {
      axios.get('/mysql/show?table=hosts').then(response => {
          var response_data = response.data;
          this.curTableName = response_data.Table_name;
          var data_list = response_data;
          var host_list = [];
          for (var i = 0; i < data_list.length; i++) {
              var id_host = data_list[i].ID.toString() + ":" + data_list[i].Ip;
              this.all_host_ip.push({value: data_list[i].ID, label: id_host});
          }
      });
    },
    loadModels:function() {
      axios.get('/mysql/show?table=models').then(response => {
          var response_data = response.data;
          var data_list = response_data;
          var model_names = [];
          for (var i = 0; i < data_list.length; i++) {
              var id_models = data_list[i].ID.toString() + ":" + data_list[i].Name;
              model_names.push({ value: data_list[i].ID, label: id_models});
          }
          this.all_models = model_names;
      }).catch(function (error) {
          console.log(error);
      });
    },
    disableStress: function(stress_id) {
      this.$confirm('确定要停止当前压测任务（ID：'+ stress_id +'）吗？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
      }).then(() => {
        axios.get('/stress/disable',{
          params : {
              stress_id: stress_id
          }
        }).then(response => {
            var _this = this
            var res_data = response.data;
            if (res_data.code != 0) {
              _this.$message({
                  type: 'error',
                  message: '停止压测任务失败,' + response.data.msg
              });
              return ;
            } else {
              _this.$message({
                  type: 'success',
                  message: '停止压测任务成功'
              });
            }
            _this.loadList();
        })
      });
      
    },
    reEnableStress: function(stress_id) {
      this.$confirm('确定要重新开启当前压测任务（ID：'+ stress_id +'）吗？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
      }).then(() => {
        axios.get('/stress/enable',{
          params : {
              stress_id: stress_id
          }
        }).then(response => {
            var _this = this
            var res_data = response.data;
            if (res_data.code != 0) {
              _this.$message({
                  type: 'error',
                  message: '重新开启压测任务失败,' + response.data.msg
              });
              return ;
            } else {
              _this.$message({
                  type: 'success',
                  message: '重新开启压测任务成功'
              });
            }
            _this.loadList();
        })
      });
      
    },
    submitStressData: function(event) {
      var _this = this;
      event.preventDefault();
      var hid = this.stressForm.host;
      var mids = [];
      var qps = [];
      console.log(this.stressForm.moreModelObject)
      for (var i =0; i < this.stressForm.moreModelObject.length; i++) {
        mids.push(this.stressForm.moreModelObject[i].model)
        qps.push(this.stressForm.moreModelObject[i].qps)
      }
      
      var midsStr = mids.join(",");
      var qpsStr = qps.join(",");
      axios.get('/stress/insert', {
          params : {
              hid : hid,
              mids: midsStr,
              qps: qpsStr,
          }
      }).then(function (response) {
          console.log(response);
          if (response.data.code == 0) {
            _this.$message({
              type: 'success',
              message: '压测任务创建成功'
            });
          } else {
            _this.$message({
              type: 'error',
              message: '压测任务创建失败, ' + response.data.msg
            });
            return;
          }
          _this.dialogAddStressVisible = false;
          _this.resetAddForm();
          _this.loadList();
      }).catch(function (error) {
          _this.$message({
              type: 'error',
              message: '创建压测任务失败, 网络异常'
          });
          return ;
      });
      
    },
    submitEditStressData: function(event) {
      var _this = this;
      event.preventDefault();
      var stress_id = this.editStressForm.stress_id;
      var qps = [];
      for (var i =0; i < this.editStressForm.moreModelObject.length; i++) {
        qps.push(this.editStressForm.moreModelObject[i].qps)
      }
      var qpsStr = qps.join(",");
      axios.get('/stress/save_qps', {
          params : {
              stress_id : stress_id,
              qps: qpsStr,
          }
      }).then(function (response) {
          console.log(response);
          if (response.data.code == 0) {
            _this.$message({
              type: 'success',
              message: '压测qps修改成功'
            });
          } else {
            _this.$message({
              type: 'error',
              message: '压测qps修改失败, ' + response.data.msg
            });
            return;
          }
          _this.dialogEditStressVisible = false;
          _this.loadList();
      }).catch(function (error) {
          _this.$message({
              type: 'error',
              message: '压测qps修改失败, 网络异常'
          });
          return ;
      });
      
    },
    resetAddForm:function() {
      this.stressForm.host = '';
      this.stressForm.moreModelObject = [{
        model: '',
        qps: ''
      }];
    },
    doSelectHost(host_value) {
        this.select_host = host_value;
    },
    doSelectModel(model_value) {
        this.select_model = model_value;
    },
    addModel() {
      this.stressForm.moreModelObject.push({
        model: '',
        qps: ''
      })
    },
    deleteModel(item, index) {
      this.index = this.stressForm.moreModelObject.indexOf(item)
      if (index !== -1) {
        this.stressForm.moreModelObject.splice(index, 1)
      }
    },
    editStress:function(item) {
      this.editStressForm.stress_id = item.id;
      this.editStressForm.host = item.hid;
      this.editStressForm.moreModelObject = [];
      for (var i=0;i < item.mids.length; i ++) {
        this.editStressForm.moreModelObject.push({
            model: item.mids[i],
            qps: item.qps[i]
        });
      }
      this.dialogEditStressVisible = true;
    }
  }, computed: {
        stress_show_data:function () {
          var search = this.search;
          if(search) {
              return this.stress_all_data.filter(function (stress_infos) {
                  return Object.keys(stress_infos).some(function (key) {
                      return String(stress_infos[key]).toLocaleLowerCase().indexOf(search) > -1
                  })
              })
          }
          return this.stress_all_data
      }
    },
}
</script>
