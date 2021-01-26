import { constantRoutes } from '@/router'

//通过meta.role判断是否与当前用户权限匹配
function hasFilterRoute(roles, route) {
  if (route.meta && route.meta.roles) {
    return roles.some(role => route.meta.roles.indexOf(role) >= 0)
  }
  return true
}

//递归过滤异步路由表，返回符合用户角色权限的路由表
function filterAsyncRouter(constantRoutes, roles) {
  const accessedRouters = constantRoutes.filter(route => {
    if (hasFilterRoute(roles, route)) {
      if (route.children && route.children.length) {
        route.children = filterAsyncRouter(route.children, roles)
      }
      return true
    }
    return false
  })
  return accessedRouters
}

const permission = {
    state: {
      routers: constantRoutes,
      addRouters: []
    },
    mutations: {
      SET_ROUTERS: (state, routers) => {
      state.addRouters = routers
      state.routers = constantRoutes.concat(routers)
    }
  },
  actions: {
    GenerateRoutes({ commit }, data) {
      return new Promise(resolve => {
        const roles  = data.role_list
        let accessedRouters
        if (roles.indexOf('admin') >= 0) {
          accessedRouters = constantRoutes
        } else {
          accessedRouters = filterAsyncRouter(constantRoutes, roles)
        }
        commit('SET_ROUTERS', accessedRouters)
        resolve(data)
        return accessedRouters
      })
    }
  }
}

//export default permission
export default permission
