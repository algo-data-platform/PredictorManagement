import { getToken, setToken, removeToken } from '@/utils/auth'
import { resetRouter } from '@/router'
import { login, logout, getInfo, self} from '@/api/user'
import axios from 'axios'

const state = {
  token: getToken(),
  name: '',
  avatar: require( '@/assets/p/user_avatar.png')
}

const mutations = {
  SET_TOKEN: (state, token) => {
    state.token = token
  },
  SET_NAME: (state, name) => {
    state.name = name
  },
  SET_AVATAR: (state, avatar) => {
    state.avatar = avatar
  }
}

const actions = {
  // user login
  login({ commit }, userInfo)  {
    if(userInfo.username == '' || userInfo.password == '') {
	    return;
    }	    
    var login_url = "/user/login?username=" + userInfo.username + "&password=" + userInfo.password;
    return new Promise((resolve, reject) =>{
      axios.get(login_url).then(function (result) {
        if (result.data.code != 0) {
          console.log("login fail:",result.data.msg);
          return false
        } else {
          commit('SET_NAME', result.data.data.username);
          console.log("result data is:",result.data.data);
          resolve()
        }
      }).catch(error => {
        reject(error)
      })
    })
  },
  // get user info
  getInfo({ commit, state }) {
    return new Promise((resolve, reject) => {
      getInfo(state.token).then(response => {
        const { data } = response

        if (!data) {
          reject('Verification failed, please Login again.')
        }
        const { name, avatar } = data
        commit('SET_NAME', name)
        commit('SET_AVATAR', avatar)
        resolve(data)
        return data
      }).catch(error => {
        reject(error)
      })
    })
  },

  // user logout
  logout({ commit, state }) {
    var logout_url = "/user/logout?userToken=admin";
    axios.get(logout_url).then(function (result) {
      console.log("the logout url is:", result);
      return result;
    }).catch(error => {
      console.log("erro:", error);
    })
  },

  // remove token
  resetToken({ commit }) {
    return new Promise(resolve => {
      commit('SET_TOKEN', '')
      removeToken()
      resolve()
    })
  }
}

export default {
  namespaced: true,
  state,
  mutations,
  actions
}

