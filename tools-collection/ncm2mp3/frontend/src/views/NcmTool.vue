<script setup>
import { SelectDirectory, Transform } from '../../wailsjs/go/main/App';
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
const tableData = ref([])
const selected = ref([])


function selectDirectory() {
    SelectDirectory().then(result => {
        tableData.value = result
    })
}

function batchTransform() {
    if (selected.value.length == 0) {
        ElMessage.error('请选择需要转换的音乐')
        return
    }

    var paths = new Array();

    selected.value.forEach(function (item) {
        paths.push(item)
    })


    console.log(paths)
    Transform(paths).then(result => {
        console.log("done")
    })

}


function handleSelectionChange(selection) {
    selected.value = selection
}

function transform(val) {
    console.log(val)
    var param = new Array();
    param.push(val)
    Transform(param).then(result => {
        console.log("done")
    })

}


</script>


<template>
    <el-container>
        <el-header>
            <div>
                <el-row>
                    <div class="header-btn">
                        <el-button @click="selectDirectory" text type="primary">选择文件夹</el-button>
                        <el-button @click="batchTransform" text type="primary">转换</el-button>
                    </div>
                </el-row>
            </div>
        </el-header>
        <el-main><el-table ref="multipleTableRef" :data="tableData" style="width: 100%" empty-text="请选择文件"
                @selection-change="handleSelectionChange">
                <el-table-column type="selection" width="55" />
                <el-table-column property="name" label="文件名称" />
                <el-table-column property="size" label="文件大小" />
                <el-table-column label="修改时间">
                    <template #default="scope">{{ scope.row.mod_time }}</template>
                </el-table-column>
                <el-table-column fixed="right" label="操作" width="120">
                    <template #default="scope">
                        <el-button link type="primary" size="small" @click="transform(scope.row)">转换</el-button>
                        <!-- <el-button link type="primary" size="small">Edit</el-button> -->
                    </template>
                </el-table-column>
            </el-table></el-main>
        <el-footer> </el-footer>

    </el-container>
</template>


<style>
.header-btn {

    margin-top: 40px;
}
</style>