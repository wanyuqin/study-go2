import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import { createRouter, createWebHashHistory } from 'vue-router'

import 'element-plus/dist/index.css'

import App from './App.vue'
import DownloadTools from "./views/DownloadTool.vue"
import NcmTool from "./views/ncmTool.vue"

const router = createRouter({
    history: createWebHashHistory(),
    routes: [{
        path: '/downloadTools', component: DownloadTools
    },{
        path: '/ncmTools', component: NcmTool
    }]
})


const app = createApp(App)

app.use(ElementPlus)
app.use(router)
app.mount('#app')
