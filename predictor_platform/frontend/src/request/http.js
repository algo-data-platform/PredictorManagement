import axios from 'axios'
import QS from 'qs'
import { Toast } from 'vant'

// 环境的切换
if (process.env.NODE_ENV == 'development') {
  axios.defaults.baseURL = '';}
else if (process.env.NODE_ENV == 'debug') {
  axios.defaults.baseURL = '';
}
else if (process.env.NODE_ENV == 'production') {
  axios.defaults.baseURL = '';
}

axios.defaults.timeout = 10000;

/**
 * get方法，对应get请求
 * @param {String} url [请求的url地址]
 * @param {Object} params [请求时携带的参数]
 */
export function get(url, params){
  return new Promise((resolve, reject) =>{
    axios.get(url, {
      params: params
    }).then(res => {
      resolve(res.data);
}).catch(err =>{
    reject(err.data)
})
});}

/**
 * post方法，对应post请求
 * @param {String} url [请求的url地址]
 * @param {Object} params [请求时携带的参数]
 */
export function post(url, params) {
  return new Promise((resolve, reject) => {
    axios.post(url, QS.stringify(params))
      .then(res => {
      resolve(res.data);
})
.catch(err =>{
    reject(err.data)
})
});
}






