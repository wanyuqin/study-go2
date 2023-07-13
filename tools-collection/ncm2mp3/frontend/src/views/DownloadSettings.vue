<script setup>
import { ref, onMounted } from 'vue'

import { GetDownloadSettings,SaveDownloadSettings} from '../../wailsjs/go/main/App';
const downloadSettingForm = ref({})

onMounted(() => {
    getDownloadSettings()
   
})


function getDownloadSettings(){
    GetDownloadSettings().then(result => {
        downloadSettingForm.value.path = result.path
    })
}


function onSubmit(){
    console.log(downloadSettingForm.value)
    SaveDownloadSettings(downloadSettingForm.value).then(result=>{
        getDownloadSettings()
    })
}

</script>


<template>
    <el-container>
        <el-header></el-header>
        <el-main> <el-form :model="downloadSettingForm" label-width="120px">
                <el-form-item label="文件保存路径">
                    <el-input v-model="downloadSettingForm.path" />
                </el-form-item>
                <el-form-item>
                    <el-button type="primary" @click="onSubmit">确认</el-button>
                    <!-- <el-button>取消</el-button> -->
                </el-form-item>
            </el-form></el-main>
    </el-container>
</template>