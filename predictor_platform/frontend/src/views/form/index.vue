<template>
  <div class="app-container">
    <el-table
      :data="list"
      element-loading-text="Loading"
      border
      fit
      highlight-current-row
    >
      <el-table-column prop="service_name" align="center" label="服务名称" sortable width="180">
      </el-table-column>

      <el-table-column prop="models" align="center" label="加载模型" sortable width="200">
        <el-tree :data="names" :props="service_list" @node-click="handleNodeClick"></el-tree>
      </el-table-column>

      <el-table-column prop="service_status" label="服务状态" width="120" sortable align="center">
      </el-table-column>

      <el-table-column prop="start_time" label="运行时间" width="120" sortable align="center">
      </el-table-column>

      <el-table-column prop='count' label="qps数量" width="100" sortable align="center">
      </el-table-column>

      <el-table-column prop="link" align="center" label="Extra" width="160">
      </el-table-column>
    </el-table>
  </div>
</template>

<script>

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
      list: [{
        service_name: 'predictor_suprefans',
        models: 'modelA',
        start_time: '2019-07-20 10:00:00',
        count: 1024,
        host_ip: '10.85.101.58',
        service_status: 'on',
        link : '',
      }, {
        service_name: 'predictor_fans',
        models: 'modelB',
        start_time: '2019-07-01 10:00:00',
        count: 1500,
        service_status: 'off',
        host_ip: '10.85.101.136',
        link: '',
      },{
        service_name: 'predictor_fisher',
        models: 'modelA/modelB',
        start_time: '2019-06-30 21:00:00',
        count: 2400,
        service_status: 'on',
        host_ip: '10.85.101.100',
        link: '',
        },
      {
        service_name: 'predictor_joiner',
        models: 'modelA',
        start_time: '2019-06-24 23:00:00',
        count: 2500,
        service_status: 'off',
        host_ip: '10.85.101.99',
        link: '',
      },{
        service_name: 'predictor_joiner',
        models: 'modelA',
        start_time: '2019-06-24 23:00:00',
        count: 2500,
        service_status: 'off',
        host_ip: '10.85.101.99',
        link: '',
      }],
      names: [{
        label: 'modelA',
        children: [{
          label: '10.85.101.58',
        }]
      }, {
        label: 'modelB',
        children: [{
          label: '10.85.101.135',
        }]
      }],
      service_list: {
        children: 'children',
        label: 'label'
      },
    };
  },
  created() {
  },
  methods: {
    fetchData() {
    },
    handleNodeClick(data) {
        new Vue({
          el: "detail_info_table",
        })
    }
  },
}
</script>
