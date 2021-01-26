import {get, post} from './http'

export const apiAddress = p => get('http://127.0.0.1:10025', p);
