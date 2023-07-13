<script setup>
import {
    Plus,
} from '@element-plus/icons-vue'

import { ref, reactive, onMounted } from 'vue'
import { ExtractLink, Download } from '../../wailsjs/go/main/App';
import { ElLoading, ElNotification } from 'element-plus'


const linkInputFormVisible = ref(false)
const linkInputForm = ref({})
const videoList = ref([])
let videoMap = {}
const linkLoading = ref(false)

function openLink() {
    linkInputFormVisible.value = true
}


function updateVideoMap() {
    videoMap = videoList.value.reduce((pre, cur) => {
        const key = cur.id
        pre[key] = cur
        return pre
    }, {})

}

function removeLink(index){
    videoList.value.splice(index,1)
}

// 确认
function confirmLink() {
    const loading = ElLoading.service({
        lock: true,
        text: '加载中',
        background: 'rgba(0, 0, 0, 0.7)',
    })
    setTimeout(() => {
        loading.close()
    }, 30000)
    linkInputFormVisible.value = false
    ExtractLink(linkInputForm.value.link).then(result => {
        result.forEach(function (item) {
            item.cancel = false
            item.download = false
            item.done = false
            item.flow = true
            item.delete=true
            // item.status = "exception"
            console.log(item)
            videoList.value.push(item)
        })
        console.log(videoList.value)
        updateVideoMap()
        loading.close()
    }).catch((err) => {
        ElNotification({
            title: '下载消息',
            message: err,
            type: 'error',
        })
    })
    linkInputForm.value.link = ""
}
// 下载
function download(param) {

    Download(param).then(result => {
        param.cancel = true
        param.download = true
        param.delete=false
    }).catch((err) => {
        ElNotification({
            title: '下载消息',
            message: err,
            type: 'error',
        })
    })
}
// 取消下载
function cancelDownload(param) {
    param.download = false
    param.cancel = false
    param.delete=true
    videoList.value.forEach(function (item) {

    })

}

onMounted(() => {
    window.runtime.EventsOn("download.percent.refresh", function (msg) {
        console.log(msg)
        videoMap[msg.Eld.id].percentage = msg.Eld.percentage
        if (msg.Eld.percentage == 100) {
            videoMap[msg.Eld.id].cancel = false
            videoMap[msg.Eld.id].done = true
        }
    })

    window.runtime.EventsOn("download.done", function (msg) {
        console.log(msg)
        videoMap[msg.id].percentage = msg.percentage
        if (msg.percentage == 100) {
            videoMap[msg.id].cancel = false
            videoMap[msg.id].done = true
        }

        ElNotification({
            title: '下载通知',
            message: msg.title + "下载完成",
        })
    })

})


</script>


<template>
    <div class="common-layout">
        <el-container>
            <el-header>
                <el-row>
                    <div class="header-btn">
                        <el-button @click="openLink"  text type="primary">添加链接</el-button>
                        <el-button @click="openLink" text type="primary">下载设置</el-button>
                        <el-button @click="" text type="primary">下载历史</el-button>


                    </div>
                </el-row>
            </el-header>
            <el-main>
                <el-table :data="videoList" style="width: 100%" empty-text="请输入链接">
                    <el-table-column prop="title" label="标题" show-overflow-tooltip />
                    <el-table-column prop="url" label="地址" show-overflow-tooltip />
                    <el-table-column prop="size" label="大小" />
                    <el-table-column prop="quality" label="品质" show-overflow-tooltip />
                    <el-table-column prop="type" label="类型" width="70" />
                    <el-table-column prop="" label="下载进度">
                        <template #default="scope">
                            <el-progress text-inside :percentage="scope.row.percentage" :stroke-width="15"
                                :striped-flow="scope.row.flow" :status="scope.row.status" />
                        </template>
                    </el-table-column>
                    <el-table-column fixed="right" label="操作" width="120">
                        <template #default="scope">
                            <el-button v-if="!scope.row.download" link type="primary" size="small"
                                @click.prevent="download(scope.row)">
                                下载
                            </el-button>
                            <!-- TODO 下载出来才显示取消 -->
                            <el-button v-if="scope.row.cancel" link type="primary" size="small"
                                @click.prevent="cancelDownload(scope.row)">
                                取消
                            </el-button>
                            <el-button v-if="scope.row.done" link type="primary" size="small">
                                完成
                            </el-button>

                            <el-button type="danger" v-if="scope.row.delete"  @click.prevent="removeLink(scope.$index)" link   size="small">
                                删除
                            </el-button>
                        </template>
                    </el-table-column>
                </el-table>

            </el-main>
            <el-footer>
                <!-- <el-row>
                    <el-popover placement="bottom" title="" :width="200" trigger="hover" content="输入需要下载的地址">
                        <template #reference>
                            <el-button @click="openLink" type="success" :icon="Plus" circle />
                        </template>
                    </el-popover>
                </el-row> -->
            </el-footer>
        </el-container>
    </div>


    <!-- 链接确认框 -->
    <el-dialog v-model="linkInputFormVisible" title="" v>

        <el-form :model="linkInputForm">
            <el-form-item label="视频地址">
                <el-input text="https://www.bilibili.com/video/BV1dM4y1E7Yu/?spm_id_from=333.1007.tianma.1-2-2.click"
                    v-model="linkInputForm.link" autocomplete="off" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="linkInputFormVisible = false">取消</el-button>
                <el-button type="primary" @click="confirmLink">
                    确认
                </el-button>
            </span>
        </template>
    </el-dialog>





    <!-- 下载列表 -->
</template>