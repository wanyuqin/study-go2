
import { createRouter, createWebHashHistory } from 'vue-router'

// import DownloadTools from "/components/DownloadTools.vue"
import DownloadTools from "../views/DownloadTool.vue"


const router = createRouter({
    history: createWebHashHistory(),
    routes: [{
        path: '/downloadTools', component: DownloadTools
    }]
})


export default router