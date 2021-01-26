import Vue from 'vue'
import Router from 'vue-router'

Vue.use(Router)

/* Layout */
import Layout from '@/layout'

/**
 * Note: sub-menu only appear when route children.length >= 1
 * Detail see: https://panjiachen.github.io/vue-element-admin-site/guide/essentials/router-and-nav.html
 *
 * hidden: true                   if set true, item will not show in the sidebar(default is false)
 * alwaysShow: true               if set true, will always show the root menu
 *                                if not set alwaysShow, when item has more than one children route,
 *                                it will becomes nested mode, otherwise not show the root menu
 * redirect: noRedirect           if set noRedirect will no redirect in the breadcrumb
 * name:'router-name'             the name is used by <keep-alive> (must set!!!)
 * meta : {
    roles: ['admin','editor'] control the page roles (you can set multiple roles)
    title: 'title'               the name show in sidebar and breadcrumb (recommend set)
    icon: 'svg-name'             the icon show in the sidebar
    breadcrumb: false            if set false, the item will hidden in breadcrumb(default is true)
    activeMenu: '/example/list'  if set path, the sidebar will highlight the path you set
  }
 */

/**
 * constantRoutes
 * a base page that does not have permission requirements
 * all roles can be accessed
 */
export const constantRoutes = [
  {
    path: '/login',
    component: () => import('@/views/login/index'),
    hidden: true,
  },
  {
    path: '/404',
    component: () => import('@/views/404'),
    hidden: true
  },
  {
    path: '/',
    component: Layout,
    redirect: '/dashboard',
    children: [{
      path: 'dashboard',
      name: 'Dashboard',
      component: () => import('@/views/dashboard/index'),
      meta: { title: 'Dashboard', icon: 'dashboard', roles: ['admin', 'algo_user']}
    }]
  },

  {
    path: '/model_time_info',
    component: Layout,
    name: '模型负责人',
    children: [
    {
      path: 'model_time_series_info',
      name: '模型负责人',
      component: () => import('@/views/model_time_info/model_time_series_info'),
      meta:{
        title: '模型负责人', icon:'user'}
    }]
  },

  {
    path: '/model_history',
    component: Layout,
    name: '模型历史',
    children: [
    {
      path: 'model_history',
      name: '模型历史',
      component: () => import('@/views/models_info/model_history'),
      meta:{
        title: '模型历史', icon:'history', roles:['admin', 'algo_user']}
    }]
  },

  {
    path: '/node_list',
    component: Layout,
    name: '机器列表',
    children: [
      {
        path: 'node_infos',
        name: '机器列表',
        component: () => import('@/views/node_tab/index'),
        meta: { title: '机器列表', icon: 'tree', roles: ['admin']},
      },
    ]
  },

  {
    path: '/predictor_service',
    component: Layout,
    children: [
      {
        path: 'index',
        name: 'predictor服务',
        component: () => import('@/views/service_list_tab/predictor_list'),
        meta: { title: '服务列表', icon: 'view-grid-list' , roles: ['admin']}
      }
    ]
  },

  {
    path: '/models',
    component: Layout,
    children: [
      {
        path: 'menu1',
        name: '模型列表',
        component: () => import('@/views/nested/model_list'),
        meta: { title: '模型列表', icon: 'page-template', roles: ['admin'] },
      },
    ]
  },

  {
    path: '/mysql',
    component: Layout,
    children: [
      {
        path : 'mysql_controller',
        name : 'mysql视图',
        component: () => import('@/views/mysql_controller/mysql_control'),
        meta: { title: 'mysql视图', icon: 'eye-open', roles: ['admin'] }
      }
    ]
  },
  {
    path: '/control',
    name: '控制面板',
    component: Layout,
    meta: { title: '控制面板', icon: 'example', roles: ['admin','algo_user'] },
    children: [
      {
        path : '/stress/list',
        name : '自动压测',
        component: () => import('@/views/stress/index'),
        meta: { title: '自动压测', icon: 'experiment', roles: ['admin'] }
      },{
        path : '/downgrade/list',
        name : '服务降级',
        component: () => import('@/views/downgrade/index'),
        meta: { title: '服务降级', icon: 'figma-flatten-selection', roles: ['admin'] }
      },{
        path : '/migrate/list',
        name : '集群调整',
        component: () => import('@/views/migrate/index'),
        meta: { title: '集群调整', icon: 'graphic_stitching_three', roles: ['admin'] }
      }
    ]
    
  },
  

  // 404 page must be placed at the end !!!
  { path: '*', redirect: '/404', hidden: true }
]

const createRouter = () => new Router({
  // mode: 'history', // require service support
  scrollBehavior: () => ({ y: 0 }),
  routes: constantRoutes
})

const router = createRouter()

export function resetRouter() {
  const newRouter = createRouter()
  router.matcher = newRouter.matcher // reset router
}

export default router
