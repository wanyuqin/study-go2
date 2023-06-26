<script setup>
import NcmList from "./components/NcmList.vue"
import DownloadTool from './components/DownloadTool.vue';
import DownloadSettings from "./components/DownloadSettings.vue";
import { ref } from 'vue'

const showDownloadSettings = ref(true)


</script>

<template>
  <DownloadTool></DownloadTool>
  <!-- <DownloadSettings v-if="showDownloadSettings"></DownloadSettings> -->
  <!-- <NcmList></NcmList> -->
</template>

<style></style>
