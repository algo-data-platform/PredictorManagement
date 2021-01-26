<template>
  <div class="cls-mysql">
    <div align="left" id="db">
      <el-tag style="padding-left: 10px" effect="dark">database:{{database}}</el-tag>
    </div>
    <el-tabs v-model="tabPosition" type="board-card" @tab-click="handleClick">
      <el-tab-pane label="hosts" name="hosts">
        <div id="app">
          <div class="clsmsg">
            <el-tag effect="dark">机器数:{{host_count}}</el-tag>
            <el-input style="margin-left: 30px; width: 160px;" align="center" size="medium" placeholder="输入搜索内容" v-model="host_info_search" prefix-icon="el-icon-search" maxlength="160">
            </el-input>
          </div>
        </div>
        <div align="right">
        <el-button type="primary" icon="el-icon-plus" @click="dialogBatchAddHostsVisible = true">批量添加机器</el-button>
          <el-button type="primary" icon="el-icon-edit" @click="dialogAddHostVisible = true">Add Data</el-button>          
        </div>
        <el-table :data="host_infos.filter(data => !host_info_search || data.ip.toLowerCase().includes(host_info_search.toLowerCase())
                                           || data.data_center.toLowerCase().includes(host_info_search.toLowerCase())
                                           || data.desc.toLowerCase().includes(host_info_search.toLowerCase())
                                           || data.created_at.toLowerCase().includes(host_info_search.toLowerCase())
                                           || data.updated_at.toLowerCase().includes(host_info_search.toLowerCase()))">
          <el-table-column v-for="col in cols_host"
                           sortable :prop="col.prop"
                           :label="col.label"
                           :filters="col.filter"
                           :filter-method="col.filter && filterHandler">
          </el-table-column>
          <el-table-column prop="operator" label="操作" width="90" align="center">
            <template slot-scope="props">
              <el-button @click="dialogUpdateHostVisible = true, updateFormHost.ip = props.row.ip,
                         updateFormHost.desc = props.row.desc,
                         updateFormHost.data_center = props.row.data_center"
                         size="mini" type="primary" icon="el-icon-edit" circle>
              </el-button>
              <el-button @click="deleteHostOperator(props.row)" size="mini" type="danger" icon="el-icon-delete" circle>
              </el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-dialog title="增加host数据" :visible.sync="dialogAddHostVisible" :modal-append-to-body='false'>
          <el-form :model="addFormHost" ref="addFormHost">
            <el-form-item label="ip" :label-width="formLabelWidth" required>
              <el-input v-model="addFormHost.ip" auto-complete="off"></el-input>
            </el-form-item>
            <el-form-item label="data_center" :label-width="formLabelWidth">
              <el-input v-model="addFormHost.data_center" autocomplete="off"></el-input>
            </el-form-item>
            <el-form-item label="desc" :label-width="formLabelWidth">
              <el-input v-model="addFormHost.desc" autocomplete="off"></el-input>
            </el-form-item>
          </el-form>
          <div slot="footer" class="dialog-footer">
            <el-button @click="dialogAddHostVisible = false">取 消</el-button>
            <el-button type="primary" @click="submitHostData($event)">确 定</el-button>
          </div>
        </el-dialog>
        <el-dialog title="修改数据" :visible.sync="dialogUpdateHostVisible" :modal-append-to-body='false'>
          <el-form :model="updateFormHost" ref="updateFormHost">
            <el-form-item label="ip" :label-width="formLabelWidth">
              <el-input v-model="updateFormHost.ip" auto-complete="off" readonly="readonly">
              </el-input>
            </el-form-item>
            <el-form-item label="data_center" :label-width="formLabelWidth">
              <el-input v-model="updateFormHost.data_center" autocomplete="off"></el-input>
            </el-form-item>
            <el-form-item label="desc" :label-width="formLabelWidth">
              <el-input v-model="updateFormHost.desc" autocomplete="off"></el-input>
            </el-form-item>
          </el-form>
          <div slot="footer" class="dialog-footer">
            <el-button @click="dialogUpdateHostVisible = false">取 消</el-button>
            <el-button type="primary" @click="updateHostData($event)">确 定</el-button>
          </div>
        </el-dialog>
        <el-dialog title="批量添加机器" :visible.sync="dialogBatchAddHostsVisible" :modal-append-to-body='false'>
          <el-form :model="addBatchHostsForm" ref="addBatchHostsForm">
            <el-form-item label="机器IP" :label-width="formLabelWidth">
              <el-input type="textarea" :rows="10" v-model="addBatchHostsForm.host_ips" placeholder="请输入ip列表，换行间隔，一行一个" >
              </el-input>
            </el-form-item>
          </el-form>
          <div slot="footer" class="dialog-footer">
            <el-button @click="dialogBatchAddHostsVisible = false">取 消</el-button>
            <el-button type="primary" @click="batchAddHosts($event)">确 定</el-button>
          </div>
        </el-dialog>
        
      </el-tab-pane>

      <el-tab-pane label="services" name="services">
        <el-input style="margin-left: 20px; width: 180px;" align="center" size="medium" placeholder="搜索内容" v-model="service_info_search" prefix-icon="el-icon-search" maxlength="180">
        </el-input>
        <div align="right">
          <el-button type="primary" icon="el-icon-edit" @click="dialogAddServiceVisible = true">Add Data</el-button>
        </div>

        <el-table :data="all_list.filter(data => !service_info_search || data.name.toLowerCase().includes(service_info_search.toLowerCase())
                                        || data.desc.toLowerCase().includes(service_info_search.toLowerCase())
                                        || data.created_at.toLowerCase().includes(service_info_search.toLowerCase())
                                        || data.updated_at.toLowerCase().includes(service_info_search.toLowerCase())
                                        )">
          <el-table-column v-for="col in cols_service"
                           sortable :prop="col.prop"
                           :label="col.label">
          </el-table-column>
          <el-table-column prop="balance_status" label="自动权重（关<->开）" width="100" align="center">
            <template slot-scope="scope">
              <el-switch @change="changeLoadBalance($event, scope.row)"
                v-model="scope.row.balance_status" active-color="#13ce66" inactive-color="#ccc">
              </el-switch>
            </template>
          </el-table-column>

          <el-table-column prop="operator" label="操作" width="150" align="center">
            <template slot-scope="props">
              <el-tooltip class="item" effect="dark" content="重置权重" placement="top">
                <el-button @click="resetServiceWeight(props.row)"
                            size="mini" type="primary"  circle><svg-icon icon-class="redo" /></el-button>
                </el-button>
              </el-tooltip>
              <el-button @click="dialogUpdateServiceVisible = true, updateFormService.service_name = props.row.name,
                         updateFormService.desc = props.row.desc"
                         size="mini" type="primary" icon="el-icon-edit" circle>
              </el-button>
              <el-button @click="deleteServiceOperator(props.row)" size="mini" type="danger" icon="el-icon-delete" circle>
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-dialog title="增加service数据" :visible.sync="dialogAddServiceVisible" :modal-append-to-body='false'>
          <el-form :model="formService" ref="formService">
            <el-form-item label="service_name" :label-width="formLabelWidth" required>
              <el-input v-model="formService.service_name" auto-complete="off"></el-input>
            </el-form-item>
            <el-form-item label="desc" :label-width="formLabelWidth">
              <el-input v-model="formService.desc" autocomplete="off"></el-input>
            </el-form-item>
          </el-form>
          <div slot="footer" class="dialog-footer">
            <el-button @click="dialogAddServiceVisible = false">取 消</el-button>
            <el-button type="primary" @click="submitServiceData($event)">确 定</el-button>
          </div>
        </el-dialog>

        <el-dialog title="修改数据" :visible.sync="dialogUpdateServiceVisible">
          <el-form :model="updateFormService" ref="updateFormService">
            <el-form-item label="service name" :label-width="formLabelWidth">
              <el-input v-model="updateFormService.service_name" auto-complete="off" readonly="readonly">
              </el-input>
            </el-form-item>
            <el-form-item label="Description" :label-width="formLabelWidth">
              <el-input v-model="updateFormService.desc" autocomplete="off"></el-input>
            </el-form-item>
          </el-form>
          <div slot="footer" class="dialog-footer">
            <el-button @click="dialogUpdateServiceVisible = false">取 消</el-button>
            <el-button type="primary" @click="updateServiceData($event)">确 定</el-button>
          </div>
        </el-dialog>
      </el-tab-pane>

      <el-tab-pane label="configs" name="configs">
        <div align="right">
          <el-button type="primary" icon="el-icon-edit" @click="dialogAddConfigVisible = true">Add Data</el-button>
        </div>

        <el-table :data="config_list">
          <el-table-column v-for="col in cols_config"
                           sortable :prop="col.prop"
                           :label="col.label">
          </el-table-column>

          <el-table-column prop="operator" label="操作" width="90" align="center">
            <template slot-scope="props">
              <el-button @click="dialogUpdateConfigVisible = true, updateFormConfig.id = props.row.id,
                         updateFormConfig.desc = props.row.desc, updateFormConfig.config = props.row.config"
                         size="mini" type="primary" icon="el-icon-edit" circle>
              </el-button>
              <el-button @click="deleteConfigOperator(props.row)" size="mini" type="danger" icon="el-icon-delete" circle>
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-dialog title="增加config数据" :visible.sync="dialogAddConfigVisible" :modal-append-to-body='false'>
          <el-form :model="formConfig" ref="formConfig">
            <el-form-item label="Description" :label-width="formLabelWidth" required>
              <el-input v-model="formConfig.desc" autocomplete="off"></el-input>
            </el-form-item>
            <el-form-item label="JSON Config" :label-width="formLabelWidth" required>
              <el-input v-model="formConfig.config" auto-complete="off"></el-input>
            </el-form-item>
          </el-form>
          <div slot="footer" class="dialog-footer">
            <el-button @click="dialogAddConfigVisible = false">取 消</el-button>
            <el-button type="primary" @click="submitConfigData($event)">确 定</el-button>
          </div>
        </el-dialog>

        <el-dialog title="修改数据" :visible.sync="dialogUpdateConfigVisible">
          <el-form :model="updateFormConfig" ref="updateFormConfig">
            <el-form-item label="Description" :label-width="formLabelWidth">
              <el-input v-model="updateFormConfig.desc" autocomplete="off"></el-input>
            </el-form-item>
            <el-form-item label="JSON Config" :label-width="formLabelWidth">
              <el-input v-model="updateFormConfig.config" auto-complete="off">
              </el-input>
            </el-form-item>
          </el-form>
          <div slot="footer" class="dialog-footer">
            <el-button @click="dialogUpdateConfigVisible = false">取 消</el-button>
            <el-button type="primary" @click="updateConfigData($event)">确 定</el-button>
          </div>
        </el-dialog>
      </el-tab-pane>

      <el-tab-pane label="models" name="models">
        <div align="right">
          <el-button type="primary" icon="el-icon-edit" @click="dialogAddModelVisible = true">Add Data</el-button>
        </div>

        <el-table :data="all_model_list">
          <el-table-column v-for="col in cols_model"
                           sortable :prop="col.prop"
                           :label="col.label">
          </el-table-column>
          <el-table-column prop="operator" label="操作" width="90" align="center">
            <template slot-scope="props">
              <el-button @click="dialogUpdateModelVisible = true,
                         updateFormModel.name = props.row.name,
                         updateFormModel.path = props.row.path,
                         updateFormModel.desc = props.row.desc,
                         updateFormModel.extension = props.row.extension"
                         size="mini" type="primary" icon="el-icon-edit" circle>
              </el-button>
              <el-button @click="deleteModelOperator(props.row)" size="mini" type="danger" icon="el-icon-delete" circle>
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-dialog title="增加model数据" :visible.sync="dialogAddModelVisible" :modal-append-to-body='false'>
          <el-form :model="formModel" ref="formModel">
            <el-form-item label="model_name" :label-width="formLabelWidth" required>
              <el-input v-model="formModel.name" auto-complete="off"></el-input>
            </el-form-item>
            <el-form-item label="model_path" :label-width="formLabelWidth">
              <el-input v-model="formModel.path" autocomplete="off"></el-input>
            </el-form-item>
            <el-form-item label="desc" :label-width="formLabelWidth">
              <el-input v-model="formModel.desc" autocomplete="off"></el-input>
            </el-form-item>
            <el-form-item label="extension" :label-width="formLabelWidth">
              <el-input v-model="formModel.extension" autocomplete="off"></el-input>
            </el-form-item>
          </el-form>
          <div slot="footer" class="dialog-footer">
            <el-button @click="dialogAddModelVisible = false">取 消</el-button>
            <el-button type="primary" @click="submitModelData($event)">确 定</el-button>
          </div>
        </el-dialog>

        <el-dialog title="修改数据" :visible.sync="dialogUpdateModelVisible">
          <el-form :model="updateFormModel" ref="updateFormModel">
            <el-form-item label="model name" :label-width="formLabelWidth">
              <el-input v-model="updateFormModel.name" auto-complete="off" readonly="readonly">
              </el-input>
            </el-form-item>
            <el-form-item label="path" :label-width="formLabelWidth">
              <el-input v-model="updateFormModel.path" autocomplete="off"></el-input>
            </el-form-item>
            <el-form-item label="description" :label-width="formLabelWidth">
              <el-input v-model="updateFormModel.desc" autocomplete="off"></el-input>
            </el-form-item>
            <el-form-item label="extension" :label-width="formLabelWidth">
              <el-input v-model="updateFormModel.extension" autocomplete="off"></el-input>
            </el-form-item>
          </el-form>
          <div slot="footer" class="dialog-footer">
            <el-button @click="dialogUpdateModelVisible = false">取 消</el-button>
            <el-button type="primary" @click="updateModelDataFunc($event)">确 定</el-button>
          </div>
        </el-dialog>
      </el-tab-pane>

      <el-tab-pane label="host_services" name="host_services">
        <div align="right">
          <el-button type="primary" icon="el-icon-edit" @click="dialogAddHostServiceVisible = true">Add Data</el-button>
        </div>
        <el-table :data="host_service_list">
          <el-table-column v-for="col in cols_host_services"
                           sortable :prop="col.prop"
                           :label="col.label" :fit="true">
          </el-table-column>
          <el-table-column prop="operator" label="操作" width="90" align="center">
            <template slot-scope="props">
              <el-button @click="dialogUpdateHostServiceVisible = true, updateHostService = props.row,
                    modify_host_service.id = props.row.id,
                    modify_host_service.hid = props.row.hid,
                    modify_host_service.sid = props.row.sid,
                    modify_host_service.load_weight = props.row.load_weight"
                         size="mini" type="primary" icon="el-icon-edit" circle @click.native="updateDisableStatus()"
              >
              </el-button>
              <el-button @click="deleteHostServiceOperator(props.row)" size="mini" type="danger" icon="el-icon-delete" circle>
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-dialog title="增加host_services数据" :visible.sync="dialogAddHostServiceVisible">
          <el-select  multiple placeholder="请选择" v-model="default_ip" @change="doSelectHost">
            <el-option
              v-for="item in all_host_ip"
              :key="item.value"
              :label="item.label"
              :value="item.value">
            </el-option>
          </el-select>

          <el-select  placeholder="请选择" v-model="update_service" @change="doSelectService">
            <el-option
              v-for="item in all_services"
              :key="item.value"
              :label="item.label"
              :value="item.value">
            </el-option>
          </el-select>
          <br></br>
          <el-input placeholder="请输入权重" v-model="load_weight_input" clearable  @change="fillLoadWeight" style="width: 380px">
          </el-input>

          <div slot="footer" class="dialog-footer">
            <el-button @click="dialogAddHostServiceVisible = false">取 消</el-button>
            <el-button type="primary" @click="submitHostServicesData($event)">确 定</el-button>
          </div>
        </el-dialog>

        <el-dialog title="修改数据" :visible.sync="dialogUpdateHostServiceVisible" :modal-append-to-body="false">
          <el-form :model="modify_host_service" ref="modify_host_service">
            <el-form-item label="host" :label-width="formLabelWidth">
              <el-input v-model="modify_host_service.hid" auto-complete="off" readonly="readonly">
              </el-input>
            </el-form-item>
            <el-form-item label="service" :label-width="formLabelWidth">
              <el-select  placeholder="请选择" v-model="modify_host_service.sid" @change="doSelectService" style="width: 280px">
                <el-option
                  v-for="item in all_update_services"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                  :disabled="item.disabled">
                </el-option>
              </el-select>
            </el-form-item>
            <el-form-item label="load_weight" :label-width="formLabelWidth">
              <el-input v-model="modify_host_service.load_weight" @change="fillLoadWeight" autocomplete="off"></el-input>
            </el-form-item>
          </el-form>
          <div slot="footer" class="dialog-footer">
            <el-button @click="dialogUpdateHostServiceVisible = false">取 消</el-button>
            <el-button type="primary" @click="updateHostServiceData($event)">确 定</el-button>
          </div>
        </el-dialog>
      </el-tab-pane>

      <el-tab-pane label="service_configs" name="service_configs">
        <div align="right">
          <el-button type="primary" icon="el-icon-edit" @click="dialogAddServiceConfigVisible = true">Add Data</el-button>
        </div>
        <el-table :data="service_config_list">
          <el-table-column v-for="col in cols_service_configs"
                           sortable :prop="col.prop"
                           :label="col.label" :fit="true">
          </el-table-column>
          <el-table-column prop="operator" label="操作" width="90" align="center">
            <template slot-scope="props">
              <el-button @click="dialogUpdateServiceConfigVisible = true, updateServiceConfig = props.row,
                    modify_service_config.id = props.row.id,
                    modify_service_config.cid = props.row.cid,
                    modify_service_config.sid = props.row.sid,
                    modify_service_config.desc = props.row.desc"
                         size="mini" type="primary" icon="el-icon-edit" circle @click.native="ConfigDisableStatus()"
              >
              </el-button>
              <el-button @click="deleteServiceConfigOperator(props.row)" size="mini" type="danger" icon="el-icon-delete" circle>
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-dialog title="增加service_config数据" :visible.sync="dialogAddServiceConfigVisible">
          <el-select  placeholder="请选择" v-model="update_service_sc" @change="doSelectServiceForSC">
            <el-option
              v-for="item in all_services_sc"
              :key="item.value"
              :label="item.label"
              :value="item.value"
              :disabled="item.disabled">
            </el-option>
          </el-select>
          <el-select  placeholder="请选择" v-model="default_config" @change="doSelectConfigForSC">
            <el-option
                v-for="item in all_config_names"
                :key="item.value"
                :label="item.label"
                :value="item.value">
            </el-option>
          </el-select>
          <br></br>
          <div slot="footer" class="dialog-footer">
            <el-button @click="dialogAddServiceConfigVisible = false">取 消</el-button>
            <el-button type="primary" @click="submitServiceConfigsData($event)">确 定</el-button>
          </div>
        </el-dialog>

        <el-dialog title="修改数据" :visible.sync="dialogUpdateServiceConfigVisible" :modal-append-to-body="false">
          <el-form :model="modify_service_config" ref="modify_service_config">
            <el-form-item label="service" :label-width="formLabelWidth">
              <el-input v-model="modify_service_config.sid" auto-complete="off" readonly="readonly">
              </el-input>
            </el-form-item>
            <el-form-item label="config" :label-width="formLabelWidth">
              <el-select  placeholder="请选择" v-model="modify_service_config.cid" @change="doSelectConfigForSC" style="width: 280px">
                <el-option
                    v-for="item in all_update_configs"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                    :disabled="item.disabled">
                </el-option>
              </el-select>
            </el-form-item>
          </el-form>
          <div slot="footer" class="dialog-footer">
            <el-button @click="dialogUpdateServiceConfigVisible = false">取 消</el-button>
            <el-button type="primary" @click="updateServiceConfigData($event)">确 定</el-button>
          </div>
        </el-dialog>
      </el-tab-pane>

      <el-tab-pane label="service_models" name="service_models">
        <div align="right">
          <el-button type="primary" icon="el-icon-edit" @click="dialogAddServiceModelVisible = true">Add Data</el-button>
        </div>
        <el-table :data="service_model_list">
          <el-table-column v-for="col in cols_service_models"
                           sortable :prop="col.prop"
                           :label="col.label">
          </el-table-column>
          <el-table-column prop="operator" label="操作" width="90" align="center">
            <template slot-scope="props">
              <el-button @click="deleteServiceModelsOperator(props.row)" size="mini" type="danger" icon="el-icon-delete" circle>
              </el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-dialog title="增加service_models数据" :visible.sync="dialogAddServiceModelVisible">
          <el-select  placeholder="请选择service" v-model="default_service" @change="doSelectService">
            <el-option
              v-for="item in all_services"
              :key="item.value"
              :label="item.label"
              :value="item.value">
            </el-option>
          </el-select>

          <el-select  multiple placeholder="请选择model" v-model="default_model" @change="doSelectModel">
            <el-option
              v-for="item in all_models"
              :key="item.value"
              :label="item.label"
              :value="item.value">
            </el-option>
          </el-select>

          <div slot="footer" class="dialog-footer">
            <el-button @click="dialogAddServiceModelVisible = false">取 消</el-button>
            <el-button type="primary" @click="submitServiceModel($event)">确 定</el-button>
          </div>
        </el-dialog>
      </el-tab-pane>

      <el-tab-pane label="model_histories" name="model_histories">
        <div align="right">
          <el-button type="primary" icon="el-icon-edit" @click="dialogAddModelHistoryVisible = true">Add Data</el-button>
        </div>
        <el-table :data="model_history_list">
          <el-table-column v-for="col in cols_model_histories"
                           sortable :prop="col.prop"
                           :label="col.label">
          </el-table-column>
          <el-table-column prop="operator" label="操作" width="90" align="center">
            <template slot-scope="props">
              <el-button @click="dialogUpdateModelHistoryVisible = true,
                updateFormModelHistory.id = props.row.id,
                updateFormModelHistory.model_name = props.row.model_name,
                updateFormModelHistory.timestamp = props.row.timestamp,
                updateFormModelHistory.is_locked = props.row.is_locked,
                updateFormModelHistory.desc = props.row.desc,
                updateFormModelHistory.md5 = props.row.md5" size="mini" type="primary" icon="el-icon-edit" circle>
              </el-button>
              <el-button @click="deleteModelHistory(props.row)" size="mini" type="danger" icon="el-icon-delete" circle>
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-dialog title="增加model_history数据" :visible.sync="dialogAddModelHistoryVisible" :modal-append-to-body='false'>
          <el-form :model="formModelHistory" ref="formModel">
            <el-form-item label="model_name" :label-width="formLabelWidth" required>
              <el-input v-model="formModelHistory.model_name" auto-complete="off"></el-input>
            </el-form-item>
            <el-form-item label="timestamp" :label-width="formLabelWidth" required="">
              <el-input v-model="formModelHistory.timestamp" autocomplete="off"></el-input>
            </el-form-item>
            <el-form-item label="md5" :label-width="formLabelWidth">
              <el-input v-model="formModelHistory.md5" autocomplete="off"></el-input>
            </el-form-item>
            <el-form-item label="is_locked" :label-width="formLabelWidth">
              <el-input v-model="formModelHistory.is_locked" autocomplete="off"></el-input>
            </el-form-item>
            <el-form-item label="desc" :label-width="formLabelWidth">
              <el-input v-model="formModelHistory.desc" autocomplete="off"></el-input>
            </el-form-item>
          </el-form>
          <div slot="footer" class="dialog-footer">
            <el-button @click="dialogAddModelHistoryVisible = false">取 消</el-button>
            <el-button type="primary" @click="submitModelHistory($event)">确 定</el-button>
          </div>
        </el-dialog>

        <el-dialog title="修改数据" :visible.sync="dialogUpdateModelHistoryVisible">
          <el-form :model="formModelHistory" ref="formService">
            <el-input v-model="updateFormModelHistory.id" auto-complete="off" type="hidden">
              </el-input>
            <el-form-item label="model name" :label-width="formLabelWidth">
              <el-input v-model="updateFormModelHistory.model_name" auto-complete="off"></el-input>
            </el-form-item>
            <el-form-item label="timestamp" :label-width="formLabelWidth">
              <el-input v-model="updateFormModelHistory.timestamp" autocomplete="off"></el-input>
            </el-form-item>
            <el-form-item label="md5" :label-width="formLabelWidth">
              <el-input v-model="updateFormModelHistory.md5" autocomplete="off"></el-input>
            </el-form-item>
            <el-form-item label="is_locked" :label-width="formLabelWidth">
              <el-input v-model="updateFormModelHistory.is_locked" autocomplete="off"></el-input>
            </el-form-item>
            <el-form-item label="desc" :label-width="formLabelWidth">
              <el-input v-model="updateFormModelHistory.desc" autocomplete="off"></el-input>
            </el-form-item>
          </el-form>
          <div slot="footer" class="dialog-footer">
            <el-button @click="dialogUpdateModelHistoryVisible = false">取 消</el-button>
            <el-button type="primary" @click="updateModelHistory($event)">确 定</el-button>
          </div>
        </el-dialog>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

function sleep (time) {
    return new Promise((resolve) => setTimeout(resolve, time));
}

<script>
    import axios from 'axios'
    import qs from 'qs'

    export default {
        data() {
            return {
                loadBalanceIp: '10.93.192.222:9508',
                tabPosition: 'hosts',
                curTableName: '',
                database: '',
                upLoadfileList: [],
                dialogAddHostVisible: false,
                dialogUpdateHostVisible: false,
                dialogAddServiceVisible: false,
                dialogAddConfigVisible: false,
                dialogUpdateServiceVisible: false,
                dialogUpdateConfigVisible: false,
                dialogAddModelVisible: false,
                dialogAddHostServiceVisible: false,
                dialogUpdateHostServiceVisible: false,
                dialogAddServiceConfigVisible: false,
                dialogUpdateServiceConfigVisible: false,
                dialogUpdateModelVisible: false,
                dialogAddServiceModelVisible: false,
                dialogUpdateServiceModelVisible: false,
                dialogAddModelHistoryVisible: false,
                dialogUpdateModelHistoryVisible: false,
                dialogFormVisible: false,
                dialogBatchAddHostsVisible: false,
                DeleteHost: false,
                formLabelWidth: '160px',
                host_count: 0,
                host_infos: [],
                all_list: [],
                all_models: [],
                all_host_ip: [],
                all_services: [],
                all_services_sc: [],
                config_list: [],
                all_config_names: [],
                all_update_services: [],
                all_update_configs: [],
                default_ip: '',
                default_config: '',
                default_service: '',
                default_model: '',
                all_model_list: [],
                host_service_list: [],
                service_config_list: [],
                service_model_list: [],
                model_history_list: [],
                select_host: [],
                original_service: '',
                select_service: '',
                select_config: '',
                select_model: '',
                host_services_map: {},
                service_configs_map: {},
                addFormHost: {
                    ip: '',
                    data_center: '',
                    desc: '',
                },
                updateFormHost: {
                    ip: '',
                    data_center: '',
                    desc: '',
                },
                addBatchHostsForm: {
                    host_ips: '',
                },
                update_service: '',
                update_service_sc: '',
                update_model: '',
                update_model_histories_name: '',
                formService: {
                    service_name: '',
                    desc: '',
                },
                formConfig: {
                    desc: '',
                    config: '',
                },
                updateFormService: {
                    service_name: '',
                    desc: '',
                },
                updateFormConfig: {
                    desc: '',
                    config: '',
                    id:'',
                },
                formModel: {
                    name: '',
                    path: '',
                    desc: '',
                    extension: '',
                },
                updateFormModel: {
                    name: '',
                    path: '',
                    desc: '',
                    extension: '',
                },
                formHostService: {
                    hid: '',
                    sid: '',
                    load_weight: '',
                    desc: '',
                },
                updateHostService: {
                    hid: '',
                    sid: '',
                    load_weight: '',
                    desc: '',
                },
                updateServiceConfig: {
                    sid: '',
                    cid: '',
                    desc: '',
                },
                modify_host_service: {
                    id: '',
                    hid: '',
                    sid: '',
                    load_weight: '',
                },
                modify_service_config: {
                    id: '',
                    sid: '',
                    cid: '',
                    desc: '',
                },
                formServiceModel: {
                    sid: '',
                    mid: '',
                    desc: '',
                },
                formModelHistory: {
                    model_name: '',
                    timestamp: '',
                    md5: '',
                    is_locked: '',
                    desc: '',
                },
                updateFormModelHistory: {
                    id: '',
                    model_name: '',
                    timestamp: '',
                    md5: '',
                    is_locked: '',
                    desc: '',
                },
                host_info_search: '',
                service_info_search: '',
                model_info_search: '',
                service_model_search: '',
                cols_host: [
                    { prop: 'id', label: 'id'},
                    { prop: 'ip', label: 'ip'},
                    { prop: 'data_center', label: 'data_center',
                        filter: [
                            {text:'北显', value: '北显'},
                            {text:'大白楼', value:'大白楼'},
                            {text:'华为云',value:'华为云'},
                            {text:'阿里云', value:'阿里云'},
                        ]
                    },
                    { prop: 'desc', label: 'desc'},
                    { prop: 'created_at', label: 'created_at'},
                    { prop: 'updated_at', label: 'updated_at'},
                ],
                cols_service: [
                    { prop: 'id', label:'id'},
                    { prop: 'name', label: 'name'},
                    { prop: 'desc', label: 'desc'},
                    { prop: 'created_at', label: 'created_at'},
                    { prop: 'updated_at', label: 'updated_at'},
                ],
                cols_config: [
                    { prop: 'id', label:'id'},
                    { prop: 'desc', label: 'desc'},
                    { prop: 'config', label: 'config'},
                    { prop: 'created_at', label: 'created_at'},
                    { prop: 'updated_at', label: 'updated_at'},
                ],
                cols_model: [
                    { prop: 'id', label: 'id'},
                    { prop: 'name', label: 'name'},
                    { prop: 'path', label: 'path'},
                    { prop: 'desc', label: 'desc'},
                    { prop: 'extension', label: 'extension'},
                    { prop: 'created_at', label: 'created_at'},
                    { prop: 'updated_at', label: 'updated_at'},
                ],
                cols_host_services: [
                    { prop: 'id', label: 'id'},
                    { prop: 'hid', label: 'hid'},
                    { prop: 'sid', label: 'sid'},
                    { prop: 'load_weight', label: 'load_weight'},
                    { prop: 'desc', label: 'desc'},
                    { prop: 'created_at', label: 'created_at'},
                    { prop: 'updated_at', label: 'updated_at'},
                ],
                cols_service_configs: [
                    { prop: 'id', label: 'id'},
                    { prop: 'sid', label: 'sid'},
                    { prop: 'cid', label: 'cid'},
                    { prop: 'desc', label: 'desc'},
                    { prop: 'created_at', label: 'created_at'},
                    { prop: 'updated_at', label: 'updated_at'},
                ],
                cols_service_models: [
                    { prop: 'id', label: 'id'},
                    { prop: 'sid', label: 'sid'},
                    { prop: 'mid', label: 'mid'},
                    { prop: 'desc', label: 'desc'},
                    { prop: 'created_at', label: 'created_at'},
                    { prop: 'updated_at', label: 'updated_at'},
                ],
                cols_model_histories: [
                    { prop: 'id', label: 'id'},
                    { prop: 'model_name', label: 'model_name' },
                    { prop: 'timestamp', label: 'timestamp' },
                    { prop: 'md5', label: 'md5' },
                    { prop: 'is_locked', label: 'is_locked'},
                    { prop: 'desc', label: 'desc'},
                    { prop: 'created_at', label: 'created_at'},
                    { prop: 'updated_at', label: 'updated_at'},
                ],
                load_weight_input: '',
                openServices: [],
            }
        },
        created () {
            var get_cur_tab = localStorage.getItem("curTab");
            this.tabPosition = get_cur_tab;
            var cache_his_sid = window.localStorage.getItem("deleteHostServices");
            if (cache_his_sid != null) {
                var delete_hid = cache_his_sid.split(":")[0];
                var delete_sid = cache_his_sid.split(":")[1];
                axios.get('/mysql/delete', {
                    params: {
                        table: "host_services",
                        hid: delete_hid,
                        sid: delete_sid
                    }
                }).then(function (response) {
                    console.log("delete succ.");
                }).catch(function (error) {
                    console.log(error);
                });
            }
            window.localStorage.removeItem("deleteHostServices");
            axios.get('/mysql/tables')
                .then(response => {
                    var response_data = response.data;
                    if (response_data != "undefined" && response_data != null) {
                        this.database = response_data.Database;
                    }
                });
            axios.get('/mysql/show?table=hosts')
                .then(response => {
                    var response_data = response.data;
                    this.curTableName = response_data.Table_name;
                    var data_list = response_data;
                    var host_list = [];
                    for (var i = 0; i < data_list.length; i++) {
                        var item = {
                            id : data_list[i].ID,
                            ip: data_list[i].Ip,
                            data_center : data_list[i].DataCenter,
                            desc: data_list[i].Desc,
                            created_at: data_list[i].CreatedAt,
                            updated_at: data_list[i].UpdatedAt,
                        }
                        host_list.push(item);
                        var id_host = data_list[i].ID.toString() + ":" + data_list[i].Ip;
                        this.all_host_ip.push({value: id_host, label: id_host});
                    }
                    this.host_infos = host_list;
                    if (this.host_infos != "undefined" && this.host_infos != null) {
                        this.host_count = this.host_infos.length;
                    }
                });
            // render service table
            axios.get('/mysql/show', {
                params: {
                    table: "services",
                }
            }).then(response => {
                var response_data = response.data;
                var data_list = response_data;
                var service_list = [];
                var all_services = [];
                for (var i = 0; i < data_list.length; i++) {
                    var item = {
                        id: data_list[i].ID,
                        name: data_list[i].Name,
                        desc: data_list[i].Desc,
                        created_at: data_list[i].CreatedAt,
                        updated_at: data_list[i].UpdatedAt,
                        balance_status: false,
                    }
                    service_list.push(item);
                    var id_service = data_list[i].ID.toString() + ":" + data_list[i].Name;
                    all_services.push({ value: id_service, label: id_service});
                }
                this.all_list = service_list;
                this.all_services = all_services;
                this.refreshLoadBalanceServices()
            }).catch(function (error) {
                console.log(error);
            });
            // render configs table
            axios.get('/mysql/show', {
                params: {
                    table: "configs",
                }
            }).then(response => {
                var response_data = response.data;
                var data_list = response_data;
                var config_list = [];
                var all_config_names = [];
                for (var i = 0; i < data_list.length; i++) {
                    var item = {
                        id: data_list[i].ID,
                        desc: data_list[i].Description,
                        config: data_list[i].Config,
                        created_at: data_list[i].CreatedAt,
                        updated_at: data_list[i].UpdatedAt,
                    }
                    config_list.push(item);
                    var id_config = data_list[i].ID.toString() + ":" + data_list[i].Description;
                    all_config_names.push({ value: id_config, label: id_config});
                }
                this.config_list = config_list;
                this.all_config_names = all_config_names;
            }).catch(function (error) {
                console.log(error);
            });
            // render models table
            axios.get('/mysql/show', {
                params: {
                    table: "models",
                }
            }).then(response => {
                var response_data = response.data;
                var data_list = response_data;
                var model_list = [];
                var model_names = [];
                for (var i = 0; i < data_list.length; i++) {
                    var item = {
                        id: data_list[i].ID,
                        name: data_list[i].Name,
                        path : data_list[i].Path,
                        desc: data_list[i].Desc,
                        extension: data_list[i].Extension,
                        created_at: data_list[i].CreatedAt,
                        updated_at: data_list[i].UpdatedAt,
                    }
                    model_list.push(item);
                    var id_models = data_list[i].ID.toString() + ":" + data_list[i].Name;
                    model_names.push({ value: id_models, label: id_models});
                }
                this.all_model_list = model_list;
                this.all_models = model_names;
            }).catch(function (error) {
                console.log(error);
            });
            // render host_services table
            axios.get('/mysql/show', {
                params: {
                    table: "host_services",
                }
            }).then(response => {
                var response_data = response.data;
                var data_list = response_data.host_services;
                var host_service_list = [];
                var all_selected_services = [];
                var select_hosts = new Set();
                for(var i = 0; i < data_list.length; i++) {
                    var cur_host = data_list[i].Desc.split('->')[0];
                    select_hosts.add(cur_host);
                }
                var host_service_map = new Map();
                for (var host of select_hosts) {
                    var service_list = [];
                    for (var i = 0; i < data_list.length; i++) {
                        var host_service_ = data_list[i].Desc.split('->');
                        if ( host_service_[0] == host) {
                            service_list.push(host_service_[1]);
                        }
                    }
                    host_service_map[host] = service_list;
                }
                this.host_services_map = host_service_map;
                // 构造serviceMap
                var services = response_data.services;
                var serviceMap = new Map();
                for (var service of services) {
                    serviceMap[service.ID] = service.Name
                }
                // 构造hostMap
                var hosts = response_data.hosts;
                var hostMap = new Map();
                for (var host of hosts) {
                    hostMap[host.ID] = host.Ip
                }
                for (var i = 0; i < data_list.length; i++) {
                    var desc_host = "";
                    var desc_service = "";
                    if(hostMap.hasOwnProperty(data_list[i].Hid)){
                        desc_host = hostMap[data_list[i].Hid];
                    }
                     if(serviceMap.hasOwnProperty(data_list[i].Sid)){
                        desc_service = serviceMap[data_list[i].Sid];
                    }
                    var item = {
                        id: data_list[i].ID,
                        hid: data_list[i].Hid + " (" + desc_host + ")",
                        sid: data_list[i].Sid + " (" + desc_service + ")",
                        desc: data_list[i].Desc,
                        load_weight: data_list[i].LoadWeight,
                        created_at: data_list[i].CreatedAt,
                        updated_at: data_list[i].UpdatedAt,
                        disabled: true
                    }
                    var id_service = data_list[i].Sid.toString() + ":" + desc_service;
                    all_selected_services.push(id_service);
                    host_service_list.push(item);
                }
                this.host_service_list = host_service_list;
            }).catch(function (error) {
                console.log(error);
            });
            // render service_configs table
            axios.get('/mysql/show', {
                params: {
                    table: "service_configs",
                }
            }).then(response => {
                var response_data = response.data;
                var data_list = response_data;
                var service_config_list = [];
                var all_services_sc = [];
                var all_services = this.all_services;
                var select_services = new Set();
                for (var i = 0; i < data_list.length; i++) {
                    var desc_split = data_list[i].Description.split('->');
                    var desc_service = "";
                    var desc_config = "";
                    if (desc_split.length >= 2) {
                        desc_service = desc_split[0];
                        desc_config = desc_split[1];
                    }
                    var item = {
                        id: data_list[i].ID,
                        sid: data_list[i].Sid + " (" + desc_service + ")",
                        cid: data_list[i].Cid + " (" + desc_config + ")",
                        desc: data_list[i].Description,
                        created_at: data_list[i].CreatedAt,
                        updated_at: data_list[i].UpdatedAt,
                        disabled: true
                    }
                    var id_service = data_list[i].Sid.toString() + ":" + desc_service;
                    service_config_list.push(item);
                    select_services.add(id_service);
                }
                for (var i = 0; i < all_services.length; i++) {
                    var flag = false;
                    if (select_services.has(all_services[i].value)) {
                        flag = true;
                    }
                    var item = {
                        value: all_services[i].value,
                        label: all_services[i].label,
                        disabled: flag
                    }
                    all_services_sc.push(item);
                }
                this.all_services_sc = all_services_sc;
                this.service_config_list = service_config_list;
            }).catch(function (error) {
                console.log(error);
            });
            //render service_models table
            axios.get('/mysql/show', {
                params: {
                    table: "service_models",
                }
            }).then(response => {
                var response_data = response.data;
                var data_list = response_data.service_models;
                var service_models = [];
                var all_models = [];

                // 构造serviceMap
                var services = response_data.services;
                var serviceMap = new Map();
                for (var service of services) {
                    serviceMap[service.ID] = service.Name
                }
                // 构造modelMap
                var models = response_data.models;
                var modelMap = new Map();
                for (var model of models) {
                    modelMap[model.ID] = model.Name
                }
                for (var i = 0; i < data_list.length; i++) {
                    var desc_service = "";
                    var desc_model = "";
                    if(serviceMap.hasOwnProperty(data_list[i].Sid)){
                        desc_service = serviceMap[data_list[i].Sid];
                    }
                    if(modelMap.hasOwnProperty(data_list[i].Mid)){
                        desc_model = modelMap[data_list[i].Mid];
                    }
                    var item = {
                        id: data_list[i].ID,
                        sid: data_list[i].Sid + " (" + desc_service + ")",
                        mid: data_list[i].Mid + " (" + desc_model + ")",
                        desc: data_list[i].Desc,
                        created_at: data_list[i].CreatedAt,
                        updated_at: data_list[i].UpdatedAt,
                    }
                    service_models.push(item);
                }
                this.service_model_list = service_models;
            }).catch(function (error) {
                console.log(error);
            });
            //render model_histories table
            axios.get('/mysql/show', {
                params: {
                    table: "model_histories",
                }
            }).then(response => {
                var response_data = response.data;
                var data_list = response_data;
                var model_histories = [];
                for (var i = 0; i < data_list.length; i++) {
                    var item = {
                        id: data_list[i].ID,
                        model_name: data_list[i].ModelName,
                        timestamp: data_list[i].Timestamp,
                        md5: data_list[i].Md5,
                        is_locked: data_list[i].IsLocked,
                        desc: data_list[i].Desc,
                        created_at: data_list[i].CreatedAt,
                        updated_at: data_list[i].UpdatedAt,
                    }
                    model_histories.push(item);
                }
                this.model_history_list = model_histories;
            }).catch(function (error) {
                console.log(error);
            });
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
        },
        methods: {
            changeLoadBalance: function($event, row) {
                var _this = this
                var service_name = row.name
                if($event) {
                    axios.get('http://'+_this.loadBalanceIp+'/load_balance/insert', {
                        params : {
                            service_name: service_name
                        }
                    }).then(function (response) {
                        if (response.data.code == 0 || response.data.code == 204) {
                            _this.$message({
                                type: 'success',
                                message: '开启自动权重成功，service_name:'+ service_name
                            });
                        } else {
                            _this.$message({
                                type: 'error',
                                message: '开启自动权重失败,' + response.data.msg
                            });
                            _this.doRefreshPage();
                        }

                    }).catch(function (error) {
                        _this.$message({
                            type: 'error',
                            message: '开启自动权重失败, 网络错误'
                        });
                        console.log(error);
                    });

                } else {
                    axios.get('http://'+_this.loadBalanceIp+'/load_balance/delete', {
                        params : {
                            service_name: service_name
                        }
                    }).then(function (response) {
                       if (response.data.code == 0) {
                            _this.$message({
                                type: 'success',
                                message: '取消自动权重成功，service_name:'+ service_name
                            });
                        } else {
                            _this.$message({
                                type: 'error',
                                message: '取消自动权重失败,' + response.data.msg
                            });
                            _this.doRefreshPage();
                        }
                    }).catch(function (error) {
                        _this.$message({
                            type: 'error',
                            message: '取消自动权重失败, 网络错误'
                        });
                        console.log(error);
                    });
                }

            },
            refreshLoadBalanceServices() {
                var _this = this
                _this.openServices = [];
                axios.get('http://'+_this.loadBalanceIp+'/load_balance/get', {
                }).then(function (response) {
                    if (response.data.code == 0) {
                        _this.openServices = response.data.data
                        for (var i = 0; i < _this.all_list.length; i++) {
                            var openServiceIndex = _this.openServices.indexOf(_this.all_list[i].name)
                            var balance_status = openServiceIndex == -1 ? false : true
                            _this.all_list[i].balance_status = balance_status
                        }
                    } else {
                        _this.$message({
                            type: 'error',
                            message: '获取自动权重数据失败,' + response.data.msg
                        });
                    }
                }).catch(function (error) {
                    _this.$message({
                        type: 'error',
                        message: '获取自动权重数据失败, 网络错误'
                    });
                    console.log(error);
                });
            },
            resetServiceWeight(rowData) {
                this.$confirm('确定要重置'+ rowData.name +'服务下所有机器权重吗？', '提示', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    type: 'warning'
                }).then(() => {
                    axios.get('/load_balance/reset_weight', {
                        params: {
                            sid: rowData.id,
                        }
                    }).then(response => {
                        var resp = response.data;
                        if(resp.code != 0) {
                            this.$message({
                                type: 'error',
                                message: resp.msg
                            });
                            return;
                        }
                        this.$message({
                            type: 'success',
                            message: resp.msg
                        });
                    }).catch(function (error) {
                        this.$message({
                            type: 'error',
                            message: '网络错误，请重试'
                        });
                        return;
                    });
                }).catch(() => {
                    
                });
            },
            doSelectHost(host_value) {
                this.select_host = host_value;
            },
            doSelectService(service_list_value) {
                this.modify_host_service.sid = service_list_value;
                this.select_service = service_list_value;
            },
            doSelectServiceForSC(service_list_value) {
                this.modify_service_config.sid = service_list_value;
                this.select_service = service_list_value;
            },
            doSelectConfigForSC(config_value) {
                this.select_config = config_value;
            },
            updateDisableStatus() {
                // 根据点击的row的host serivce动态修改all_update_service中的disable状态
                var service_name_list = [];
                for(var i = 0; i < this.all_services.length; i++) {
                    service_name_list.push(this.all_services[i].value);
                }
                var prev_host = this.modify_host_service.hid.split(' ')[1];
                // 找出当前host上已经存在service
                var host_ = prev_host.substring(1, prev_host.length -1);
                var update_service_with_disabled = [];
                var exist_services_list = this.host_services_map[host_];
                var item;
                for( var i = 0; i < service_name_list.length; i++) {
                    var service_name = service_name_list[i].split(':')[1];
                    if (exist_services_list != null &&  exist_services_list.indexOf(service_name) != -1) {
                        item = {
                            value: service_name_list[i],
                            label: service_name_list[i],
                            disabled: true
                        }
                    } else {
                        item = {
                            value: service_name_list[i],
                            label: service_name_list[i]
                        }
                    }
                    update_service_with_disabled.push(item);
                }
                this.all_update_services = update_service_with_disabled;
            },
            ConfigDisableStatus() {
                // 根据点击的row的serivce config动态修改all_update_configs中的disable状态
                var config_name_list = [];
                for(var i = 0; i < this.all_config_names.length; i++) {
                    config_name_list.push(this.all_config_names[i].value);
                }
                var cur_config = this.modify_service_config.cid.split(' ')[1];
                var cur_config_name = cur_config.substring(1, cur_config.length -1);
                var update_config_with_disabled = [];
                for( var i = 0; i < config_name_list.length; i++) {
                    var item;
                    var config_name = config_name_list[i].split(':')[1];
                    if (cur_config_name == config_name) {
                        item = {
                            value: config_name_list[i],
                            label: config_name_list[i],
                            disabled: true
                        }
                    } else {
                        item = {
                            value: config_name_list[i],
                            label: config_name_list[i]
                        }
                    }
                    update_config_with_disabled.push(item);
                }
                this.all_update_configs = update_config_with_disabled;
            },
            fillLoadWeight(load_weight_value) {
                this.load_weight_input = load_weight_value;
                if (parseInt(this.load_weight_input) < 0 ) {
                    this.$message({
                        message: 'load_weight应大于0的整数',
                        type: 'success',
                        customClass: 'login_alert',
                        duration: 2000
                    })
                }
            },
            doSelectModel(model_list_value) {
                this.service_model_list = model_list_value;
            },
            handleClick(tab) {
                this.curTableName = tab.label;
                this.tabPosition = tab.label;
                if (this.curTableName == "hosts") {
                    axios.get('/mysql/show?table=hosts')
                        .then(response => {
                            var response_data = response.data;
                            this.curTableName = response_data.Table_name;
                            var data_list = response_data;
                            var host_list = [];
                            for (var i = 0; i < data_list.length; i++) {
                                var item = {
                                    id : data_list[i].ID,
                                    ip: data_list[i].Ip,
                                    data_center : data_list[i].DataCenter,
                                    desc: data_list[i].Desc,
                                    created_at: data_list[i].CreatedAt,
                                    updated_at: data_list[i].UpdatedAt,
                                }
                                host_list.push(item);
                            }
                            this.host_infos = host_list;
                            this.host_count = this.host_infos.length;
                        });
                } else if (this.curTableName == "services") {
                    axios.get('/mysql/show', {
                        params: {
                            table: "services",
                        }
                    }).then(response => {
                        var response_data = response.data;
                        var data_list = response_data;
                        var service_list = [];
                        var all_services = [];
                        for (var i = 0; i < data_list.length; i++) {
                            var item = {
                                id: data_list[i].ID,
                                name: data_list[i].Name,
                                desc: data_list[i].Desc,
                                created_at: data_list[i].CreatedAt,
                                updated_at: data_list[i].UpdatedAt,
                                balance_status: false,
                            }
                            service_list.push(item);
                            var id_service = data_list[i].ID.toString() + ":" + data_list[i].Name;
                            all_services.push({ value: id_service, label: id_service});
                        }
                        this.all_list = service_list;
                        this.all_services = all_services;
                        this.refreshLoadBalanceServices();
                    }).catch(function (error) {
                        console.log(error);
                    });
                } else if (this.curTableName == "configs") {
                    axios.get('/mysql/show', {
                        params: {
                            table: "configs",
                        }
                    }).then(response => {
                        var response_data = response.data;
                        var data_list = response_data;
                        var config_list = [];
                        var all_config_names = [];
                        for (var i = 0; i < data_list.length; i++) {
                            var item = {
                                id: data_list[i].ID,
                                desc: data_list[i].Description,
                                config: data_list[i].Config,
                                created_at: data_list[i].CreatedAt,
                                updated_at: data_list[i].UpdatedAt,
                            }
                            config_list.push(item);
                            var id_config = data_list[i].ID.toString() + ":" + data_list[i].Description;
                            all_config_names.push({ value: id_config, label: id_config});
                        }
                        this.config_list = config_list;
                        this.all_config_names = all_config_names;
                    }).catch(function (error) {
                        console.log(error);
                    });
                } else if ( this.curTableName == "models") {
                    axios.get('/mysql/show', {
                        params: {
                            table: "models",
                        }
                    }).then(response => {
                        var response_data = response.data;
                        var data_list = response_data;
                        var model_list = [];
                        var model_names = [];
                        for (var i = 0; i < data_list.length; i++) {
                            var item = {
                                id: data_list[i].ID,
                                name: data_list[i].Name,
                                path : data_list[i].Path,
                                desc: data_list[i].Desc,
                                extension: data_list[i].Extension,
                                created_at: data_list[i].CreatedAt,
                                updated_at: data_list[i].UpdatedAt,
                            }
                            model_list.push(item);
                            var id_models = data_list[i].ID.toString() + ":" + data_list[i].Name;
                            model_names.push({ value: id_models, label: id_models});
                        }
                        this.all_model_list = model_list;
                        this.all_models = model_names;
                    }).catch(function (error) {
                        console.log(error);
                    });
                } else if (this.curTableName == "service_configs"){
                    axios.get('/mysql/show', {
                        params: {
                            table: "service_configs",
                        }
                    }).then(response => {
                        var response_data = response.data;
                        var data_list = response_data;
                        var service_config_list = [];
                        var all_services_sc = [];
                        var all_services = this.all_services;
                        var select_services = new Set();
                        for (var i = 0; i < data_list.length; i++) {
                            var desc_split = data_list[i].Description.split('->');
                            var desc_service = "";
                            var desc_config = "";
                            if (desc_split.length >= 2) {
                                desc_service = desc_split[0];
                                desc_config = desc_split[1];
                            }
                            var item = {
                                id: data_list[i].ID,
                                sid: data_list[i].Sid + " (" + desc_service + ")",
                                cid: data_list[i].Cid + " (" + desc_config + ")",
                                desc: data_list[i].Description,
                                created_at: data_list[i].CreatedAt,
                                updated_at: data_list[i].UpdatedAt,
                                disabled: true
                            }
                            var id_service = data_list[i].Sid.toString() + ":" + desc_service;
                            service_config_list.push(item);
                            select_services.add(id_service);
                        }
                        for (var i = 0; i < all_services.length; i++) {
                            var flag = false;
                            if (select_services.has(all_services[i].value)) {
                                flag = true;
                            }
                            var item = {
                                value: all_services[i].value,
                                label: all_services[i].label,
                                disabled: flag
                            }
                            all_services_sc.push(item);
                        }
                        this.all_services_sc = all_services_sc;
                        this.service_config_list = service_config_list;
                    }).catch(function (error) {
                        console.log(error);
                    });
                 } else if (this.curTableName == "host_services"){
                    axios.get('/mysql/show', {
                        params: {
                            table: "host_services",
                        }
                    }).then(response => {
                        var response_data = response.data;
                        var data_list = response_data.host_services;
                        var host_service_list = [];
                        // 构造serviceMap
                        var services = response_data.services;
                        var serviceMap = new Map();
                        for (var service of services) {
                            serviceMap[service.ID] = service.Name
                        }
                        // 构造hostMap
                        var hosts = response_data.hosts;
                        var hostMap = new Map();
                        for (var host of hosts) {
                            hostMap[host.ID] = host.Ip
                        }
                        for (var i = 0; i < data_list.length; i++) {
                            var desc_host = "";
                            var desc_service = "";
                            if(hostMap.hasOwnProperty(data_list[i].Hid)){
                                desc_host = hostMap[data_list[i].Hid];
                            }
                            if(serviceMap.hasOwnProperty(data_list[i].Sid)){
                                desc_service = serviceMap[data_list[i].Sid];
                            }
                            var item = {
                                id: data_list[i].ID,
                                hid: data_list[i].Hid + " (" + desc_host + ")",
                                sid: data_list[i].Sid + " (" + desc_service + ")",
                                load_weight: data_list[i].LoadWeight,
                                desc: data_list[i].Desc,
                                created_at: data_list[i].CreatedAt,
                                updated_at: data_list[i].UpdatedAt,
                            }
                            host_service_list.push(item);
                        }
                        this.host_service_list = host_service_list;
                    }).catch(function (error) {
                        console.log(error);
                    });
                } else if (this.curTableName == "service_models") {
                    axios.get('/mysql/show', {
                        params: {
                            table: "service_models",
                        }
                    }).then(response => {
                        var response_data = response.data;
                        var data_list = response_data.service_models;
                        var service_models = [];
                        var all_models = [];
                        // 构造serviceMap
                        var services = response_data.services;
                        var serviceMap = new Map();
                        for (var service of services) {
                            serviceMap[service.ID] = service.Name
                        }
                        // 构造modelMap
                        var models = response_data.models;
                        var modelMap = new Map();
                        for (var model of models) {
                            modelMap[model.ID] = model.Name
                        }
                        for (var i = 0; i < data_list.length; i++) {
                            var desc_service = "";
                            var desc_model = "";
                            if(serviceMap.hasOwnProperty(data_list[i].Sid)){
                                desc_service = serviceMap[data_list[i].Sid];
                            }
                            if(modelMap.hasOwnProperty(data_list[i].Mid)){
                                desc_model = modelMap[data_list[i].Mid];
                            }
                            var item = {
                                id: data_list[i].ID,
                                sid: data_list[i].Sid + " (" + desc_service + ")",
                                mid: data_list[i].Mid + " (" + desc_model + ")",
                                desc: data_list[i].Desc,
                                created_at: data_list[i].CreatedAt,
                                updated_at: data_list[i].UpdatedAt,
                            }
                            service_models.push(item);
                        }
                        this.service_model_list = service_models;
                    }).catch(function (error) {
                        console.log(error);
                    });
                } else if (this.curTableName == "model_histories") {
                    axios.get('/mysql/show', {
                        params: {
                            table: "model_histories",
                        }
                    }).then(response => {
                        var response_data = response.data;
                        var data_list = response_data;
                        var model_histories = [];
                        for (var i = 0; i < data_list.length; i++) {
                            var item = {
                                id: data_list[i].ID,
                                model_name: data_list[i].ModelName,
                                timestamp: data_list[i].Timestamp,
                                md5: data_list[i].Md5,
                                is_locked: data_list[i].IsLocked,
                                desc: data_list[i].Desc,
                                created_at: data_list[i].CreatedAt,
                                updated_at: data_list[i].UpdatedAt,
                            }
                            model_histories.push(item);
                        }
                        this.model_history_list = model_histories;
                    }).catch(function (error) {
                        console.log(error);
                    });
                }
            },
            submitHostData(event) {
                event.preventDefault();
                var host_ip = this.addFormHost.ip.trim();
                var data_center = this.addFormHost.data_center;
                var cur_desc = this.addFormHost.desc;
                axios.get('/mysql/insert', {
                    params : {
                        table: "hosts",
                        ip:    host_ip,
                        data_center: data_center,
                        desc: cur_desc
                    }
                }).then(function (response) {
                    console.log(response);
                }).catch(function (error) {
                    console.log(error);
                });
                this.dialogAddHostVisible = false;
                this.doRefreshPage();
            },
            batchAddHosts(event) {
                event.preventDefault();
                var _this = this
                var host_ips = this.addBatchHostsForm.host_ips.trim();
                axios.post('/mysql/batch_insert_hosts', qs.stringify({
                    host_ips: host_ips
                })).then(function (response) {
                    var resp = response.data;
                    if (resp.code != 0) {
                        _this.$message({
                            type: 'error',
                            message: resp.msg
                        });
                        return;
                    } else {
                        _this.$message({
                            type: 'success',
                            message: resp.msg
                        });
                        _this.addBatchHostsForm.host_ips = '';
                        _this.dialogBatchAddHostsVisible = false;
                        var tab = {label: 'hosts'};
                        _this.handleClick(tab);
                        return;
                    }
                }).catch(function (error) {
                    console.log(error)
                    _this.$message({
                        type: 'error',
                        message: '网络错误，请重试'
                    });
                    return;
                });
            },
            updateHostData(event) {
                event.preventDefault();
                axios.get('/mysql/update', {
                    params : {
                        table: "hosts", ip: this.updateFormHost.ip.trim(),
                        data_center:  this.updateFormHost.data_center,
                        desc: this.updateFormHost.desc
                    }
                }).then(function (response) {
                    console.log(response);
                }).catch(function (error) {
                    console.log(error);
                });
                this.dialogUpdateHostVisible = false;
                this.doRefreshPage();
            },
            submitServiceData(event) {
                event.preventDefault();
                var cur_table = this.curTableName;
                var name = this.formService.service_name.trim();
                var desc = this.formService.desc;
                axios.get('/mysql/insert', {
                    params : {
                        table: cur_table,
                        name : name,
                        desc: desc
                    }
                }).then(function (response) {
                    console.log(response);
                }).catch(function (error) {
                    console.log(error);
                });
                this.dialogAddServiceVisible = false;
                this.doRefreshPage();
                this.curTableName = "services";
            },
            submitConfigData(event) {
                event.preventDefault();
                var config = this.formConfig.config.trim();
                var desc = this.formConfig.desc;
                axios.get('/mysql/insert', {
                    params : {
                        table: "configs",
                        config : config,
                        desc: desc
                    }
                }).then(function (response) {
                    console.log(response);
                }).catch(function (error) {
                    console.log(error);
                });
                this.dialogAddConfigVisible = false;
                this.doRefreshPage();
                this.curTableName = "configs";
            },
            updateServiceData(event) {
                event.preventDefault();
                axios.get('/mysql/update', {
                    params : {
                        table: "services",
                        name: this.updateFormService.service_name.trim(),
                        desc: this.updateFormService.desc
                    }
                }).then(function (response) {
                    console.log(response);
                }).catch(function (error) {
                    console.log(error);
                });
                this.dialogUpdateServiceVisible = false;
                this.doRefreshPage();
            },
            updateConfigData(event) {
                event.preventDefault();
                axios.get('/mysql/update', {
                    params : {
                        table: "configs",
                        config: this.updateFormConfig.config.trim(),
                        desc: this.updateFormConfig.desc,
                        id: this.updateFormConfig.id
                    }
                }).then(function (response) {
                    console.log(response);
                }).catch(function (error) {
                    console.log(error);
                });
                this.dialogUpdateConfigVisible = false;
                this.doRefreshPage();
            },
            submitModelData(event) {
                event.preventDefault();
                var model_name = this.formModel.name.trim();
                var desc = this.formModel.desc;
                var path = this.formModel.path;
                axios.get('/mysql/insert', {
                    params : {
                        table : "models",
                        name : model_name,
                        path : path,
                        desc: desc,
                    }
                }).then(function (response) {
                    console.log(response);
                })
                this.dialogAddModelVisible = false;
                this.doRefreshPage();
                this.curTableName = "models";
            },
            updateHostServiceData(event) {
                event.preventDefault();
                var update_id = this.modify_host_service.id;
                var hid, hid_str;
                if (this.modify_host_service.hid.length != 0) {
                    hid = this.modify_host_service.hid.split(' ')[0];
                    hid_str = this.modify_host_service.hid;
                    var start = hid_str.indexOf('(');
                    var end   = hid_str.indexOf(')');
                    if (start < end) {
                        hid_str = hid_str.substring(start+1, end);
                    } else {
                        console.error("error host parameter:", hid)
                        return;
                    }
                }
                var sid, sid_str;
                if( this.modify_host_service.sid == this.updateHostService.sid) {
                    sid = this.modify_host_service.sid.split(' ')[0];
                    var tmp = this.modify_host_service.sid.split(' ')[1];
                    sid_str = tmp.substring(1, tmp.length - 1).trim();
                } else {
                    sid = this.modify_host_service.sid.split(':')[0];
                    sid_str = this.modify_host_service.sid.split(':')[1];
                }
                var load_weight = this.modify_host_service.load_weight;
                if (load_weight.length == 0) {
                    load_weight = '0';
                }
                var desc = hid_str.trim() + "->" + sid_str;
                axios.get('/mysql/update', {
                  params : {
                      table : "host_services",
                      id : update_id,
                      hid: hid,
                      sid: sid,
                      load_weight: load_weight,
                      desc: desc
                  }
                }).then(function (response) {
                    console.log("update host service ret:",response);
                }).catch(function (error) {
                    console.log(error);
                })
                this.dialogUpdateHostServiceVisible = false;
                this.doRefreshPage();
            },
            updateServiceConfigData(event) {
                event.preventDefault();
                var update_id = this.modify_service_config.id;
                var config_id_desc = this.select_config.split(':');
                var cid, cid_str;
                if (config_id_desc.length == 2) {
                    cid = config_id_desc[0]
                    cid_str = config_id_desc[1]
                } else {
                    console.error("error cid parameter:", config_id_desc)
                    return;
                }
                var sid, sid_str;
                if( this.modify_service_config.sid == this.updateServiceConfig.sid) {
                    sid = this.modify_service_config.sid.split(' ')[0];
                    var tmp = this.modify_service_config.sid.split(' ')[1];
                    sid_str = tmp.substring(1, tmp.length - 1).trim();
                } else {
                    sid = this.modify_service_config.sid.split(':')[0];
                    sid_str = this.modify_service_config.sid.split(':')[1];
                }

                var desc = sid_str + "->" + cid_str;
                axios.get('/mysql/update', {
                  params : {
                      table : "service_configs",
                      id : update_id,
                      cid: cid,
                      sid: sid,
                      desc: desc
                  }
                }).then(function (response) {
                    console.log("update service config ret:",response);
                }).catch(function (error) {
                    console.log(error);
                })
                this.dialogUpdateServiceConfigVisible = false;
                this.doRefreshPage();
	        },
            updateModelDataFunc(event) {
                event.preventDefault();
                axios.get('/mysql/update', {
                    params : {
                        table: "models",
                        name: this.updateFormModel.name,
                        path: this.updateFormModel.path,
                        desc: this.updateFormModel.desc,
                        extension: this.updateFormModel.extension
                    }
                }).then(function (response) {
                    console.log(response);
                }).catch(function (error) {
                    console.log(error);
                });
                this.dialogUpdateModelVisible = false;
                this.doRefreshPage();
            },
            updateModelHistory(event) {
                event.preventDefault();
                axios.get('/mysql/update', {
                    params : {
                        table: "model_histories",
                        id: this.updateFormModelHistory.id,
                        model_name: this.updateFormModelHistory.model_name,
                        timestamp: this.updateFormModelHistory.timestamp,
                        md5: this.updateFormModelHistory.md5,
                        is_locked: this.updateFormModelHistory.is_locked,
                        desc: this.updateFormModelHistory.desc
                    }
                }).then(function (response) {
                    console.log(response);
                }).catch(function (error) {
                    console.log(error);
                });
                this.dialogUpdateModelHistoryVisible = false;
                this.doRefreshPage();
            },
            submitHostServicesData(event) {
                event.preventDefault();
                var service_id = this.select_service.split(':');
                for(var i = 0; i < this.select_host.length; i++) {
                    var cur_host_id = this.select_host[i];
                    var host_id = cur_host_id.split(':');
                    var cur_desc = host_id[1] + '->' + service_id[1];
                    if (this.load_weight_input.length == 0 ) {
                        this.load_weight_input = '0';
                    }
                    axios.get('/mysql/insert', {
                        params : {
                            table: "host_services",
                            hid: host_id[0],
                            sid: service_id[0],
                            load_weight: this.load_weight_input,
                            desc: cur_desc,
                        }
                    }).then(function (response) {
                        console.log(response);
                    })
                }
                this.dialogAddHostServiceVisible = false;
                this.doRefreshPage();
                this.curTableName = "host_services";
            },
            submitServiceConfigsData(event) {
                event.preventDefault();
                var service_id_name = this.select_service.split(':');
                var config_id_name = this.select_config.split(':');
                if (service_id_name.length == 2 && config_id_name.length == 2) {
                    var config_id = config_id_name[0]
                    var service_id = service_id_name[0]
                    var desc = service_id_name[1] + '->' + config_id_name[1];
                    axios.get('/mysql/insert', {
                        params : {
                            table: "service_configs",
                            cid: config_id,
                            sid: service_id,
                            desc: desc,
                        }
                    }).then(function (response) {
                        console.log(response);
                    })
                }

                this.dialogAddServiceConfigVisible = false;
                this.doRefreshPage();
                this.curTableName = "service_configs";
            },
            submitServiceModel(event) {
                event.preventDefault();
                var id_service = this.select_service.split(':');
                for (var i = 0; i < this.service_model_list.length; i++) {
                    var sid_model = this.service_model_list[i].split(':');
                    var cur_desc = id_service[1] + '->' + sid_model[1];
                    axios.get('/mysql/insert', {
                        params : {
                            table: "service_models",
                            sid: id_service[0],
                            mid: sid_model[0],
                            desc: cur_desc,
                        }
                    }).then(function (response) {
                        console.log(response);
                    })
                }
                this.dialogAddServiceModelVisible = false;
                this.doRefreshPage();
            },
            submitModelHistory(event) {
                event.preventDefault();
                var model_name = this.formModelHistory.model_name;
                var timestamp = this.formModelHistory.timestamp;
                var md5 = this.formModelHistory.md5;
                var is_locked = this.formModelHistory.is_locked;
                var desc = this.formModelHistory.desc;
                axios.get('/mysql/insert', {
                    params : {
                        table: "model_histories",
                        model_name: model_name,
                        timestamp: timestamp,
                        md5: md5,
                        is_locked: is_locked,
                        desc: desc,
                    }
                }).then(function (response) {
                    console.log(response);
                })
                this.dialogAddModelHistoryVisible = false;
                this.doRefreshPage();
            },
            deleteHostOperator(row_data) {
                this.$confirm('删除继续？', '提示', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    type: 'warning'
                }).then(() => {
                    this.$message({
                        type: 'success',
                        message: '删除成功!'
                    });
                    // 调用后端删除接口
                    var cur_host_ip = row_data.ip;
                    axios.get('/mysql/delete', {
                        params: {
                            table: "hosts",
                            ip: cur_host_ip,
                        }
                    }).then(response => {
                        var response_data = response.data;
                    }).catch(function (error) {
                        console.log(error);
                    });
                    this.doRefreshPage();
                }).catch(() => {
                    this.$message({
                        type: 'info',
                        message: '已取消删除'
                    });
                });
            },
            deleteConfigOperator(row_data) {
                console.log("select row data:", row_data);
                this.$confirm('删除继续？', '提示', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    type: 'warning'
                }).then(() => {
                    this.$message({
                        type: 'success',
                        message: '删除成功!'
                    });
                    // 调用后端删除接口
                    var cur_config_id = row_data.id;
                    axios.get('/mysql/delete', {
                        params: {
                            table: "configs",
                            id: cur_config_id,
                        }
                    }).then(response => {
                        var response_data = response.data;
                    }).catch(function (error) {
                        console.log(error);
                    });
                    this.doRefreshPage();
                }).catch(() => {
                    this.$message({
                        type: 'info',
                        message: '已取消删除'
                    });
                });
            },
            deleteServiceOperator(row_data) {
                console.log("select row data:", row_data);
                this.$confirm('删除继续？', '提示', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    type: 'warning'
                }).then(() => {
                    this.$message({
                        type: 'success',
                        message: '删除成功!'
                    });
                    // 调用后端删除接口
                    var cur_service_name = row_data.name;
                    axios.get('/mysql/delete', {
                        params: {
                            table: "services",
                            name: cur_service_name,
                        }
                    }).then(response => {
                        var response_data = response.data;
                    }).catch(function (error) {
                        console.log(error);
                    });
                    this.doRefreshPage();
                }).catch(() => {
                    this.$message({
                        type: 'info',
                        message: '已取消删除'
                    });
                });
            },
            deleteModelOperator(row_data) {
                this.$confirm('删除继续？', '提示', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    type: 'warning'
                }).then(() => {
                    this.$message({
                        type: 'success',
                        message: '删除成功!'
                    });
                    // 调用后端删除接口
                    var cur_model_name = row_data.name;
                    axios.get('/mysql/delete', {
                        params: {
                            table: "models",
                            name: cur_model_name,
                        }
                    }).then(response => {
                        var response_data = response.data;
                    }).catch(function (error) {
                        console.log(error);
                    });
                    this.doRefreshPage();
                }).catch(() => {
                    this.$message({
                        type: 'info',
                        message: '已取消删除'
                    });
                });
            },
            deleteHostServiceOperator(row_data) {
                this.$confirm('删除继续？', '提示', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    type: 'warning'
                }).then(() => {
                    this.$message({
                        type: 'success',
                        message: '删除成功!'
                    });
                    // 调用后端删除接口
                    axios.get('/mysql/delete', {
                        params: {
                            table: "host_services",
                            hid:   row_data.hid,
                            sid:   row_data.sid,
                        }
                    }).then(response => {
                        var response_data = response.data;
                    }).catch(function (error) {
                        console.log(error);
                    });
                    this.doRefreshPage();
                }).catch(() => {
                    this.$message({
                        type: 'info',
                        message: '已取消删除'
                    });
                });
            },
            deleteServiceConfigOperator(row_data) {
                this.$confirm('删除继续？', '提示', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    type: 'warning'
                }).then(() => {
                    this.$message({
                        type: 'success',
                        message: '删除成功!'
                    });
                    // 调用后端删除接口
                    axios.get('/mysql/delete', {
                        params: {
                            table: "service_configs",
                            cid:   row_data.cid,
                            sid:   row_data.sid,
                        }
                    }).then(response => {
                        var response_data = response.data;
                    }).catch(function (error) {
                        console.log(error);
                    });
                    this.doRefreshPage();
                }).catch(() => {
                    this.$message({
                        type: 'info',
                        message: '已取消删除'
                    });
                });
            },
            deleteServiceModelsOperator(row_data) {
                this.$confirm('删除继续？', '提示', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    type: 'warning'
                }).then(() => {
                    this.$message({
                        type: 'success',
                        message: '删除成功!'
                    });
                    // 调用后端删除接口
                    var cur_model_name = row_data.name;
                    axios.get('/mysql/delete', {
                        params: {
                            table: "service_models",
                            sid:   row_data.sid,
                            mid:   row_data.mid,
                        }
                    }).then(response => {
                        var response_data = response.data;
                    }).catch(function (error) {
                        console.log(error);
                    });
                    this.doRefreshPage();
                }).catch(() => {
                    this.$message({
                        type: 'info',
                        message: '已取消删除'
                    });
                });
            },
            deleteModelHistory(row_data) {
                this.$confirm('删除继续？', '提示', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    type: 'warning'
                }).then(() => {
                    this.$message({
                        type: 'success',
                        message: '删除成功!'
                    });
                    // 调用后端删除接口
                    var cur_model_id = row_data.id;
                    axios.get('/mysql/delete', {
                        params: {
                            table: "model_histories",
                            id: cur_model_id
                        }
                    }).then(response => {
                        var response_data = response.data;
                    }).catch(function (error) {
                        console.log(error);
                    });
                    this.doRefreshPage();
                }).catch(() => {
                    this.$message({
                        type: 'info',
                        message: '已取消删除'
                    });
                });
            },
            doRefreshPage() {
                window.localStorage.setItem("curTab", this.tabPosition);
                location.reload();
                this.$route.push("/views/mysql_controller/mysql_control").andThen(() => {
                    console.log("curTabName:", this.tabPosition);
                });
            },
            filterHandler(value, row) {
                if(row.data_center.indexOf(value) > -1) {
                    return true;
                }
            },
        },
    }
</script>
