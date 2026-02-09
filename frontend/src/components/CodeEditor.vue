<script setup lang="ts">
import { VueMonacoEditor } from '@guolao/vue-monaco-editor'

// 定义组件接收的参数
interface Props {
  modelValue: string      // 双向绑定的内容
  language?: string       // 语言：json, shell, html 等
  readOnly?: boolean      // 是否只读
  height?: string         // 高度
}

// 设置默认值
withDefaults(defineProps<Props>(), {
  language: 'json',
  readOnly: false,
  height: '100%'
})

// 定义发出的事件（用于 v-model 更新）
const emit = defineEmits(['update:modelValue', 'mount'])

// 处理内容变化
const handleChange = (val: string | undefined) => {
  emit('update:modelValue', val || '')
}

// 编辑器加载完成后的配置（可选）
const handleMount = (editor: any, monaco: any) => {
  // 这里可以做一些高级配置，比如设置自动换行
  editor.updateOptions({
    wordWrap: 'on',
    minimap: { enabled: false } // 关闭右侧缩略图，节省空间
  })
  emit('mount', editor, monaco)
}
</script>

<template>
  <div class="border border-gray-700 rounded overflow-hidden" :style="{ height: height }">
    <vue-monaco-editor
        :value="modelValue"
        :language="language"
        theme="vs-dark"
        :options="{
        automaticLayout: true,
        readOnly: readOnly,
        fontSize: 13,
        scrollBeyondLastLine: false,
        fontFamily: 'JetBrains Mono, Consolas, monospace'
      }"
        @update:value="handleChange"
        @mount="handleMount"
    />
  </div>
</template>