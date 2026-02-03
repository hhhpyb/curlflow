<script setup lang="ts">
import {darkTheme} from 'naive-ui'
// 引入刚才创建的业务组件
import MainLayout from './components/MainLayout.vue'
</script>

<template>
  <n-config-provider :theme="darkTheme">
    <n-global-style/>
    <n-message-provider>
      <n-dialog-provider>
        <MainLayout/>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>