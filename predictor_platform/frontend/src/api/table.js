import request from '@/utils/request'
import axios from 'axios'

export function logout() {
  return request({
    url: '/user/logout',
    method: 'post'
  })
}
