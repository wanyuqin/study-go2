<script setup>

import { ref,onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElLoading, ElNotification } from 'element-plus'
import DownloadSettings from "./views/DownloadSettings.vue"
const showDownloadSettings = ref(true)

const router = useRouter()
const route = useRoute()


function changeRoute(route) {
  router.push(route)

}

onMounted(() => {
    window.runtime.EventsOn("download.done", function (msg) {
        console.log(msg)
        ElNotification({
            title: '下载通知',
            message: msg.title + "下载完成",
            duration:1000

        })
    })
    window.runtime.EventsOn("ncm.transform.done", function (msg) {
        console.log(msg)
        ElNotification({
            title: '转换通知',
            message: msg + "转换完成",
            duration:1000
        })
    })

})

</script>

<template>
  <div class="common-layout">
    <el-container>
      <el-header>
        <div class="topBar" >
          <el-row class="mb-4">
            <el-button type="info" text @click="changeRoute('/ncmTools')">Ncm转换工具</el-button>
            <el-button type="info" text @click="changeRoute('/downloadTools')">视频下载器</el-button>
            <el-button type="info" text @click="changeRoute('/downloadSettings')">下载设置</el-button>
          </el-row>
        </div>
      </el-header>
      <el-main>
        <router-view>
        </router-view>
      </el-main>
    </el-container>
  </div>

  <!-- <DownloadSettings></DownloadSettings> -->

</template>

<style>

.topBar {
 margin-top: 10px;
}
</style>
