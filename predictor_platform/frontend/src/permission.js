import router from './router'
import store from './store'
import { Message } from 'element-ui'
import NProgress from 'nprogress' // progress bar
import 'nprogress/nprogress.css' // progress bar style
import { getToken } from '@/utils/auth' // get token from cookie
import getPageTitle from '@/utils/get-page-title'
import { GenerateRoutes } from '@/store/modules/permission'

NProgress.configure({ showSpinner: false }) // NProgress Configuration

const whiteList = ['/login'] // no redirect whitelist

router.beforeEach(async(to, from, next) => {
  // start progress bar
  NProgress.start()

  // set page title
  document.title = getPageTitle(to.meta.title)

  // determine whether the user has logged in
  const hasToken = getToken()

  if (hasToken) {
    if (to.path === '/login') {
      next({ path: '/' })
    } else {
      if ( typeof(store.getters.roles) != "undefined") {
        next()
      } else {
         try {
           var user_ = hasToken.split(' ')
           if (user_.length < 1) {
             console.error("the user info invalid!");
             return
           }
           var role_list = [];
           role_list.push(user_[0]);
           store.dispatch('GenerateRoutes', {role_list}).then( router_res => {
              router.addRoutes(store.getters.addRoutes);
           })
           next()
        } catch (error) {
           // remove token and go to login page to re-login
           await store.dispatch('user/resetToken')
           Message.error(error || 'Has Error')
           next(`/login?redirect=${to.path}`)
           NProgress.done()
        }
      }
    }
  } else {
      if (whiteList.indexOf(to.path) !== -1) {
           // in the free login whitelist, go directly
      	   next()
      } else {
           // other pages that do not have permission to access are redirected to the login page.
      	   next(`/login?redirect=${to.path}`)
           NProgress.done()
      }
  }
})

router.afterEach(() => {
  // finish progress bar
  NProgress.done()
})

